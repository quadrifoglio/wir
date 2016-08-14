package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/amoghe/go-crypt"

	"github.com/quadrifoglio/wir/shared"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}

	return true
}

func TarDirectory(path, output string) error {
	cmd := exec.Command("tar", "cf", output, "--numeric-owner", "-C", path, ".")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tar directory: %s", OneLine(out))
	}

	return nil
}

func UntarDirectory(input, path string) error {
	cmd := exec.Command("tar", "xf", input, "-C", path)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("untar directory: %s", OneLine(out))
	}

	return nil
}

func MakeRemoteDirectories(dst shared.Remote, dstDir string) error {
	cmd := exec.Command("ssh", fmt.Sprintf("%s@%s", dst.SSHUser, dst.Addr), "mkdir -p "+dstDir)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("create remote dir: %s", OneLine(out))
	}

	return nil
}

func SCP(srcFile string, dst shared.Remote, dstFile string) error {
	dstf := fmt.Sprintf("%s:%s", dst.Addr, dstFile)
	cmd := exec.Command("scp", "-r", srcFile, dstf)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("scp to remote: %s", OneLine(out))
	}

	return nil
}

func CopyFolder(src, dst string) error {
	cmd := exec.Command("cp", "-R", src, dst)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("copy folder: %s", OneLine(out))
	}

	return nil
}

func CopyFile(src, dst string) error {
	f1, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("copy file: %s", err)
	}

	defer f1.Close()

	f2, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("copy file: %s", err)
	}

	defer f2.Close()

	_, err = io.Copy(f2, f1)
	if err != nil {
		return fmt.Errorf("copy file: %s", err)
	}

	return nil
}

func RewriteFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return fmt.Errorf("rewrite %s: %s", path, err)
	}

	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("rewrite %s: %s", path, err)
	}

	return nil
}

func ReplaceInFile(path, search, replace string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("replace in %s: %s", path, err)
	}

	if strings.Contains(string(data), search) {
		newData := strings.Replace(string(data), search, replace, -1)

		err = RewriteFile(path, []byte(newData))
		if err != nil {
			return fmt.Errorf("replace in %s: %s", path, err)
		}
	}

	return nil
}

func DeleteLinesInFile(path, prefix string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("delete lines in %s: %s", path, err)
	}

	var newData []byte
	for _, l := range strings.Split(string(data), "\n") {
		if len(l) > 1 && !strings.HasPrefix(l, prefix) {
			newData = append(newData, []byte(fmt.Sprintf("%s\n", l))...)
		}
	}

	err = RewriteFile(path, newData)
	if err != nil {
		return fmt.Errorf("delete lines in %s: %s", path, err)
	}

	return nil
}

func ChangeHostname(hostnamePath, hostname string) error {
	err := RewriteFile(hostnamePath, []byte(hostname))
	if err != nil {
		return fmt.Errorf("change hostname: %s", err)
	}

	return nil
}

func ChangeRootPassword(shadowPath, root string) error {
	data, err := ioutil.ReadFile(shadowPath)
	if err != nil {
		return fmt.Errorf("change root passwd: %s", err)
	}

	n := strings.Index(string(data), ":")
	if n == -1 {
		return fmt.Errorf("change root passwd: no ':' char in /etc/shadow")
	}

	nn := strings.Index(string(data[n+1:]), ":")
	if n == -1 {
		return fmt.Errorf("change root passwd: no second ':' char in /etc/shadow")
	}

	n = n + nn + 1
	salt := UniqueID(0)

	str, err := crypt.Crypt(root, fmt.Sprintf("$6$%s$", string(salt[:8])))
	if err != nil {
		return fmt.Errorf("change root passwd: crypt: %s", err)
	}

	str = "root:" + str

	newData := make([]byte, len(str))
	copy(newData, str)
	newData = append(newData, data[n:]...)

	err = RewriteFile(shadowPath, newData)
	if err != nil {
		return fmt.Errorf("change root passwd: %s", err)
	}

	return nil
}
