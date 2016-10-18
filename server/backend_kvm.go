package server

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/quadrifoglio/go-qemu"
	"github.com/quadrifoglio/wir/shared"
)

const (
	GiB      = 1073741824
	DiskSize = 25 * GiB
)

// MachineKvmGetProcess returns the hypervisor's process
// ID corresponding to the specified machine ID, if any
func MachineKvmGetPID(id string) (int, error) {
	pidStr, err := DBMachineGetVal(id, "pid")
	if err != nil {
		return -1, err
	}

	if len(pidStr) == 0 {
		return 0, nil
	}

	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return -1, err
	}

	return pid, nil
}

// MachineKvmIsRunning checks if the speicifed machine
// is currently running
func MachineKvmIsRunning(id string) bool {
	pid, err := MachineKvmGetPID(id)
	if err != nil {
		log.Printf("Not Fatal: KVM machine '%s' is running check: %s\n", id, err)
		return false
	}

	if pid == 0 {
		return false
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		DBMachineSetVal(id, "pid", "")

		log.Printf("Not Fatal: KVM machine '%s' is running check: %s\n", id, err)
		return false
	}

	err = proc.Signal(syscall.Signal(0))
	if err != nil {
		DBMachineSetVal(id, "pid", "")

		log.Printf("Not Fatal: KVM machine '%s' is running check: %s\n", id, err)
		return false
	}

	return true
}

// MachineKvmCreate will acutally create the virtual machine
// based on the specified machine & image definitions
func MachineKvmCreate(def shared.MachineDef, img shared.ImageDef) error {
	err := os.MkdirAll(filepath.Dir(MachineDisk(def.ID)), 0755)
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
		netdev.SetHostInterfaceName(MachineIface(id, i))

		m.AddNetworkDevice(netdev)
	}

	if def.KvmVNC.Enable {
		m.AddVNC(def.KvmVNC.Addr, def.KvmVNC.Port)
	}

	pid, err := m.Start("x86_64", true) // x86_64 arch (using qemu-system-x86_64), with kvm
	if err != nil {
		return err
	}

	// Wait 1 second, just to make sure that the interfaces have been created by QEMU
	time.Sleep(1 * time.Second)
	for i, iface := range def.Interfaces {
		err := AttachInterfaceToNetwork(id, i, iface.Network)
		if err != nil {
			return err
		}
	}

	err = DBMachineSetVal(def.ID, "pid", strconv.Itoa(pid))
	if err != nil {
		return err
	}

	return nil
}

// MachineKvmStop stops the machine by finding its
// PID an sending it a SIGTERM signal
func MachineKvmStop(id string) error {
	pid, err := MachineKvmGetPID(id)
	if err != nil {
		return err
	}

	if pid == 0 {
		return fmt.Errorf("Machine already stopped")
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	err = DBMachineSetVal(id, "pid", "")
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

	return def, nil
}
