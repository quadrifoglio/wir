package machine

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/image"
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
		fmt.Println("mdr")
		args = append(args, "-enable-kvm")
	}

	tap, err := NetOpenTAP(m.IfName())
	if err != nil {
		return fmt.Errorf("open tap: %s", err)
	}

	err = NetTAPPersist(tap, true)
	if err != nil {
		return fmt.Errorf("tap persist: %s", err)
	}

	tap.Close()

	if len(m.Network.BridgeOn) > 0 {
		err := NetCreateBridge("wir0")
		if err != nil {
			return err
		}

		err = NetBridgeAddIf("wir0", m.IfName())
		if err != nil {
			return err
		}

		err = NetBridgeAddIf("wir0", m.Network.BridgeOn)
		if err != nil {
			return err
		}

		args = append(args, "-netdev")
		args = append(args, fmt.Sprintf("tap,id=net0,ifname=%s,script=no", m.IfName()))
		args = append(args, "-device")
		args = append(args, "driver=virtio-net,netdev=net0")
	}

	cmd := exec.Command(qemuCmd, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("qemu's stdout: %s", err)
	}

	go func() {
		in := bufio.NewScanner(stdout)
		for in.Scan() {
			log.Printf("Qemu machine %s: %s\n", m.Name, in.Text())
		}
	}()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("qemu's stderr: %s", err)
	}

	go func() {
		in := bufio.NewScanner(stderr)
		for in.Scan() {
			log.Printf("Qemu machine %s: %s\n", m.Name, in.Text())
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

		log.Printf("Qemu machine %s: process exited: %s", m.Name, errs)
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
	if len(m.Network.BridgeOn) != 0 {
		tap, err := NetOpenTAP(m.IfName())
		if err != nil {
			return fmt.Errorf("open tap: %s", err)
		}

		err = NetTAPPersist(tap, false)
		if err != nil {
			return fmt.Errorf("tap persist: %s", err)
		}

		tap.Close()
	}

	return nil
}
