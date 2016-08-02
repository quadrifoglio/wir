package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

func ZfsCreate(path, mountpoint string) error {
	cmd := exec.Command("zfs", "create", "-o", fmt.Sprintf("mountpoint=%s", mountpoint), path)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("zfs: %s", out)
	}

	return nil
}

func ZfsSet(path, key, value string) error {
	cmd := exec.Command("zfs", "set", fmt.Sprintf("%s=%s", key, value), path)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("zfs: %s", out)
	}

	return nil
}

func ZfsIsMounted(path string) (bool, error) {
	cmd := exec.Command("zfs", "mount")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("zfs: %s", err)
	}

	for _, l := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(l, path) {
			return true, nil
		}
	}

	return false, nil
}

func ZfsMount(path string) error {
	cmd := exec.Command("zfs", "mount", path)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("zfs: %s", out)
	}

	return nil
}

func ZfsUnmount(path string) error {
	cmd := exec.Command("zfs", "umount", path)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("zfs: %s", out)
	}

	return nil
}
