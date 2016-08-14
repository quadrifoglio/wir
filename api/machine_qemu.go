package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/quadrifoglio/go-qmp"

	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/net"
	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/utils"
)

type QemuMachine struct {
	shared.MachineInfo

	PID      int
	MainPart string
}

func (m *QemuMachine) Info() *shared.MachineInfo {
	return &m.MachineInfo
}

func (m *QemuMachine) Type() string {
	return shared.TypeQemu
}

func (m *QemuMachine) Create(img shared.Image, info shared.MachineInfo) error {
	m.Name = info.Name
	m.Image = img.Name
	m.Cores = info.Cores
	m.Memory = info.Memory
	m.Disk = info.Disk

	path := fmt.Sprintf("%s/%s", shared.MachinePath("qemu"), m.Name)
	disk := fmt.Sprintf("%s/disk.qcow2", path)

	if shared.IsStorage("zfs") {
		ds := fmt.Sprintf("%s/%s", shared.APIConfig.ZfsPool, m.Name)

		err := utils.ZfsCreate(ds, path)
		if err != nil {
			return err
		}
	} else if shared.IsStorage("dir") {
		err := os.MkdirAll(path, 0775)
		if err != nil {
			return err
		}
	}

	mig := fmt.Sprintf("%s/%s.tar.gz", shared.APIConfig.MigrationPath, m.Name)

	// If the migration archive exists, extract it and then remove it
	if utils.FileExists(mig) {
		err := utils.UntarDirectory(mig, path)
		if err != nil {
			return err
		}

		err = os.Remove(mig)
		if err != nil {
			return err
		}
	}

	// If the disk file already exists, it means that this is a migration
	// In that case, we should rebase the disk to correspond to the local image
	if utils.FileExists(disk) {
		err := utils.RebaseQcow2(disk, img.Source)
		if err != nil {
			return err
		}
	} else {
		err := utils.CreateQcow2(disk, img.Source)
		if err != nil {
			return err
		}
	}

	// Check the image's size
	imgSize, err := utils.SizeQcow2(img.Source)
	if err != nil {
		return err
	}

	// If the machine should have a disk space limit, and that limit is greater than
	// the image size, then resize the disk and it's main partition to fit the limit
	if m.Disk != 0 && m.Disk > imgSize {
		err = utils.ResizeQcow2(disk, m.Disk)
		if err != nil {
			return err
		}

		err = utils.NBDConnectQcow2(disk)
		if err != nil {
			return err
		}

		defer utils.NBDDisconnectQcow2()

		err = utils.ResizePartition("/dev/nbd0", img.MainPartition, m.Disk)
		if err != nil {
			return err
		}
	}

	var i int = 0
	for _, iface := range info.Interfaces {
		if iface.Mode != shared.NetworkModeNone {
			m.Interfaces = append(m.Interfaces, iface)

			err := net.SetupInterface(&m.Interfaces[i])
			if err != nil {
				return err
			}

			i++
		}
	}

	return nil
}

