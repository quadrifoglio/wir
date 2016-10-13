package server

import (
	"github.com/quadrifoglio/go-qemu"
	"github.com/quadrifoglio/wir/shared"
)

const (
	GiB      = 1073741824
	DiskSize = 25 * GiB
)

// MachineKvmCreate will acutally create the virtual machine
// based on the specified definition
func MachineKvmCreate(def shared.MachineDef) error {
	img, err := DBImageGet(def.Image)
	if err != nil {
		return err
	}

	disk := qemu.NewImage(MachineDisk(def.ID), qemu.ImageFormatQCOW2, DiskSize)
	disk.SetBackingFile(img.Source)

	err = disk.Create()
	if err != nil {
		return err
	}

	return nil
}

// MachineKvmStart starts the mahine based on the definition
// and returns the PID of the hypervisor's process
func MachineKvmStart(def shared.MachineDef) (int, error) {
	disk, err := qemu.OpenImage(MachineDisk(def.ID))
	if err != nil {
		return -1, err
	}

	m := qemu.NewMachine(def.Cores, int(def.Memory))
	m.AddDrive(disk)

	pid, err := m.Start("x86_64", true) // x86_64 arch (using qemu-system-x86_64), with kvm
	if err != nil {
		return -1, err
	}

	/*for _, v := range def.Volumes {
		m.AddDrive(qemu.Drive{qemu.ImageFormatQCOW2, volumePath})
	}*/

	return pid, nil
}
