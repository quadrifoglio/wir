package machine

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/net"
)

func VzCreate(vzCmd, basePath, name string, i uint64, img *image.Image, cores, memory int) (Machine, error) {
	var m Machine
	m.Name = name
	m.Type = img.Type
	m.Image = img.Name
	m.Cores = cores
	m.Memory = memory
	m.Vz.CTID = strconv.Itoa(100 + int(i))
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
			log.Println("OpenVZ create:", string(out))
		}

		return m, err
	}

	return m, nil
}

func VzStart(vzCmd string, m *Machine) error {
	cmd := exec.Command(vzCmd, "start", m.Vz.CTID, "--wait")

	out, err := cmd.CombinedOutput()
	if err != nil {
		if out != nil {
			log.Println("OpenVZ:", string(out))
		}

		return err
	}

	/*args = []string{"set", m.Vz.CTID, "--ram", strconv.Itoa(m.Memory), "--cpus", strconv.Itoa(m.Cores)}
	cmd = exec.Command(vzCmd, args...)

	out, err = cmd.CombinedOutput()
	if err != nil {
		if out != nil {
			log.Println("OpenVZ set ram and cpu: ", string(out))
		}

		return m, err
	}*/

	if m.Network.Mode == NetworkModeBridge {
		args := []string{"set", m.Vz.CTID, "--netif", "add", "eth0", "--save"}
		cmd = exec.Command(vzCmd, args...)

		out, err = cmd.CombinedOutput()
		if err != nil {
			if out != nil {
				log.Println("OpenVZ:", string(out))
			}

			return err
		}

		err = net.BridgeAddIf("wir0", fmt.Sprintf("veth%s.0", m.Vz.CTID))
		if err != nil {
			return err
		}

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
