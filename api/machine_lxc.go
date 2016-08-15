package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"gopkg.in/lxc/go-lxc.v2"

	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/net"
	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/utils"
)

type LxcMachine struct {
	shared.MachineInfo

	Interfaces []shared.NetworkDevice
}

func (m *LxcMachine) Info() *shared.MachineInfo {
	return &m.MachineInfo
}

func (m *LxcMachine) Type() string {
	return shared.TypeLXC
}

func (m *LxcMachine) Create(img shared.Image, info shared.MachineInfo) error {
	m.Name = info.Name
	m.Image = img.Name
	m.Cores = info.Cores
	m.Memory = info.Memory
	m.Disk = info.Disk

	path := shared.MachinePath("lxc")
	rootfs := fmt.Sprintf("%s/%s/rootfs", path, m.Name)

	err := os.MkdirAll(path, 0777)
	if err != nil {
		return err
	}

	if shared.IsStorage("zfs") {
		ds := fmt.Sprintf("%s/%s", shared.APIConfig.ZfsPool, m.Name)

		err := utils.ZfsCreate(ds, rootfs)
		if err != nil {
			return err
		}

		if m.Disk != 0 {
			err = utils.ZfsSet(ds, "quota", strconv.Itoa(int(m.Disk)))
			if err != nil {
				return err
			}
		}
	} else if shared.IsStorage("dir") {
		err := os.MkdirAll(rootfs, 0775)
		if err != nil {
			return err
		}
	}

	basePath := fmt.Sprintf("%s/%s", path, m.Name)
	mig := fmt.Sprintf("%s/%s.tar.gz", shared.APIConfig.MigrationPath, m.Name)

	if utils.FileExists(mig) {
		// Tthis is a migration
		err := utils.UntarDirectory(mig, basePath)
		if err != nil {
			return err
		}

		err = os.Remove(mig)
		if err != nil {
			return err
		}
	} else {
		err = utils.UntarDirectory(img.Source, basePath)
		if err != nil {
			return err
		}

		err := utils.ReplaceInFile(fmt.Sprintf("%s/config", basePath), "LXC_TEMPLATE_CONFIG", "/usr/share/lxc/config")
		if err != nil {
			return err
		}
	}

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return err
	}

	c.SetVerbosity(lxc.Verbose)

	if err = c.SetLogFile(fmt.Sprintf("%s/%s/log.txt", path, m.Name)); err != nil {
		return err
	}

	if err := c.SetConfigItem("lxc.rootfs.backend", shared.APIConfig.StorageBackend); err != nil {
		return fmt.Errorf("lxc.rootfs.backend config: %s", err)
	}
	if err := c.SetConfigItem("lxc.rootfs", rootfs); err != nil {
		return fmt.Errorf("lxc.rootfs config: %s", err)
	}
	if err := c.SetConfigItem("lxc.console", "none"); err != nil {
		return fmt.Errorf("lxc.console config: %s", err)
	}
	if err := c.SetConfigItem("lxc.tty", "0"); err != nil {
		return fmt.Errorf("lxc.tty config: %s", err)
	}
	if err := c.SetConfigItem("lxc.cgroup.devices.deny", "c 5:1 rwm"); err != nil {
		return fmt.Errorf("lxc.cgroup.devices.deny config: %s", err)
	}
	if err := c.SetConfigItem("lxc.utsname", m.Name); err != nil {
		return fmt.Errorf("lxc.utsname config: %s", err)
	}

	if err := c.SaveConfigFile(fmt.Sprintf("%s/%s/config", path, m.Name)); err != nil {
		return fmt.Errorf("can not write config: %s", err)
	}

	return nil
}

func (m *LxcMachine) Update(info shared.MachineInfo) error {
	if info.Cores != 0 && info.Cores != m.Cores {
		m.Cores = info.Cores
	}

	if info.Memory != 0 && info.Memory != m.Memory {
		m.Memory = info.Memory
	}

	return nil
}

