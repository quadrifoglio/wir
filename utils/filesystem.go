package utils

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var (
	nbdMutex sync.Mutex
)

type Partition struct {
	Number     int
	Start      uint64
	End        uint64
	Size       uint64
	Filesystem string
}

func SizeQcow2(file string) (uint64, error) {
	var size uint64

	cmd := exec.Command("qemu-img", "info", file)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("get qcow2 size: %s", OneLine(out))
	}

	for _, l := range strings.Split(string(out), "\n") {
		if !strings.Contains(l, ":") {
			continue
		}

		r := regexp.MustCompile("\\(([0-9]+)\\)")
		if d := r.Find([]byte(l)); d != nil {
			s, err := strconv.ParseInt(string(d), 10, 64)
			if err != nil {
				return 0, fmt.Errorf("get qcow2 size: %s", err)
			}

			size = uint64(s)
			break
		}
	}

	return size, nil
}

func ResizeQcow2(file string, size uint64) error {
	cmd := exec.Command("qemu-img", "resize", file, strconv.FormatUint(size, 10))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("resize qcow2: %s", OneLine(out))
	}

	return nil
}

func ListSnapshotsQcow2(file string) ([]string, error) {
	cmd := exec.Command("qemu-img", "snapshot", "-l", file)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("list qcow2 snapshots: %s", OneLine(out))
	}

	var sns []string
	for i, l := range strings.Split(string(out), "\n") {
		if i < 2 {
			continue
		}

		f := strings.Fields(l)

		if len(f) > 1 {
			sns = append(sns, f[1])
		}
	}

	return sns, nil
}

func SnapshotQcow2(file, name string) error {
	cmd := exec.Command("qemu-img", "snapshot", "-c", name, file)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("create qcow2 snapshot: %s", OneLine(out))
	}

	return nil
}

func RestoreQcow2(file, name string) error {
	cmd := exec.Command("qemu-img", "snapshot", "-a", name, file)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("restore qcow2 snapshot: %s", OneLine(out))
	}

	return nil
}

func DeleteSnapshotQcow2(file, name string) error {
	cmd := exec.Command("qemu-img", "snapshot", "-d", name, file)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("delete qcow2 snapshot: %s", OneLine(out))
	}

	return nil
}

func NBDConnectQcow2(file string) error {
	dev := "/dev/nbd0"
	cmd := exec.Command("qemu-nbd", "-c", dev, file)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("connect nbd: %s", OneLine(out))
	}

	nbdMutex.Lock()

	cmd = exec.Command("partx", "-a", dev)

	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("connect nbd: %s", OneLine(out))
	}

	return nil
}

func NBDDisconnectQcow2() error {
	defer nbdMutex.Unlock()

	dev := "/dev/nbd0"
	cmd := exec.Command("qemu-nbd", "-d", dev)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("disconnect nbd: %s", OneLine(out))
	}

	return nil
}

func ListPartitions(dev string) ([]Partition, error) {
	cmd := exec.Command("parted", "-m", dev, "unit", "B", "print", "free")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("list paritions in %s: %s", dev, OneLine(out))
	}

	var parts []Partition

	for _, l := range strings.Split(string(out), "\n") {
		if len(l) == 0 {
			break
		}

		if l[1] != ':' {
			continue
		}

		t := strings.Split(l[:len(l)-1], ":")
		if len(t) < 5 {
			return nil, fmt.Errorf("list partitions in %s: invalid output from parted", dev)
		}

		var p Partition

		if v, err := strconv.Atoi(t[0]); err == nil {
			p.Number = v
		}
		if v, err := strconv.ParseInt(t[1][:len(t[1])-1], 10, 64); err == nil {
			p.Start = uint64(v)
		}
		if v, err := strconv.ParseInt(t[2][:len(t[2])-1], 10, 64); err == nil {
			p.End = uint64(v)
		}
		if v, err := strconv.ParseInt(t[3][:len(t[3])-1], 10, 64); err == nil {
			p.Size = uint64(v)
		}

		p.Filesystem = t[4]

		if p.Filesystem == "free" {
			p.Number = 0
		}

		parts = append(parts, p)
	}

	return parts, nil
}

func ResizePartition(dev string, num int, size uint64) error {
	parts, err := ListPartitions(dev)
	if err != nil {
		return fmt.Errorf("resize %s part %d: %s", dev, num, err)
	}

	var mainPart Partition
	var freeSpace Partition
	for i, p := range parts {
		if p.Number == num {
			mainPart = p

			if i+1 < len(parts) {
				pp := parts[i+1]
				if pp.Filesystem == "free" {
					freeSpace = pp
				}
			}

			break
		}
	}

	if mainPart.Number == 0 {
		return fmt.Errorf("resize %s part %d: main part not found", dev, num)
	}
	if freeSpace.End == 0 {
		return fmt.Errorf("resize %s part %d: no free space available after partition", dev, num)
	}

	cmd := exec.Command("parted", dev, "unit", "B", "resizepart", strconv.Itoa(mainPart.Number), strconv.FormatUint(freeSpace.End, 10))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("resize %s part %d: %s", dev, num, OneLine(out))
	}

	cmd = exec.Command("e2fsck", "-f", "-y", fmt.Sprintf("%sp%d", dev, num))
	cmd.Run()

	cmd = exec.Command("resize2fs", fmt.Sprintf("%sp%d", dev, num))

	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("resize %s part %d: %s", dev, num, OneLine(out))
	}

	fmt.Println()

	return nil
}

func Mount(dev, path string) error {
	err := os.MkdirAll(path, 0775)
	if err != nil {
		return fmt.Errorf("mount: %s", err)
	}

	cmd := exec.Command("mount", dev, path)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mount: %s", OneLine(out))
	}

	return nil
}

func Unmount(path string) error {
	err := syscall.Unmount(path, 0)
	if err != nil {
		return fmt.Errorf("unmount %s: %s", path, err)
	}

	return nil
}
