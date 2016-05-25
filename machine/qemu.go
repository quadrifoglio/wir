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
	"github.com/quadrifoglio/wir/utils"
)

func QemuCreate(basePath string, img *image.Image, cores, memory int) (Machine, error) {
	var m Machine
	m.ID = utils.UniqueID()
	m.Type = img.Type
	m.Image = img.Name
	m.Cores = cores
	m.Memory = memory

	path := basePath + "qemu/" + m.ID + ".img"

	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return m, err
	}

	cmd := exec.Command("qemu-img", "create", "-b", img.Source, "-f", "qcow2", path)

	err = cmd.Run()
	if err != nil {
		return m, err
	}

	return m, nil
}

func QemuStart(m *Machine, basePath string) error {
	m.State = StateDown

	args := make([]string, 7)
	args[0] = "-enable-kvm"
	args[1] = "-m"
	args[2] = strconv.Itoa(m.Memory)
	args[3] = "-smp"
	args[4] = strconv.Itoa(m.Cores)
	args[5] = "-hda"
	args[6] = basePath + "qemu/" + m.ID + ".img"

	if len(m.NetBridgeOn) > 0 {
		id := m.ID[:14]

		err := NetCreateBridge("wir0")
		if err != nil {
			return err
		}

		NetDeleteTAP(id)

		err = NetCreateTAP(id)
		if err != nil {
			return err
		}

		err = NetBridgeAddIf("wir0", id)
		if err != nil {
			return err
		}

		err = NetBridgeAddIf("wir0", m.NetBridgeOn)
		if err != nil {
			return err
		}

		args = append(args, "-netdev")
		args = append(args, fmt.Sprintf("tap,id=net0,ifname=%s,script=no", id))
		args = append(args, "-device")
		args = append(args, "driver=virtio-net,netdev=net0")
	}

	cmd := exec.Command("qemu-system-x86_64", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	go func() {
		in := bufio.NewScanner(stdout)
		for in.Scan() {
			log.Printf("Qemu machine %s: %s\n", m.ID, in.Text())
		}
	}()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	go func() {
		in := bufio.NewScanner(stderr)
		for in.Scan() {
			log.Printf("Qemu machine %s: %s\n", m.ID, in.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		return err
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

		log.Printf("Qemu machine %s: process exited: %s", m.ID, errs)
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

	m.PID = cmd.Process.Pid
	return nil
}

func QemuCheck(m *Machine) {
	out, err := exec.Command("kill", "-s", "0", strconv.Itoa(m.PID)).CombinedOutput()
	if m.PID == 0 || err != nil {
		m.State = StateDown
		m.PID = 0
		return
	}

	if string(out) == "" {
		m.State = StateUp
		return
	}

	log.Println(string(out))

	m.State = StateDown
	m.PID = 0
}

func QemuStop(m *Machine) error {
	proc, err := os.FindProcess(m.PID)
	if err != nil {
		m.State = StateDown
		return nil
	}

	err = proc.Kill()
	if err != nil {
		return errors.KillFailed
	}

	m.State = StateDown
	m.PID = 0

	return nil
}

func QemuDelete(m *Machine) error {
	if len(m.NetBridgeOn) != 0 {
		return NetDeleteTAP(m.ID)
	}

	return nil
}
