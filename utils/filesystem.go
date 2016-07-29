package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/amoghe/go-crypt"

	"github.com/quadrifoglio/wir/shared"
)

func TarDirectory(path, output string) error {
	base := filepath.Base(path)
	dir := filepath.Dir(path)

	cmd := exec.Command("tar", "cf", output, "--numeric-owner", "-C", dir, base)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("failed to tar directory: %s", string(out))
		return err
	}

	return nil
}

func UntarDirectory(input, path string) error {
	cmd := exec.Command("tar", "xf", input, "-C", path)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("failed to untar directory: %s", string(out))
		return err
	}

	return nil
}

func MakeRemoteDirectories(dst shared.Remote, dstDir string) error {
	cmd := exec.Command("ssh", fmt.Sprintf("%s@%s", dst.SSHUser, dst.Addr), "mkdir -p "+dstDir)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("failed to create remote directory: %s", string(out))
		return err
	}

	return nil
}

func SCP(srcFile string, dst shared.Remote, dstFile string) error {
	dstf := fmt.Sprintf("%s:%s", dst.Addr, dstFile)
	cmd := exec.Command("scp", "-r", srcFile, dstf)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("scp failed: %s", string(out))
		return err
	}

	return nil
}

func ClearFolder(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}

	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}

	return nil
}

func NBDConnectQcow2(qemuNbd, dev, file string) error {
	cmd := exec.Command(qemuNbd, "-c", dev, file)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nbd-connect: qemu-nbd: %s", string(out))
	}

	cmd = exec.Command("partx", "-a", dev)

	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nbd-connect: partx: %s", string(out))
	}

	return nil
}

func NBDDisconnectQcow2(qemuNbd, dev string) error {
	return exec.Command(qemuNbd, "-d", dev).Run()
}

func Mount(dev, path string) error {
	err := os.MkdirAll(path, 0777)
	if err != nil {
		return fmt.Errorf("mount: mkdir: %s", err)
	}

	cmd := exec.Command("mount", dev, path)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("mount: %s", err)
	}

	return nil
}

func Unmount(path string) error {
	return syscall.Unmount(path, 0)
}

func RewriteFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return fmt.Errorf("rewrite-file: open %s: %s", path, err)
	}

	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("rewrite-file: write to %s: %s", path, err)
	}

	return nil
}

func ChangeHostname(hostnamePath, hostname string) error {
	return RewriteFile(hostnamePath, []byte(hostname))
}

func ChangeRootPassword(shadowPath, root string) error {
	data, err := ioutil.ReadFile(shadowPath)
	if err != nil {
		return fmt.Errorf("root-password: can not read entire file: %s", err)
	}

	n := strings.Index(string(data), ":")
	if n == -1 {
		return fmt.Errorf("root-password: invalid file (no ':' char)")
	}

	nn := strings.Index(string(data[n+1:]), ":")
	if n == -1 {
		return fmt.Errorf("root-password: invalid file (no second ':' char)")
	}

	n = n + nn + 1
	salt := UniqueID(0)

	str, err := crypt.Crypt(root, fmt.Sprintf("$6$%s$", string(salt[:8])))
	if err != nil {
		return fmt.Errorf("can not crypt password: %s", err)
	}

	str = "root:" + str

	newData := make([]byte, len(str))
	copy(newData, str)
	newData = append(newData, data[n:]...)

	err = RewriteFile(shadowPath, newData)
	if err != nil {
		return err
	}

	return nil
}