func (m *LxcMachine) Delete() error {
	if m.State() != shared.StateDown {
		return errors.InvalidMachineState
	}

	base := fmt.Sprintf("%s/%s", shared.MachinePath("lxc"), m.Name)

	if shared.IsStorage("zfs") {
		err := utils.ZfsDestroy(fmt.Sprintf("%s/%s", shared.APIConfig.ZfsPool, m.Name))
		if err != nil {
			return err
		}
	}

	err := os.RemoveAll(base)
	if err != nil {
		return err
	}

	for _, iface := range m.ListInterfaces() {
		err := net.DeleteInterface(iface)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *LxcMachine) Sysprep(os, hostname, root string) error {
	if m.State() != shared.StateDown {
		return errors.InvalidMachineState
	}

	path := shared.MachinePath("lxc")

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return err
	}

	if err := c.SetConfigItem("lxc.utsname", hostname); err != nil {
		return err
	}

	err = utils.ChangeRootPassword(fmt.Sprintf("%s/%s/rootfs/etc/shadow", path, m.Name), root)
	if err != nil {
		return err
	}

	return nil
}

func (m *LxcMachine) Start() error {
	if m.State() != shared.StateDown {
		return errors.InvalidMachineState
	}

	fs := fmt.Sprintf("%s/%s", shared.APIConfig.ZfsPool, m.Name)

	if shared.IsStorage("zfs") {
		mounted, err := utils.ZfsIsMounted(fs)
		if err != nil {
			return err
		}

		if !mounted {
			err := utils.ZfsMount(fs)
			if err != nil {
				return err
			}
		}
	}

	path := shared.MachinePath("lxc")

	err := utils.DeleteLinesInFile(fmt.Sprintf("%s/%s/config", path, m.Name), "lxc.network")
	if err != nil {
		return err
	}

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return err
	}

	for i, iface := range m.ListInterfaces() {
		if iface.Mode == shared.NetworkModeBridge {
			if err := c.SetConfigItem("lxc.network.type", "veth"); err != nil {
				return err
			}
			if err := c.SetConfigItem("lxc.network.veth.pair", MachineIfName(m, i)); err != nil {
				return err
			}
			if err := c.SetConfigItem("lxc.network.flags", "up"); err != nil {
				return err
			}
			if err := c.SetConfigItem("lxc.network.link", "wir0"); err != nil {
				return err
			}
			if err := c.SetConfigItem("lxc.network.hwaddr", iface.MAC); err != nil {
				return err
			}

			err = net.CheckInterface(iface)
			if err != nil {
				return err
			}

			if err := c.SaveConfigFile(fmt.Sprintf("%s/%s/config", path, m.Name)); err != nil {
				return fmt.Errorf("can not write config: %s", err)
			}
		}
	}

	if m.HasCheckpoint("wird_migration") {
		err := m.RestoreCheckpoint("wird_migration")
		if err != nil {
			return err
		}

		err = m.DeleteCheckpoint("wird_migration")
		if err != nil {
			return err
		}
	} else {
		cmd := exec.Command("lxc-start", "-P", path, "-n", m.Name)

		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to restore container")
		}
	}

	if shared.APIConfig.EnableNetMonitor && len(m.Interfaces) > 0 {
		MonitorNetwork(m)
	}

	/*err = c.SetMemoryLimit(lxc.ByteSize(m.Memory) * lxc.MB)
	if err != nil {
		return err
	}*/

	return nil
}

func (m *LxcMachine) Stop() error {
	if m.State() != shared.StateUp {
		return errors.InvalidMachineState
	}

	path := shared.MachinePath("lxc")

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return err
	}

	c.SetVerbosity(lxc.Verbose)

	err = c.Stop()
	if err != nil {
		return err
	}

	return nil
}

func (m *LxcMachine) State() shared.MachineState {
	path := shared.MachinePath("lxc")

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		log.Println("state: %s", err)
	}

	s := c.State()
	if s == lxc.RUNNING {
		return shared.StateUp
	}

	return shared.StateDown
}

