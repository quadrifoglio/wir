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

func ZfsListSnapshots(path string) ([]string, error) {
	cmd := exec.Command("zfs", "list", "-t", "snapshot")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("zfs: %s", out)
	}

	var snaps []string
	for i, l := range strings.Split(string(out), "\n") {
		if len(l) == 0 {
			break
		}
		if i == 0 {
			continue
		}

		f := strings.Fields(l)
		if len(f) < 1 {
			return nil, fmt.Errorf("zfs: invalid output from zfs list")
		}

		s := strings.Split(f[0], "@")
		if len(s) != 2 {
			return nil, fmt.Errorf("zfs: invalid output from zfs list")
		}

		if s[0] == path {
			snaps = append(snaps, s[1])
		}
	}

	return snaps, nil
}

func ZfsSnapshot(path, snap string) error {
	cmd := exec.Command("zfs", "snapshot", fmt.Sprintf("%s@%s", path, snap))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("zfs: %s", out)
	}

	return nil
}

func ZfsRestore(path, snap string) error {
	cmd := exec.Command("zfs", "rollback", fmt.Sprintf("%s@%s", path, snap))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("zfs: %s", out)
	}

	return nil
}

func ZfsDeleteSnapshot(path, snap string) error {
	cmd := exec.Command("zfs", "destroy", fmt.Sprintf("%s@%s", path, snap))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("zfs: %s", out)
	}

	return nil
}

func ZfsClone(path, new string) error {
	err := ZfsSnapshot(path, "clone")
	if err != nil {
		return err
	}

	cmd := exec.Command("zfs", "clone", fmt.Sprintf("%s@clone", path), new)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("zfs: %s", out)
	}

	err = ZfsDeleteSnapshot(path, "clone")
	if err != nil {
		return err
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

func ZfsDestroy(path string) error {
	cmd := exec.Command("zfs", "destroy", "-r", path)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("zfs: %s", out)
	}

	return nil
}
