package utils

import (
	"fmt"
	"os"
	"os/exec"
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

func ResizeQcow2(file string, size int) error {
	cmd := exec.Command("qemu-img", "resize", file, strconv.Itoa(size))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(out))
	}

	return nil
}

func NBDConnectQcow2(file string) error {
	dev := "/dev/nbd0"
	cmd := exec.Command("qemu-nbd", "-c", dev, file)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nbd-connect: qemu-nbd: %s", string(out))
	}

	nbdMutex.Lock()

	cmd = exec.Command("partx", "-a", dev)

	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nbd-connect: partx: %s", string(out))
	}

	return nil
}

func NBDDisconnectQcow2() error {
	defer nbdMutex.Unlock()

	dev := "/dev/nbd0"
	return exec.Command("qemu-nbd", "-d", dev).Run()
}

func ListPartitions(dev string) ([]Partition, error) {
	cmd := exec.Command("parted", "-m", dev, "unit", "B", "print", "free")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s", out)
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
			return nil, fmt.Errorf("invalid output from parted")
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

func ResizePartition(dev string, num, size int) error {
	parts, err := ListPartitions(dev)
	if err != nil {
		return err
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
		return fmt.Errorf("can not find main parition (was supposed to be number %d)", num)
	}
	if freeSpace.End == 0 {
		return fmt.Errorf("no free space available after partition %d. can not resize", num)
	}

	cmd := exec.Command("parted", dev, "unit", "B", "resizepart", strconv.Itoa(mainPart.Number), strconv.Itoa(int(freeSpace.End)))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", out)
	}

	cmd = exec.Command("e2fsck", "-f", "-y", fmt.Sprintf("%sp%d", dev, num))
	cmd.Run()

	cmd = exec.Command("resize2fs", fmt.Sprintf("%sp%d", dev, num))

	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", out)
	}

	fmt.Println()

	return nil
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