func (m *LxcMachine) Stats() (shared.MachineStats, error) {
	var stats shared.MachineStats

	if m.State() != shared.StateUp {
		return stats, errors.InvalidMachineState
	}

	path := shared.MachinePath("lxc")

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return stats, err
	}

	c1, err := c.CPUStats()
	if err != nil {
		return stats, err
	}

	c1time := c1["user"] + c1["system"]

	s1, err := utils.GetCpuUsage()
	if err != nil {
		return stats, err
	}

	time.Sleep(50 * time.Millisecond)

	c2, err := c.CPUStats()
	if err != nil {
		return stats, err
	}

	c2time := c2["user"] + c2["system"]

	s2, err := utils.GetCpuUsage()
	if err != nil {
		return stats, err
	}

	stats.CPU = (float32(c2time-c1time) / float32(s2.Total-s1.Total)) * 100

	mem, err := c.MemoryUsage()
	if err != nil {
		return stats, err
	}

	stats.RAMUsed = uint64(mem * lxc.MB)
	stats.RAMFree = uint64(m.Memory) - stats.RAMUsed

	return stats, nil
}

func (m *LxcMachine) ListInterfaces() []shared.NetworkDevice {
	return m.Interfaces
}

func (m *LxcMachine) CreateInterface(iface shared.NetworkDevice) (shared.NetworkDevice, error) {
	err := net.SetupInterface(&iface)
	if err != nil {
		return iface, err
	}

	m.Interfaces = append(m.Interfaces, iface)
	return iface, nil
}

func (m *LxcMachine) UpdateInterface(index int, iface shared.NetworkDevice) (shared.NetworkDevice, error) {
	if index >= len(m.Interfaces) {
		return iface, fmt.Errorf("interface index %d dost not exist", index)
	}

	miface := &m.Interfaces[index]

	err := net.DeleteInterface(*miface)
	if err != nil {
		return iface, err
	}

	if len(iface.IP) > 0 && iface.IP != miface.IP {
		miface.IP = iface.IP
	}
	if len(iface.MAC) > 0 && iface.MAC != miface.MAC {
		miface.MAC = iface.MAC
	}

	err = net.SetupInterface(miface)
	if err != nil {
		return iface, err
	}

	return *miface, nil
}

func (m *LxcMachine) DeleteInterface(index int) error {
	if index >= len(m.Interfaces) {
		return fmt.Errorf("interface index %d dost not exist", index)
	}

	err := net.DeleteInterface(m.Interfaces[index])
	if err != nil {
		return err
	}

	m.Interfaces = append(m.Interfaces[:index], m.Interfaces[index+1:]...)
	return nil
}

func (m *LxcMachine) ListVolumes() ([]shared.Volume, error) {
	return nil, fmt.Errorf("volumes not supported for lxc machines")
}

func (m *LxcMachine) CreateVolume(shared.Volume) error {
	return fmt.Errorf("volumes not supported for lxc machines")
}

func (m *LxcMachine) DeleteVolume(name string) error {
	return fmt.Errorf("volumes not supported for lxc machines")
}

func (m *LxcMachine) ListBackups() ([]shared.MachineBackup, error) {
	bks := make([]shared.MachineBackup, 0)

	if shared.IsStorage("dir") {
		path := fmt.Sprintf("%s/%s", shared.MachinePath("lxc"), m.Name)

		files, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}

		for _, f := range files {
			n := f.Name()
			bn := fmt.Sprintf("backup_")

			if strings.HasPrefix(n, bn) {
				b, err := strconv.ParseInt(n[7:], 10, 64)
				if err != nil {
					return nil, err
				}

				bks = append(bks, shared.MachineBackup(b))
			}
		}
	} else if shared.IsStorage("zfs") {
		ds := fmt.Sprintf("%s/%s", shared.APIConfig.ZfsPool, m.Name)

		sns, err := utils.ZfsListSnapshots(ds)
		if err != nil {
			return nil, err
		}

		for _, s := range sns {
			t, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return nil, err
			}

			bks = append(bks, shared.MachineBackup(t))
		}
	}

	return bks, nil
}

func (m *LxcMachine) CreateBackup() (shared.MachineBackup, error) {
	var b shared.MachineBackup

	if m.State() != shared.StateDown {
		return b, errors.InvalidMachineState
	}

	t := time.Now().Unix()

	if shared.IsStorage("dir") {
		path := shared.MachinePath("lxc")
		now := time.Now().Unix()

		src := fmt.Sprintf("%s/%s/rootfs", path, m.Name)
		dst := fmt.Sprintf("%s/%s/backup_%d", path, m.Name, now)

		err := utils.CopyFolder(src, dst)
		if err != nil {
			return b, err
		}
	} else if shared.IsStorage("zfs") {
		ds := fmt.Sprintf("%s/%s", shared.APIConfig.ZfsPool, m.Name)

		err := utils.ZfsSnapshot(ds, strconv.FormatInt(t, 10))
		if err != nil {
			return b, err
		}
	}

	return shared.MachineBackup(t), nil
}

