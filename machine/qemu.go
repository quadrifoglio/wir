package machine

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/amoghe/go-crypt"

	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/net"
	"github.com/quadrifoglio/wir/utils"
)

func QemuCreate(imgCmd, basePath, name string, img *image.Image, cores, memory int) (Machine, error) {
	var m Machine
	m.Name = name
	m.Type = img.Type
	m.Image = img.Name
	m.Cores = cores
	m.Memory = memory

	path := basePath + "qemu/" + name + ".qcow2"

	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return m, fmt.Errorf("mkdirall: %s", err)
	}

	var cmd *exec.Cmd

	if _, err := os.Stat(path); os.IsNotExist(err) {
		cmd = exec.Command(imgCmd, "create", "-b", img.Source, "-f", "qcow2", path)
	} else {
		cmd = exec.Command(imgCmd, "rebase", "-b", img.Source, path)
	}

	err = cmd.Run()
	if err != nil {
		return m, fmt.Errorf("qemu-img: %s", err)
	}

	return m, nil
}

func QemuStart(qemuCmd string, kvm bool, m *Machine, basePath string) error {
	m.State = StateDown

	args := make([]string, 8)
	args[0] = "-m"
	args[1] = strconv.Itoa(m.Memory)
	args[2] = "-smp"
	args[3] = strconv.Itoa(m.Cores)
	args[4] = "-hda"
	args[5] = basePath + "qemu/" + m.Name + ".qcow2"
	args[6] = "-vnc"
	args[7] = fmt.Sprintf(":%d", m.Index)

	if kvm {
		args = append(args, "-enable-kvm")
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

	cmd := exec.Command(qemuCmd, args...)

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

	m.Qemu.PID = cmd.Process.Pid
	return nil
}

func QemuLinuxSysprep(basePath, qemuNbd string, m *Machine, mainPart int, hostname, root string) error {
	// TODO: Find a way to use an available NBD device (/dev/nbdX)
	cmd := exec.Command(qemuNbd, "-c", "/dev/nbd0", basePath+"qemu/"+m.Name+".qcow2")

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("qemu-nbd: %s", err)
	}

	defer exec.Command(qemuNbd, "-d", "/dev/nbd0").Run()

	cmd = exec.Command("partx", "-a", "/dev/nbd0")

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("partx: %s", err)
	}

	path := "/tmp/wir/machines/" + m.Name

	err = os.MkdirAll(path, 0777)
	if err != nil {
		return fmt.Errorf("mkdir: %s", err)
	}

	cmd = exec.Command("mount", fmt.Sprintf("/dev/nbd0p%d", mainPart), path)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("mount: %s", err)
	}

	defer exec.Command("umount", path).Run()

	hostnameFile, err := os.OpenFile(path+"/etc/hostname", os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return fmt.Errorf("open /etc/hostname: %s", err)
	}

	fmt.Fprintf(hostnameFile, hostname)
	hostnameFile.Close()

	shadowFile, err := os.OpenFile(path+"/etc/shadow", os.O_RDWR, 0640)
	if err != nil {
		return fmt.Errorf("open /etc/shadow: %s", err)
	}

	defer shadowFile.Close()

	data, err := ioutil.ReadAll(shadowFile)
	if err != nil {
		return fmt.Errorf("/etc/shadow: can not read entire file: %s", err)
	}

	n := strings.Index(string(data), ":")
	if n == -1 {
		return fmt.Errorf("/etc/shadow: invalid file (no ':' char)")
	}

	nn := strings.Index(string(data[n+1:]), ":")
	if n == -1 {
		return fmt.Errorf("/etc/shadow: invalid file (no second ':' char)")
	}

	n = n + nn + 1

	// TODO: Random salt
	str, err := crypt.Crypt(root, "$6$HoN0Q1DH$")
	if err != nil {
		return fmt.Errorf("can not crypt password: %s", err)
	}

	str = "root:" + str

	mdr := make([]byte, len(str))
	copy(mdr, str)
	mdr = append(mdr, data[n:]...)

	_, err = shadowFile.WriteAt(mdr, 0)
	if err != nil {
		return fmt.Errorf("can not write to /etc/shadow: %s", err)
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

	// TODO: Increase precision: determine on which core(s) qemu is running
	stats.CPU = (float32(mtime2-mtime1) / float32(s2.Total-s1.Total)) * 100 * float32(runtime.NumCPU()) / float32(m.Cores)

	return stats, nil
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

	return nil
}
