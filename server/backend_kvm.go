package server

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/quadrifoglio/go-qemu"
	"github.com/quadrifoglio/go-qmp"

	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/system"
	"github.com/quadrifoglio/wir/utils"
)

const (
	GiB             = 1073741824
	DefaultDiskSize = 25 * GiB
)

// MachineKvmIsRunning checks if the speicifed machine
// is currently running
func MachineKvmIsRunning(id string) bool {
	opts, err := DBMachineGetKvmOpts(id)
	if err != nil {
		log.Printf("Not Fatal: KVM machine '%s' is running check: %s\n", id, err)
		return false
	}

	if opts.PID == 0 {
		return false
	}

	proc, err := os.FindProcess(opts.PID)
	if err != nil {
		opts.PID = 0
		DBMachineSetKvmOpts(id, opts)

		log.Printf("Not Fatal: KVM machine '%s' is running check: %s\n", id, err)
		return false
	}

	err = proc.Signal(syscall.Signal(0))
	if err != nil {
		opts.PID = 0
		DBMachineSetKvmOpts(id, opts)

		log.Printf("Not Fatal: KVM machine '%s' is running check: %s\n", id, err)
		return false
	}

	return true
}

// MachineKvmCreate will acutally create the virtual machine
// based on the specified machine & image definitions
func MachineKvmCreate(def *shared.MachineDef) error {
	err := os.MkdirAll(filepath.Dir(MachineDisk(def.ID)), 0755)
	if err != nil {
		return err
	}

	if !utils.FileExists(MachineDisk(def.ID)) { // If this is a creation, not a migration
		disk := qemu.NewImage(MachineDisk(def.ID), qemu.ImageFormatQCOW2, 0)

		// If we are using an image, then create the disk file
		// with the same size as the image, and use a backing file
		if len(def.Image) > 0 {
			img, err := DBImageGet(def.Image)
			if err != nil {
				return err
			}

			size, err := system.SizeQcow2(ImageFile(def.Image))
			if err != nil {
				return err
			}

			if def.Disk == 0 {
				def.Disk = size // Update the disk size in definition (to be saved in database)
			}

			disk.Size = size
			disk.SetBackingFile(img.Source)
		} else {
			disk.Size = def.Disk
		}

		err = disk.Create()
		if err != nil {
			return err
		}

		// If both an image and a disk size is specified, then we have
		// to resize the disk image and its paritions to fit the requested size
		if len(def.Image) > 0 && def.Disk != 0 && def.Disk > disk.Size {
			err := system.ResizeQcow2(MachineDisk(def.ID), def.Disk)
			if err != nil {
				return err
			}
		}
	} else if len(def.Image) > 0 { // If this is a migration, rebase the disk to the image
		img, err := DBImageGet(def.Image)
		if err != nil {
			return err
		}

		disk, err := qemu.OpenImage(MachineDisk(def.ID))
		if err != nil {
			return err
		}

		err = disk.Rebase(ImageFile(img.ID))
		if err != nil {
			return err
		}
	}

	return nil
}

