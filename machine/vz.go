package machine

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/quadrifoglio/wir/global"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/net"
)

func VzCreate(m *Machine, name string, i uint64, img image.Image, cores, memory int) error {
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

	cmd := exec.Command(global.APIConfig.Vzctl, args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		if out != nil {
			log.Println("OpenVZ create:", string(out))
		}

		return err
	}

	return nil
}

func VzStart(m *Machine) error {
	cmd := exec.Command(global.APIConfig.Vzctl, "start", m.Vz.CTID, "--wait")

	out, err := cmd.CombinedOutput()
	if err != nil {
		if out != nil {
			log.Println("OpenVZ:", string(out))
		}

		return err
	}

	/*args = []string{"set", m.Vz.CTID, "--ram", strconv.Itoa(m.Memory), "--cpus", strconv.Itoa(m.Cores)}
	cmd = exec.Command(global.APIConfig.Vzctl, args...)

	out, err = cmd.CombinedOutput()
	if err != nil {
		if out != nil {
			log.Println("OpenVZ set ram and cpu: ", string(out))
		}

		return m, err
	}*/

	if m.Network.Mode == NetworkModeBridge {
		args := []string{"set", m.Vz.CTID, "--netif", "add", "eth0", "--save"}
		cmd = exec.Command(global.APIConfig.Vzctl, args...)

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

func VzCheck(m *Machine) {
	cmd := exec.Command(global.APIConfig.Vzctl, "status", m.Vz.CTID)

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

func VzStop(m *Machine) error {
	cmd := exec.Command(global.APIConfig.Vzctl, "stop", m.Vz.CTID)

	out, err := cmd.CombinedOutput()
	if err != nil {
		if out != nil {
			log.Println("OpenVZ:", string(out))
		}

		return err
	}

	return nil
}

func VzDelete(m *Machine) error {
	cmd := exec.Command(global.APIConfig.Vzctl, "destroy", m.Vz.CTID)

	out, err := cmd.CombinedOutput()
	if err != nil {
		if out != nil {
			log.Println("OpenVZ:", string(out))
		}

		return err
	}

	return nil
}