func (m *LxcMachine) RestoreBackup(name string) error {
	if m.State() != shared.StateDown {
		return errors.InvalidMachineState
	}

	if shared.IsStorage("dir") {
		path := shared.MachinePath("lxc")
		rootfs := fmt.Sprintf("%s/%s/rootfs", path, m.Name)

		err := os.RemoveAll(rootfs)
		if err != nil {
			return err
		}

		err = utils.CopyFolder(fmt.Sprintf("%s/%s/backup_%s", path, m.Name, name), rootfs)
		if err != nil {
			return err
		}
	} else if shared.IsStorage("zfs") {
		ds := fmt.Sprintf("%s/%s", shared.APIConfig.ZfsPool, m.Name)

		err := utils.ZfsRestore(ds, name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *LxcMachine) DeleteBackup(name string) error {
	path := shared.MachinePath("lxc")

	if shared.IsStorage("dir") {
		err := os.RemoveAll(fmt.Sprintf("%s/%s_backup_%s", path, m.Name, name))
		if err != nil {
			return err
		}
	} else if shared.IsStorage("zfs") {
		ds := fmt.Sprintf("%s/%s", shared.APIConfig.ZfsPool, m.Name)

		err := utils.ZfsDeleteSnapshot(ds, name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *LxcMachine) ListCheckpoints() ([]string, error) {
	path := shared.MachinePath("lxc")

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	chks := make([]string, 0)
	for _, f := range files {
		n := f.Name()

		if strings.HasPrefix(n, "checkpoint_") && len(n) > 11 {
			chks = append(chks, n[11:])
		}
	}

	return chks, nil
}

func (m *LxcMachine) HasCheckpoint(name string) bool {
	path := fmt.Sprintf("%s/%s/checkpoint_%s", shared.MachinePath("lxc"), m.Name, name)

	if utils.FileExists(path) {
		return true
	}

	return false
}

func (m *LxcMachine) CreateCheckpoint(name string) error {
	if m.State() != shared.StateUp {
		return errors.InvalidMachineState
	}

	path := shared.MachinePath("lxc")

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return err
	}

	c.SetVerbosity(lxc.Verbose)

	err = c.Freeze()
	if err != nil {
		return err
	}

	path = fmt.Sprintf("%s/%s/checkpoint_%s", path, m.Name, name)

	if utils.FileExists(path) {
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	} else {
		err := os.MkdirAll(path, 0777)
		if err != nil {
			return err
		}
	}

	err = c.Checkpoint(lxc.CheckpointOptions{path, false, true})
	if err != nil {
		return err
	}

	err = c.Unfreeze()
	if err != nil {
		return err
	}

	return nil
}

func (m *LxcMachine) RestoreCheckpoint(name string) error {
	if m.State() != shared.StateDown {
		return errors.InvalidMachineState
	}

	if !m.HasCheckpoint(name) {
		return fmt.Errorf("checkpoint does not exists")
	}

	path := shared.MachinePath("lxc")
	chk := fmt.Sprintf("%s/%s/checkpoint_%s", path, m.Name, name)

	cmd := exec.Command("lxc-checkpoint", "-P", path, "-r", "-D", chk, "-n", m.Name)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to restore container")
	}

	err = os.RemoveAll(chk)
	if err != nil {
		return err
	}

	return nil
}

func (m *LxcMachine) DeleteCheckpoint(name string) error {
	path := fmt.Sprintf("%s/%s/checkpoint_%s", shared.MachinePath("lxc"), m.Name, name)
	return os.RemoveAll(path)
}

func (m *LxcMachine) MarshalJSON() ([]byte, error) {
	type machine struct {
		LxcMachine

		Type  string
		State shared.MachineState
	}

	return json.Marshal(machine{*m, m.Type(), m.State()})
}