// MachineKvmStart starts the mahine based on the machine ID
// and returns the PID of the hypervisor's process
func MachineKvmStart(id string) error {
	if MachineKvmIsRunning(id) {
		return fmt.Errorf("Machine already running")
	}

	def, err := DBMachineGet(id)
	if err != nil {
		return err
	}

	opts, err := DBMachineGetKvmOpts(def.ID)
	if err != nil {
		return err
	}

	disk, err := qemu.OpenImage(MachineDisk(def.ID))
	if err != nil {
		return err
	}

	m := qemu.NewMachine(def.Cores, def.Memory)
	m.AddDriveImage(disk)

	for _, v := range def.Volumes {
		m.AddDrive(qemu.Drive{VolumeFile(v), qemu.ImageFormatQCOW2})
	}

	for i, iface := range def.Interfaces {
		netdev, err := qemu.NewNetworkDevice("tap", fmt.Sprintf("net%d", i))
		if err != nil {
			return err
		}

		netdev.SetMacAddress(iface.MAC)
		netdev.SetHostInterfaceName(MachineNicName(id, i))

		m.AddNetworkDevice(netdev)
	}

	if opts.VNC.Enabled {
		m.AddVNC(opts.VNC.Address, opts.VNC.Port)
	}

	m.AddMonitorUnix(MachineMonitorPath(def.ID))

	proc, err := m.Start("x86_64", true) // x86_64 arch (using qemu-system-x86_64), with kvm
	if err != nil {
		return err
	}

	// Wait 1 second, just to make sure that the interfaces have been created by QEMU
	time.Sleep(1 * time.Second)
	for i, iface := range def.Interfaces {
		err := AttachInterfaceToNetwork(id, i, iface)
		if err != nil {
			proc.Kill()

			return err
		}
	}

	opts.PID = proc.Pid

	err = DBMachineSetKvmOpts(def.ID, opts)
	if err != nil {
		return err
	}

	return nil
}

// MachineKvmStop stops the machine by finding its
// PID an sending it a SIGTERM signal
func MachineKvmStop(id string) error {
	opts, err := DBMachineGetKvmOpts(id)
	if err != nil {
		return err
	}

	if opts.PID == 0 {
		return fmt.Errorf("Machine already stopped")
	}

	proc, err := os.FindProcess(opts.PID)
	if err != nil {
		return err
	}

	opts.PID = 0

	err = DBMachineSetKvmOpts(id, opts)
	if err != nil {
		return err
	}

	err = proc.Kill()
	if err != nil {
		return err
	}

	return nil
}

// MachineKvmStatus returns a MachineStatusDef
// representing the current status of the machine
func MachineKvmStatus(id string) (shared.MachineStatusDef, error) {
	var def shared.MachineStatusDef
	def.Running = MachineKvmIsRunning(id)

	if def.Running {
		opts, err := DBMachineGetKvmOpts(id)
		if err != nil {
			return def, err
		}

		ram, err := system.ProcessRamUsage(opts.PID)
		if err != nil {
			return def, err
		}

		cpu, err := system.ProcessCpuUsage(opts.PID)
		if err != nil {
			return def, err
		}

		def.CpuUsage = cpu
		def.RamUsage = ram
	}

	disk, err := utils.FileSize(MachineDisk(id))
	if err != nil {
		return def, err
	}

	def.DiskUsage = disk

	return def, nil
}

// MachineKvmCreateCheckpoint creates a checkpoint of
// a running machine under the specified name
func MachineKvmCreateCheckpoint(id string, checkpoint string) error {
	if !MachineKvmIsRunning(id) {
		return fmt.Errorf("Machine is not running")
	}

	c, err := qmp.Open("unix", MachineMonitorPath(id))
	if err != nil {
		return err
	}

	defer c.Close()

	_, err = c.HumanMonitorCommand(fmt.Sprintf("savevm %s", checkpoint))
	if err != nil {
		return err
	}

	return nil
}

// MachineKvmRestoreCheckpoint restores the machine to
// the specified checkpoint
func MachineKvmRestoreCheckpoint(id, checkpoint string) error {
	if !MachineKvmIsRunning(id) {
		return fmt.Errorf("Machine must be running to be restored")
	}

	c, err := qmp.Open("unix", MachineMonitorPath(id))
	if err != nil {
		return err
	}

	defer c.Close()

	_, err = c.HumanMonitorCommand(fmt.Sprintf("loadvm %s", checkpoint))
	if err != nil {
		return err
	}

	return nil
}

// MachineKvmDeleteCheckpoint delete the checkpoint of
// the machine corresponding to the specified name
func MachineKvmDeleteCheckpoint(id, checkpoint string) error {
	img, err := qemu.OpenImage(MachineDisk(id))
	if err != nil {
		return err
	}

	err = img.DeleteSnapshot(checkpoint)
	if err != nil {
		return err
	}

	return nil
}
