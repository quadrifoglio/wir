package system

import (
	"github.com/quadrifoglio/go-qemu"
)

const (
	KiB = 1024
)

// CreateVolume creates a volume of the specified
// type and size, at the precised file
func CreateVolume(t, file string, size uint64) error {
	if t == "kvm" {
		img := qemu.NewImage(file, qemu.ImageFormatQCOW2, size*KiB)

		err := img.Create()
		if err != nil {
			return err
		}
	}

	if t == "qemu" {
		// TODO: Add LXC volumes
	}

	return nil
}
