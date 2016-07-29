package machine

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/quadrifoglio/go-qmp"

	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/net"
	"github.com/quadrifoglio/wir/utils"
)

var (
	sysprepMutex sync.Mutex
)

func QemuCreate(m *Machine, name string, img image.Image, cores, memory int) error {
	path := fmt.Sprintf("%s/qemu/%s.qcow2", shared.APIConfig.MachinePath, name)

	m.Name = name
	m.Type = img.Type
	m.Image = img.Name
	m.Cores = cores
	m.Memory = memory
	m.State = StateDown

	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return fmt.Errorf("mkdirall: %s", err)
	}

	var cmd *exec.Cmd

	if _, err := os.Stat(path); os.IsNotExist(err) {
		cmd = exec.Command(shared.APIConfig.QemuImg, "create", "-b", img.Source, "-f", "qcow2", path)
	} else {
		cmd = exec.Command(shared.APIConfig.QemuImg, "rebase", "-b", img.Source, path)
	}

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("qemu-img: %s", err)
	}

	return nil
}

func QemuStart(m *Machine) error {
	m.State = StateDown

	args := make([]string, 10)
	args[0] = "-m"
	args[1] = strconv.Itoa(m.Memory)
	args[2] = "-smp"
	args[3] = strconv.Itoa(m.Cores)
	args[4] = "-hda"
	args[5] = fmt.Sprintf("%s/qemu/%s.qcow2", shared.APIConfig.MachinePath, m.Name)
	args[6] = "-vnc"
	args[7] = fmt.Sprintf(":%d", m.Index)
	args[8] = "-qmp"
	args[9] = fmt.Sprintf("unix:%s/qemu/%s.sock,server,nowait", shared.APIConfig.MachinePath, m.Name)

	if shared.APIConfig.EnableKVM {
		args = append(args, "-enable-kvm")
	}

	if QemuHasCheckpoint(m) {
		args = append(args, "-loadvm")
		args = append(args, "checkpoint")
	}

	if m.Network.Mode == NetworkModeBridge {
		tap, err := net.OpenTAP(m.IfName())
		if err != nil {
			return fmt.Errorf("open tap: %s", err)
		}

		err = net.TAPPersist(tap, true)
		if err != nil {
			return fmt.Errorf("tap persist: %s", err)
		}

		tap.Close()

		err = net.BridgeAddIf("wir0", m.IfName())
		if err != nil {
			return err
		}

		args = append(args, "-netdev")
		args = append(args, fmt.Sprintf("tap,id=net0,ifname=%s,script=no", m.IfName()))
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
		m.State = StateUp
		break
	}

	if QemuHasCheckpoint(m) {
		err := QemuDeleteCheckpoint(m)
		if err != nil {
			return err
		}
	}

	if shared.APIConfig.EnableNetMonitor && m.Network.Mode != NetworkModeNone {
		go func(m *Machine) {
			for {
				a := net.MonitorInterface(m.IfName(), "rx")

				m.Check()

				if m.State != StateUp {
					break
				}

				if a == net.MonitorCancel {
					break
				}
				if a == net.MonitorAlert {
					// TODO: Send email
				}
				if a == net.MonitorStop {
					// TODO: Send email

					log.Println("iface monitor %s: shuting down (to many pps)", m.IfName())

					err := QemuStop(m)
					if err != nil {
						log.Println(err)
					}

					break
				}

				time.Sleep(10 * time.Second)
			}
		}(m)
	}

	m.Qemu.PID = cmd.Process.Pid
	return nil
}

func QemuLinuxSysprep(m *Machine, mainPart int, hostname, root string) error {
	sysprepMutex.Lock()
	defer sysprepMutex.Unlock()

	path := "/tmp/wir/machines/" + m.Name
	hostnameFile := path + "/etc/hostname"
	shadowFile := path + "/etc/shadow"

	err := utils.NBDConnectQcow2(shared.APIConfig.QemuNbd, "/dev/nbd0", fmt.Sprintf("%s/qemu/%s.qcow2", shared.APIConfig.MachinePath, m.Name))
	if err != nil {
		return err
	}

	defer utils.NBDDisconnectQcow2(shared.APIConfig.QemuNbd, "/dev/nbd0")

	err = utils.Mount(fmt.Sprintf("/dev/nbd0p%d", mainPart), path)
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

func QemuCheck(m *Machine) {
	out, err := exec.Command("kill", "-s", "0", strconv.Itoa(m.Qemu.PID)).CombinedOutput()
	if m.Qemu.PID == 0 || err != nil {
		m.State = StateDown
		m.Qemu.PID = 0
		return
	}

	if string(out) == "" {
		m.State = StateUp
		return
	}

	log.Println(string(out))

	m.State = StateDown
	m.Qemu.PID = 0
}

func QemuStats(m *Machine) (Stats, error) {
	var stats Stats

	utime1, stime1, err := utils.GetProcessCpuStats(m.Qemu.PID)
	if err != nil {
		return stats, err
	}

	mtime1 := utime1 + stime1

	s1, err := utils.GetCpuUsage()
	if err != nil {
		return stats, err
	}

	time.Sleep(100 * time.Millisecond)

	utime2, stime2, err := utils.GetProcessCpuStats(m.Qemu.PID)
	if err != nil {
		return stats, err
	}

	mtime2 := utime2 + stime2

	s2, err := utils.GetCpuUsage()
	if err != nil {
		return stats, err
	}

	stats.CPU = (float32(mtime2-mtime1) / float32(s2.Total-s1.Total)) * 100

	ram, err := utils.GetProcessRamUsage(m.Qemu.PID)
	if err != nil {
		return stats, err
	}

	stats.RAMUsed = ram

	return stats, nil
}

func QemuHasCheckpoint(m *Machine) bool {
	path := fmt.Sprintf("%s/qemu/%s.qcow2", shared.APIConfig.MachinePath, m.Name)
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

func QemuCheckpoint(m *Machine) error {
	c, err := qmp.Open("unix", fmt.Sprintf("%s/qemu/%s.sock", shared.APIConfig.MachinePath, m.Name))
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

func QemuDeleteCheckpoint(m *Machine) error {
	path := fmt.Sprintf("%s/qemu/%s.qcow2", shared.APIConfig.MachinePath, m.Name)
	cmd := exec.Command(shared.APIConfig.QemuImg, "snapshot", "-d", "checkpoint", path)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("delete checkpoint: %s", err)
	}

	return nil
}

func QemuStop(m *Machine) error {
	proc, err := os.FindProcess(m.Qemu.PID)
	if err != nil {
		m.State = StateDown
		return nil
	}

	err = proc.Kill()
	if err != nil {
		return errors.KillFailed
	}

	m.State = StateDown
	m.Qemu.PID = 0

	return nil
}

func QemuDelete(m *Machine) error {
	if m.Network.Mode != NetworkModeNone {
		tap, err := net.OpenTAP(m.IfName())
		if err != nil {
			return fmt.Errorf("open tap: %s", err)
		}

		err = net.TAPPersist(tap, false)
		if err != nil {
			return fmt.Errorf("tap persist: %s", err)
		}

		tap.Close()
	}

	err := os.Remove(fmt.Sprintf("%s/qemu/%s.qcow2", shared.APIConfig.MachinePath, m.Name))
	if err != nil {
		return fmt.Errorf("remove disk file: %s", err)
	}

	return nil
}