func (m *QemuMachine) Update(info shared.MachineInfo) error {
	if info.Cores != 0 && info.Cores != m.Cores {
		m.Cores = info.Cores
	}

	if info.Memory != 0 && info.Cores != m.Cores {
		m.Memory = info.Cores
	}

	if len(info.Interfaces) > 0 {
		for i, iface := range info.Interfaces {
			if len(m.Interfaces) > i {
				miface := &m.Interfaces[i]

				err := net.DeleteInterface(*miface)
				if err != nil {
					return err
				}

				if len(iface.Mode) > 0 && iface.Mode != miface.Mode {
					miface.Mode = iface.Mode
				}
				if len(iface.MAC) > 0 {
					miface.MAC = iface.MAC
				}
				if len(iface.IP) > 0 {
					miface.IP = iface.IP
				}

				err = net.SetupInterface(&m.Interfaces[i])
				if err != nil {
					return err
				}
			} else {
				m.Interfaces = append(m.Interfaces, iface)
				err := net.SetupInterface(&m.Interfaces[len(m.Interfaces)-1])
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (m *QemuMachine) Delete() error {
	if m.State() != shared.StateDown {
		return errors.InvalidMachineState
	}

	if shared.IsStorage("zfs") {
		err := utils.ZfsDestroy(fmt.Sprintf("%s/%s", shared.APIConfig.ZfsPool, m.Name))
		if err != nil {
			return err
		}
	} else if shared.IsStorage("dir") {
		err := os.RemoveAll(fmt.Sprintf("%s/%s", shared.MachinePath("qemu"), m.Name))
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *QemuMachine) Sysprep(os, hostname, root string) error {
	if m.State() != shared.StateDown {
		return errors.InvalidMachineState
	}

	disk := fmt.Sprintf("%s/%s/disk.qcow2", shared.MachinePath("qemu"), m.Name)

	path := "/tmp/wir/machines/" + m.Name
	hostnameFile := path + "/etc/hostname"
	shadowFile := path + "/etc/shadow"

	err := utils.NBDConnectQcow2(disk)
	if err != nil {
		return err
	}

	defer utils.NBDDisconnectQcow2()

	err = utils.Mount(fmt.Sprintf("/dev/nbd0p%d", m.MainPart), path)
	if err != nil {
		return err
	}

	defer utils.Unmount(path)

	err = utils.ChangeHostname(hostnameFile, hostname)
	if err != nil {
		return err
	}

	err = utils.ChangeRootPassword(shadowFile, root)
	if err != nil {
		return err
	}

	// TODO: remove ssh keys

	return nil
}

func (m *QemuMachine) Start() error {
	if m.State() != shared.StateDown {
		return errors.InvalidMachineState
	}

	args := make([]string, 10)
	args[0] = "-m"
	args[1] = strconv.Itoa(m.Memory)
	args[2] = "-smp"
	args[3] = strconv.Itoa(m.Cores)
	args[4] = "-hda"
	args[5] = fmt.Sprintf("%s/%s/disk.qcow2", shared.MachinePath("qemu"), m.Name)
	args[6] = "-vnc"
	args[7] = fmt.Sprintf(":%d", m.Index)
	args[8] = "-qmp"
	args[9] = fmt.Sprintf("unix:%s/%s/qmp.sock,server,nowait", shared.MachinePath("qemu"), m.Name)

	if shared.APIConfig.EnableKVM {
		args = append(args, "-enable-kvm")
	}

	if m.HasCheckpoint("checkpoint_wird_migration") {
		args = append(args, "-loadvm")
		args = append(args, "checkpoint_wird_migration")
	}

	if len(m.Interfaces) == 0 {
		args = append(args, "-net")
		args = append(args, "none")
	}

	for i, iface := range m.Interfaces {
		if iface.Mode == shared.NetworkModeBridge {
			tap, err := net.OpenTAP(MachineIfName(m, i))
			if err != nil {
				return fmt.Errorf("open tap: %s", err)
			}

			err = net.TAPPersist(tap, true)
			if err != nil {
				tap.Close()
				return fmt.Errorf("tap persist: %s", err)
			}

			tap.Close()

			err = net.BridgeAddIf("wir0", MachineIfName(m, i))
			if err != nil {
				return err
			}

			err = net.CheckInterface(iface)
			if err != nil {
				return err
			}

			args = append(args, "-netdev")
			args = append(args, fmt.Sprintf("tap,id=net0,ifname=%s,script=no", MachineIfName(m, i)))
			args = append(args, "-device")
			args = append(args, fmt.Sprintf("driver=virtio-net,netdev=net0,mac=%s", iface.MAC))
		}
	}

	vols, err := m.ListVolumes()
	if err != nil {
		return err
	}

	for _, vol := range vols {
		args = append(args, "-drive")
		args = append(args, fmt.Sprintf("file=%s/%s/volume_%s.qcow2", shared.MachinePath("qemu"), m.Name, vol.Name))
	}

	cmd := exec.Command("qemu-system-x86_64", args...)
	cmd.SysProcAttr = new(syscall.SysProcAttr)
	cmd.SysProcAttr.Setsid = true

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("qemu's stdout: %s", err)
	}

	go func() {
		in := bufio.NewScanner(stdout)
		for in.Scan() {
			log.Printf("error in qemu machine %s: %s\n", m.Name, in.Text())
		}
	}()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("qemu's stderr: %s", err)
	}

	go func() {
		in := bufio.NewScanner(stderr)
		for in.Scan() {
			log.Printf("error in qemu machine %s: %s\n", m.Name, in.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("qemu: %s", err)
	}

	errc := make(chan bool)

	go func() {
		err := cmd.Wait()

		var errs string
		if err != nil {
			errs = err.Error()
		} else {
			errs = "exit status 0"
		}

		log.Printf("qemu machine %s: process exited: %s", m.Name, errs)
		errc <- true
	}()

	time.Sleep(500 * time.Millisecond)

	select {
	case <-errc:
		return errors.StartFailed
	default:
		break
	}

	if m.HasCheckpoint("wird_migration") {
		err := m.DeleteCheckpoint("wird_migration")
		if err != nil {
			return err
		}
	}

	if shared.APIConfig.EnableNetMonitor && len(m.Interfaces) > 0 {
		MonitorNetwork(m)
	}

	m.PID = cmd.Process.Pid
	return nil
}

func (m *QemuMachine) Stop() error {
	if m.State() != shared.StateUp {
		return errors.InvalidMachineState
	}

	for i, _ := range m.Interfaces {
		tap, err := net.OpenTAP(MachineIfName(m, i))
		if err != nil {
			continue
		}

		net.TAPPersist(tap, false)
		tap.Close()
	}

	proc, err := os.FindProcess(m.PID)
	if err != nil {
		return nil
	}

	err = proc.Kill()
	if err != nil {
		return errors.KillFailed
	}

	m.PID = 0
	return nil
}

func (m *QemuMachine) State() shared.MachineState {
	out, err := exec.Command("kill", "-s", "0", strconv.Itoa(m.PID)).CombinedOutput()
	if m.PID == 0 || err != nil {
		m.PID = 0
		return shared.StateDown
	}

	if string(out) == "" {
		return shared.StateUp
	}

	log.Println(string(out))

	m.PID = 0
	return shared.StateDown
}

func (m *QemuMachine) Stats() (shared.MachineStats, error) {
	var stats shared.MachineStats

	if m.State() != shared.StateUp {
		return stats, errors.InvalidMachineState
	}

	utime1, stime1, err := utils.GetProcessCpuStats(m.PID)
	if err != nil {
		return stats, err
	}

	mtime1 := utime1 + stime1

	s1, err := utils.GetCpuUsage()
	if err != nil {
		return stats, err
	}

	time.Sleep(100 * time.Millisecond)

	utime2, stime2, err := utils.GetProcessCpuStats(m.PID)
	if err != nil {
		return stats, err
	}

	mtime2 := utime2 + stime2

	s2, err := utils.GetCpuUsage()
	if err != nil {
		return stats, err
	}

	stats.CPU = (float32(mtime2-mtime1) / float32(s2.Total-s1.Total)) * 100

	ram, err := utils.GetProcessRamUsage(m.PID)
	if err != nil {
		return stats, err
	}

	stats.RAMUsed = ram

	return stats, nil
}

func (m *QemuMachine) ListVolumes() ([]shared.Volume, error) {
	vols := make([]shared.Volume, 0)
	path := fmt.Sprintf("%s/%s", shared.MachinePath("qemu"), m.Name)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if strings.HasPrefix(f.Name(), "volume_") && len(f.Name()) > 7 {
			size, err := utils.SizeQcow2(fmt.Sprintf("%s/%s", path, f.Name()))
			if err != nil {
				return nil, err
			}

			var vol shared.Volume
			vol.Name = f.Name()[7 : len(f.Name())-6]
			vol.Size = size

			vols = append(vols, vol)
		}
	}

	return vols, nil
}

func (m *QemuMachine) CreateVolume(vol shared.Volume) error {
	if m.State() != shared.StateDown {
		return errors.InvalidMachineState
	}

	path := fmt.Sprintf("%s/%s/volume_%s.qcow2", shared.MachinePath("qemu"), m.Name, vol.Name)

	err := utils.CreateBaseQcow2(path, vol.Size)
	if err != nil {
		return err
	}

	return nil
}

func (m *QemuMachine) DeleteVolume(name string) error {
	if m.State() != shared.StateDown {
		return errors.InvalidMachineState
	}

	path := fmt.Sprintf("%s/%s/volume_%s.qcow2", shared.MachinePath("qemu"), m.Name, name)

	err := os.Remove(path)
	if err != nil {
		return err
	}

	return nil
}

func (m *QemuMachine) ListCheckpoints() ([]string, error) {
	path := fmt.Sprintf("%s/%s/disk.qcow2", shared.MachinePath("qemu"), m.Name)

	snaps, err := utils.ListSnapshotsQcow2(path)
	if err != nil {
		return nil, err
	}

	chks := make([]string, 0)
	for _, s := range snaps {
		if strings.HasPrefix(s, "checkpoint_") && len(s) > 9 {
			chks = append(chks, s[11:])
		}
	}

	return chks, nil
}

func (m *QemuMachine) HasCheckpoint(name string) bool {
	path := fmt.Sprintf("%s/%s/disk.qcow2", shared.MachinePath("qemu"), m.Name)
	cmd := exec.Command("qemu-img", "snapshot", "-l", path)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	if strings.Contains(string(out), fmt.Sprintf("checkpoint_%s", name)) {
		return true
	}

	return false
}

func (m *QemuMachine) CreateCheckpoint(name string) error {
	if m.State() != shared.StateUp {
		return errors.InvalidMachineState
	}

	c, err := qmp.Open("unix", fmt.Sprintf("%s/%s/qmp.sock", shared.MachinePath("qemu"), m.Name))
	if err != nil {
		return err
	}

	defer c.Close()

	_, err = c.Command("stop", nil)
	if err != nil {
		return err
	}

	_, err = c.HumanMonitorCommand(fmt.Sprintf("savevm checkpoint_%s", name))
	if err != nil {
		return err
	}

	return nil
}

func (m *QemuMachine) RestoreCheckpoint(name string) error {
	if m.State() == shared.StateDown {
		err := m.Start()
		if err != nil {
			return err
		}
	}

	c, err := qmp.Open("unix", fmt.Sprintf("%s/%s/qmp.sock", shared.MachinePath("qemu"), m.Name))
	if err != nil {
		return err
	}

	defer c.Close()

	_, err = c.Command("stop", nil)
	if err != nil {
		return err
	}

	_, err = c.HumanMonitorCommand(fmt.Sprintf("loadvm checkpoint_%s", name))
	if err != nil {
		return err
	}

	_, err = c.Command("cont", nil)
	if err != nil {
		return err
	}

	return nil
}

func (m *QemuMachine) DeleteCheckpoint(name string) error {
	path := fmt.Sprintf("%s/%s/disk.qcow2", shared.MachinePath("qemu"), m.Name)
	cmd := exec.Command("qemu-img", "snapshot", "-d", fmt.Sprintf("checkpoint_%s", name), path)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("delete checkpoint: %s", err)
	}

	return nil
}

func (m *QemuMachine) ListBackups() ([]shared.MachineBackup, error) {
	disk := fmt.Sprintf("%s/%s/disk.qcow2", shared.MachinePath("qemu"), m.Name)

	sns, err := utils.ListSnapshotsQcow2(disk)
	if err != nil {
		return nil, err
	}

	bks := make([]shared.MachineBackup, 0)
	for _, s := range sns {
		if strings.HasPrefix(s, "backup_") {
			t, err := strconv.ParseInt(s[7:], 10, 64)
			if err != nil {
				return nil, err
			}

			bks = append(bks, shared.MachineBackup(t))
		}
	}

	return bks, nil
}

func (m *QemuMachine) CreateBackup() (shared.MachineBackup, error) {
	var b shared.MachineBackup

	if m.State() != shared.StateDown {
		return b, errors.InvalidMachineState
	}

	disk := fmt.Sprintf("%s/%s/disk.qcow2", shared.MachinePath("qemu"), m.Name)

	now := time.Now().Unix()

	err := utils.SnapshotQcow2(disk, fmt.Sprintf("backup_%d", now))
	if err != nil {
		return b, err
	}

	return shared.MachineBackup(now), nil
}

func (m *QemuMachine) RestoreBackup(name string) error {
	if m.State() != shared.StateDown {
		return errors.InvalidMachineState
	}

	disk := fmt.Sprintf("%s/%s/disk.qcow2", shared.MachinePath("qemu"), m.Name)

	return utils.RestoreQcow2(disk, fmt.Sprintf("backup_%s", name))
}

func (m *QemuMachine) DeleteBackup(name string) error {
	disk := fmt.Sprintf("%s/%s/disk.qcow2", shared.MachinePath("qemu"), m.Name)
	return utils.DeleteSnapshotQcow2(disk, fmt.Sprintf("backup_%s", name))
}

func (m *QemuMachine) MarshalJSON() ([]byte, error) {
	type machine struct {
		QemuMachine

		Type  string
		State shared.MachineState
	}

	return json.Marshal(machine{*m, m.Type(), m.State()})
}
