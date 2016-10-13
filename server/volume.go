package server

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/quadrifoglio/go-qemu"

	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/utils"
)

const (
	KiB = 1024
)

// CreateVolume creates a volume based on
// the specified definition
func CreateVolume(def shared.VolumeDef) error {
	file := VolumeFile(def.ID)

	if !utils.FileExists(filepath.Dir(file)) {
		err := os.MkdirAll(filepath.Dir(file), 0755)
		if err != nil {
			return err
		}
	}

	if def.Type == "kvm" {
		img := qemu.NewImage(file, qemu.ImageFormatQCOW2, def.Size*KiB)

		err := img.Create()
		if err != nil {
			return err
		}
	}
	if def.Type == "qemu" {
		// TODO: Add LXC volumes
	}

	return nil
}

// DeleteVolume deletes the specified volume's
// data file
func DeleteVolume(id string) error {
	file := VolumeFile(id)

	if !utils.FileExists(file) {
		return fmt.Errorf("Volume file not found")
	}

	err := os.Remove(filepath.Dir(file))
	if err != nil {
		return err
	}

	return nil
}
