package utils

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
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
		return fmt.Errorf("mount-tmp: mkdir: %s", err)
	}

	cmd := exec.Command("mount", dev, path)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("mount-tmp: mount: %s", err)
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
