package system

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/quadrifoglio/wir/utils"
)

var (
	nbdMutex sync.Mutex
)

// Partition represents a parition on a disk
type Partition struct {
	Number     int
	Start      uint64
	End        uint64
	Size       uint64
	Filesystem string
}

// ResizeQcow2 resizes the image to the specified size
// and extends the last partition in the image to
// fit the new size
func ResizeQcow2(path string, size uint64) error {
	dev := "/dev/nbd0"
	cmd := exec.Command("qemu-img", "resize", path, strconv.FormatUint(size, 10))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("qemu-img: %s", utils.OneLine(out))
	}

	err = NBDConnectQcow2(path)
	if err != nil {
		return err
	}

	defer NBDDisconnectQcow2()

	err = ResizeLastPartition(dev)
	if err != nil {
		return err
	}

	return nil
}

// NBDConnectQcow2 connects the specified QCOW2 image
// to the NBD device on the host
func NBDConnectQcow2(file string) error {
	dev := "/dev/nbd0"
	cmd := exec.Command("qemu-nbd", "-c", dev, file)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("connect nbd: %s", utils.OneLine(out))
	}

	nbdMutex.Lock()

	/*cmd = exec.Command("partx", "-a", dev)

	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("connect nbd: %s", utils.OneLine(out))
	}*/

	return nil
}

// NBDDisconnectQcow2 disconnects the image connected
// to the NBD device on the host
func NBDDisconnectQcow2() error {
	dev := "/dev/nbd0"
	cmd := exec.Command("qemu-nbd", "-d", dev)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("disconnect nbd: %s", utils.OneLine(out))
	}

	nbdMutex.Unlock()

	return nil
}

// ListPartitions lists all the partitions on
// the specified device
func ListPartitions(dev string) ([]Partition, error) {
	cmd := exec.Command("parted", "-m", dev, "unit", "B", "print", "free")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("list paritions in %s: %s", dev, utils.OneLine(out))
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

// ResizeLastPartition resizes the last partition on
// the device to fit the maximum size of the device
func ResizeLastPartition(dev string) error {
	parts, err := ListPartitions(dev)
	if err != nil {
		return fmt.Errorf("resize %s: %s", dev, err)
	}

	if len(parts) < 2 {
		return fmt.Errorf("resize %s: not enough partitions", dev)
	}

	num := len(parts) - 2

	freeSpace := parts[num+1]
	mainPart := parts[num]

	if freeSpace.Filesystem != "free" {
		return fmt.Errorf("resize %s part %d: no free space available", dev, num)
	}

	cmd := exec.Command("parted", dev, "unit", "B", "resizepart", strconv.Itoa(mainPart.Number), strconv.FormatUint(freeSpace.End, 10))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("resize %s part %d: %s", dev, num, utils.OneLine(out))
	}

	cmd = exec.Command("e2fsck", "-f", "-y", fmt.Sprintf("%sp%d", dev, num))
	cmd.Run()

	cmd = exec.Command("resize2fs", fmt.Sprintf("%sp%d", dev, num))

	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("resize %s part %d: %s", dev, num, utils.OneLine(out))
	}

	return nil
}
