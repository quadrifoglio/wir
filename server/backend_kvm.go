package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/amoghe/go-crypt"
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
				// TODO: Find a better way
				//return err

				log.Println("Create machine: disk was not resized:", err)
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

// MachineKvmSetLinuxHostname sets the hostname for
// the specified Linux machine
func MachineKvmSetLinuxHostname(id, hostname string) error {
	err := system.NBDConnectQcow2(MachineDisk(id))
	if err != nil {
		return err
	}

	defer system.NBDDisconnectQcow2()

	path := fmt.Sprintf("/tmp/wir/%s", id)

	if !utils.FileExists(path) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}

	partitions, err := system.ListPartitions("/dev/nbd0")
	if err != nil {
		return err
	}

	if len(partitions) <= 1 {
		return fmt.Errorf("Not enough partitions (<= 1)")
	}

	mainPart := len(partitions) - 1
	if partitions[mainPart].Filesystem == "free" {
		mainPart--
	}

	err = system.Mount(fmt.Sprintf("/dev/nbd0p%d", mainPart), path)
	if err != nil {
		return err
	}

	defer system.Unmount(path)

	err = utils.ReplaceFileContents(fmt.Sprintf("%s/etc/hostname", path), []byte(hostname))
	if err != nil {
		return err
	}

	return nil
}

// MachineKvmSetLinuxRootPassword sets the root password for
// the specified Linux machine
func MachineKvmSetLinuxRootPassword(id string, passwd string) error {
	err := system.NBDConnectQcow2(MachineDisk(id))
	if err != nil {
		return err
	}

	defer system.NBDDisconnectQcow2()

	path := fmt.Sprintf("/tmp/wir/%s", id)
	shadowPath := fmt.Sprintf("%s/etc/shadow", path)

	if !utils.FileExists(path) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}

	partitions, err := system.ListPartitions("/dev/nbd0")
	if err != nil {
		return err
	}

	if len(partitions) <= 1 {
		return fmt.Errorf("Not enough partitions (<= 1)")
	}

	mainPart := len(partitions) - 1
	if partitions[mainPart].Filesystem == "free" {
		mainPart--
	}

	err = system.Mount(fmt.Sprintf("/dev/nbd0p%d", mainPart), path)
	if err != nil {
		return err
	}

	defer system.Unmount(path)

	data, err := ioutil.ReadFile(shadowPath)
	if err != nil {
		return err
	}

	salt := utils.RandID()
	str, err := crypt.Crypt(passwd, fmt.Sprintf("$6$%s$", salt[:8]))
	if err != nil {
		return err
	}

	regex := regexp.MustCompile("^root:[^:]+:")
	dataStr := regex.ReplaceAllLiteralString(string(data), fmt.Sprintf("root:%s:", str))

	err = utils.ReplaceFileContents(shadowPath, []byte(dataStr))
	if err != nil {
		return err
	}

	return nil
}

func MachineKvmSetOpts(id string, opts shared.KvmOptsDef) error {
	if MachineKvmIsRunning(id) {
		return fmt.Errorf("Cannot set KVM options while the machine is running")
	}

	if len(opts.Linux.Hostname) > 0 {
		err := MachineKvmSetLinuxHostname(id, opts.Linux.Hostname)
		if err != nil {
			return err
		}
	}
	if len(opts.Linux.RootPassword) > 0 {
		err := MachineKvmSetLinuxRootPassword(id, opts.Linux.RootPassword)
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

	if len(opts.CDRom) > 0 {
		m.AddCDRom(opts.CDRom)
	}

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
		m.AddVNC(opts.VNC.Address, opts.VNC.Port, opts.VNC.WebsocketPort, true)
	}

	m.AddMonitorUnix(MachineMonitorPath(def.ID))
	m.AddOption("-balloon", "virtio")
	m.AddOption("-usbdevice", "tablet")

	// x86_64 arch (using qemu-system-x86_64), with kvm
	proc, err := m.Start("x86_64", true, func(s string) {
		log.Printf("machine %s stderr: %s\n", def.ID, utils.OneLine([]byte(s)))
	})

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

	c, err := qmp.Open("unix", MachineMonitorPath(def.ID))
	if err == nil {
		defer c.Close()

		_, err := c.Command("qom-set", map[string]interface{}{
			"path":     "/machine/peripheral-anon/device[1]",
			"property": "guest-stats-polling-interval",
			"value":    3,
		})

		if err != nil {
			log.Printf("Not fatal - Machine %s - Set balloon stats polling interval: %s\n", def.ID, err)
		}
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

// MachineKvmGetBallonFreeMem retreives the guest's free memory
// in MiB from the virtio-balloon driver via QMP
func MachineKvmGetBallonFreeMem(id string) (uint64, error) {
	c, err := qmp.Open("unix", MachineMonitorPath(id))
	if err != nil {
		return 0, err
	}

	defer c.Close()

	res, err := c.Command("qom-get", map[string]interface{}{
		"path":     "/machine/peripheral-anon/device[1]",
		"property": "guest-stats",
	})

	if err != nil {
		return 0, err
	}

	if rr, ok := res.(map[string]interface{}); ok {
		if stats, ok := rr["stats"].(map[string]interface{}); ok {
			if freeRam, ok := stats["stat-free-memory"].(float64); ok {
				if freeRam < 0 {
					return 0, fmt.Errorf("stat-free-memory negative")
				}

				return uint64(freeRam / float64(1048576.0)), nil
			}
		}
	}

	return 0, fmt.Errorf("Invalid output from qom-get")
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

		ram, err := MachineKvmGetBallonFreeMem(id)
		if err != nil {
			log.Printf("Not fatal - Machine %s - Failed to get free memory from balloon: %s\n", id, err)

			ram, err = system.ProcessRamUsage(opts.PID)
			if err != nil {
				return def, err
			}
		}

		cpu, err := system.ProcessCpuUsage(opts.PID)
		if err != nil {
			return def, err
		}

		machine, err := DBMachineGet(id)
		if err != nil {
			return def, err
		}

		def.CpuUsage = cpu
		def.RamUsage = machine.Memory - ram
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

	_, err = c.HumanMonitorCommand(fmt.Sprintf("savevm checkpoint_%s", checkpoint))
	if err != nil {
		return err
	}

	return nil
}

// MachineKvmListCheckpoints returns the list of
// the specified machine's checkpoints
func MachineKvmListCheckpoints(id string) ([]shared.CheckpointDef, error) {
	chks := make([]shared.CheckpointDef, 0)

	disk, err := qemu.OpenImage(MachineDisk(id))
	if err != nil {
		return nil, err
	}

	snaps, err := disk.Snapshots()
	if err != nil {
		return nil, err
	}

	for _, snap := range snaps {
		if strings.HasPrefix(snap.Name, "checkpoint_") {
			chks = append(chks, shared.CheckpointDef{snap.Name[11:], snap.Date.Unix()})
		}
	}

	return chks, nil
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

	_, err = c.HumanMonitorCommand(fmt.Sprintf("loadvm checkpoint_%s", checkpoint))
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

	err = img.DeleteSnapshot(fmt.Sprintf("checkpoint_%s", checkpoint))
	if err != nil {
		return err
	}

	return nil
}

func MachineKvmDelete(id string) error {
	if MachineKvmIsRunning(id) {
		return fmt.Errorf("Machine is running")
	}

	err := os.RemoveAll(MachinePath(id))
	if err != nil {
		return err
	}

	return nil
}
