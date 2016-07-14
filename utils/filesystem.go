package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/amoghe/go-crypt"
)

func NBDConnectQcow2(qemuNbd, dev, file string) error {
	cmd := exec.Command(qemuNbd, "-c", dev, file)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("nbd-connect: qemu-nbd: %s", err)
	}

	cmd = exec.Command("partx", "-a", "/dev/nbd0")

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("nbd-connect: partx: %s", err)
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
		return fmt.Errorf("mount: mount: %s", err)
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
