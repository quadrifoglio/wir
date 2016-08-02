package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/quadrifoglio/go-qmp"

	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/net"
	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/utils"
)

var (
	sysprepMutex sync.Mutex
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

func (m *QemuMachine) Create(img Image, info shared.MachineInfo) error {
	m.Name = info.Name
	m.Image = img.Info().Name
	m.Cores = info.Cores
	m.Memory = info.Memory
	m.Disk = info.Disk

	path := fmt.Sprintf("%s/qemu/%s", shared.APIConfig.MachinePath, m.Name)
	disk := fmt.Sprintf("%s/disk.qcow2", path)

	if shared.APIConfig.StorageBackend == "zfs" {
		ds := fmt.Sprintf("%s/%s", shared.APIConfig.ZfsPool, m.Name)

		err := utils.ZfsCreate(ds, path)
		if err != nil {
			return err
		}

		time.Sleep(1 * time.Second)
	} else if shared.APIConfig.StorageBackend == "dir" {
		err := os.MkdirAll(path, 0775)
		if err != nil {
			return err
		}
	}

	var cmd *exec.Cmd

	if _, err := os.Stat(disk); os.IsNotExist(err) {
		cmd = exec.Command(shared.APIConfig.QemuImg, "create", "-b", img.Info().Source, "-f", "qcow2", disk, strconv.Itoa(m.Disk))
	} else {
		cmd = exec.Command(shared.APIConfig.QemuImg, "rebase", "-b", img.Info().Source, disk)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", out)
	}

	return SetupMachineNetwork(m, info.Network)
}

func (m *QemuMachine) Update(info shared.MachineInfo) error {
	if info.Cores != 0 && info.Cores != m.Cores {
		m.Cores = info.Cores
	}

	if info.Memory != 0 && info.Cores != m.Cores {
		m.Memory = info.Cores
	}

	return UpdateMachineNetwork(m, info.Network)
}

func (m *QemuMachine) Delete() error {
	if m.Network.Mode != shared.NetworkModeNone {
		tap, err := net.OpenTAP(MachineIfName(m))
		if err != nil {
			return fmt.Errorf("open tap: %s", err)
		}

		err = net.TAPPersist(tap, false)
		if err != nil {
			return fmt.Errorf("tap persist: %s", err)
		}

		tap.Close()
	}

	err := os.Remove(fmt.Sprintf("%s/qemu/%s", shared.APIConfig.MachinePath, m.Name))
	if err != nil {
		return fmt.Errorf("remove machine folder: %s", err)
	}

	return nil
}

func (m *QemuMachine) Sysprep(os, hostname, root string) error {
	sysprepMutex.Lock()
	defer sysprepMutex.Unlock()

	disk := fmt.Sprintf("%s/qemu/%s/disk.qcow2", shared.APIConfig.MachinePath, m.Name)

	path := "/tmp/wir/machines/" + m.Name
	hostnameFile := path + "/etc/hostname"
	shadowFile := path + "/etc/shadow"

	err := utils.NBDConnectQcow2(shared.APIConfig.QemuNbd, "/dev/nbd0", disk)
	if err != nil {
		return err
	}

	defer utils.NBDDisconnectQcow2(shared.APIConfig.QemuNbd, "/dev/nbd0")

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
	args := make([]string, 10)
	args[0] = "-m"
	args[1] = strconv.Itoa(m.Memory)
	args[2] = "-smp"
	args[3] = strconv.Itoa(m.Cores)
	args[4] = "-hda"
	args[5] = fmt.Sprintf("%s/qemu/%s/disk.qcow2", shared.APIConfig.MachinePath, m.Name)
	args[6] = "-vnc"
	args[7] = fmt.Sprintf(":%d", m.Index)
	args[8] = "-qmp"
	args[9] = fmt.Sprintf("unix:%s/qemu/%s/qmp.sock,server,nowait", shared.APIConfig.MachinePath, m.Name)

	if shared.APIConfig.EnableKVM {
		args = append(args, "-enable-kvm")
	}

	if m.HasCheckpoint() {
		args = append(args, "-loadvm")
		args = append(args, "checkpoint")
	}

	if m.Network.Mode == shared.NetworkModeBridge {
		tap, err := net.OpenTAP(MachineIfName(m))
		if err != nil {
			return fmt.Errorf("open tap: %s", err)
		}

		err = net.TAPPersist(tap, true)
		if err != nil {
			return fmt.Errorf("tap persist: %s", err)
		}

		tap.Close()

		err = net.BridgeAddIf("wir0", MachineIfName(m))
		if err != nil {
			return err
		}

		err = CheckMachineNetwork(m)
		if err != nil {
			return err
		}

		args = append(args, "-netdev")
		args = append(args, fmt.Sprintf("tap,id=net0,ifname=%s,script=no", MachineIfName(m)))
		args = append(args, "-device")
		args = append(args, fmt.Sprintf("driver=virtio-net,netdev=net0,mac=%s", m.Network.MAC))
	}

	cmd := exec.Command(shared.APIConfig.Qemu, args...)

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

	if m.HasCheckpoint() {
		err := m.DeleteCheckpoint()
		if err != nil {
			return err
		}
	}

	if shared.APIConfig.EnableNetMonitor && m.Network.Mode != shared.NetworkModeNone {
		MonitorNetwork(m)
	}

	m.PID = cmd.Process.Pid
	return nil
}

func (m *QemuMachine) Stop() error {
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

func (m *QemuMachine) HasCheckpoint() bool {
	path := fmt.Sprintf("%s/qemu/%s/disk.qcow2", shared.APIConfig.MachinePath, m.Name)
	cmd := exec.Command(shared.APIConfig.QemuImg, "snapshot", "-l", path)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	if strings.Contains(string(out), "checkpoint") {
		return true
	}

	return false
}

func (m *QemuMachine) CreateCheckpoint() error {
	c, err := qmp.Open("unix", fmt.Sprintf("%s/qemu/%s/qmp.sock", shared.APIConfig.MachinePath, m.Name))
	if err != nil {
		return err
	}

	defer c.Close()

	_, err = c.Command("stop", nil)
	if err != nil {
		return err
	}

	_, err = c.HumanMonitorCommand("savevm checkpoint")
	if err != nil {
		return err
	}

	return nil
}

func (m *QemuMachine) RestoreCheckpoint() error {
	c, err := qmp.Open("unix", fmt.Sprintf("%s/qemu/%s/qmp.sock", shared.APIConfig.MachinePath, m.Name))
	if err != nil {
		return err
	}

	defer c.Close()

	_, err = c.Command("stop", nil)
	if err != nil {
		return err
	}

	_, err = c.HumanMonitorCommand("loadvm checkpoint")
	if err != nil {
		return err
	}

	_, err = c.Command("cont", nil)
	if err != nil {
		return err
	}

	return nil
}

func (m *QemuMachine) DeleteCheckpoint() error {
	path := fmt.Sprintf("%s/qemu/%s/disk.qcow2", shared.APIConfig.MachinePath, m.Name)
	cmd := exec.Command(shared.APIConfig.QemuImg, "snapshot", "-d", "checkpoint", path)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("delete checkpoint: %s", err)
	}

	return nil
}

func (m *QemuMachine) MarshalJSON() ([]byte, error) {
	type mdr struct {
		QemuMachine

		Type  string
		State shared.MachineState
	}

	return json.Marshal(mdr{*m, m.Type(), m.State()})
}
