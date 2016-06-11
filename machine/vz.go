package machine

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/quadrifoglio/wir/image"
)

func VzCreate(vzCmd, basePath, name string, img *image.Image, cores, memory int) (Machine, error) {
	var m Machine
	m.Name = name
	m.Type = img.Type
	m.Image = img.Name
	m.Cores = cores
	m.Memory = memory
	m.Vz.CTID = "101"
	m.State = StateDown

	args := make([]string, 6)
	args[0] = "create"
	args[1] = m.Vz.CTID
	args[2] = "--ostemplate"
	args[3] = m.Image
	args[4] = "--name"
	args[5] = name

	cmd := exec.Command(vzCmd, args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		if out != nil {
			log.Println("OpenVZ:", string(out))
		}

		return m, err
	}

	args[0] = "set"
	args[1] = m.Vz.CTID
	args[2] = "--ram"
	args[3] = strconv.Itoa(memory)
	args[4] = "--cpus"
	args[5] = strconv.Itoa(cores)

	cmd = exec.Command(vzCmd, args...)

	out, err = cmd.CombinedOutput()
	if err != nil {
		if out != nil {
			log.Println("OpenVZ:", string(out))
		}

		return m, err
	}

	if len(m.Network.BridgeOn) > 0 {
		err := NetCreateBridge("wir0")
		if err != nil {
			return m, err
		}

		args := []string{"set", m.Vz.CTID, "--netif", "add", "eth0", "--save"}
		cmd = exec.Command(vzCmd, args...)

		out, err = cmd.CombinedOutput()
		if err != nil {
			if out != nil {
				log.Println("OpenVZ:", string(out))
			}

			return m, err
		}

		err = NetBridgeAddIf("wir0", fmt.Sprintf("veth%s.0", m.Vz.CTID))
		if err != nil {
			return m, err
		}

	}

	return m, nil
}

func VzStart(vzCmd string, m *Machine) error {
	cmd := exec.Command(vzCmd, "start", m.Vz.CTID)

	out, err := cmd.CombinedOutput()
	if err != nil {
		if out != nil {
			log.Println("OpenVZ:", string(out))
		}

		return err
	}

	return nil
}

func VzCheck(vzCmd string, m *Machine) {
	cmd := exec.Command(vzCmd, "status", m.Vz.CTID)

	out, err := cmd.CombinedOutput()
	if err != nil {
		m.State = StateDown
		return
	}

	if strings.Contains(string(out), "running") {
		m.State = StateUp
	} else {
		m.State = StateDown
	}
}

func VzStop(vzCmd string, m *Machine) error {
	cmd := exec.Command(vzCmd, "stop", m.Vz.CTID)

	out, err := cmd.CombinedOutput()
	if err != nil {
		if out != nil {
			log.Println("OpenVZ:", string(out))
		}

		return err
	}

	return nil
}

func VzDelete(vzCmd string, m *Machine) error {
	cmd := exec.Command(vzCmd, "destroy", m.Vz.CTID)

	out, err := cmd.CombinedOutput()
	if err != nil {
		if out != nil {
			log.Println("OpenVZ:", string(out))
		}

		return err
	}

	return nil
}