package net

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"github.com/quadrifoglio/wir/shared"
)

const (
	IFF_TUN = 0x0001
	IFF_TAP = 0x0002
)

type ifreq struct {
	name  [0x10]byte
	flags uint16
	osef  [0x16]byte
}

func InitEbtables(ebtables string) error {
	cmd := exec.Command(ebtables, "-L", "WIR")

	err := cmd.Run()
	if err != nil {
		cmd = exec.Command(ebtables, "-N", "WIR", "-P", "DROP")

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("ebtables: creating WIR chain: %s", err)
		}

		cmd = exec.Command(ebtables, "-A", "FORWARD", "-p", "ip", "-i", "v+", "-j", "WIR")

		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf(string(out))
			return fmt.Errorf("ebtables: adding forwarding rule to WIR chain: %s", err)
		}
	}

	return nil
}

func IsGranted(mac, ip string) (bool, error) {
	cmd := exec.Command(shared.APIConfig.Ebtables, "-L", "WIR")

	out, err := cmd.StdoutPipe()
	if err != nil {
		return false, fmt.Errorf("can not open ebtables stdout: %s", err)
	}

	err = cmd.Start()
	if err != nil {
		return false, fmt.Errorf("granting traffic (0.0.0.0): %s", err)
	}

	defer cmd.Wait()

	sc := bufio.NewScanner(out)

	for sc.Scan() {
		s := sc.Text()
		mac = strings.Replace(mac, ":0", ":", -1)

		if strings.Contains(s, mac) && strings.Contains(s, ip) && strings.Contains(s, "ACCEPT") {
			return true, nil
		}
	}

	if err := sc.Err(); err != nil {
		return false, fmt.Errorf("failed to read ebtables output: %s", err)
	}

	return false, nil
}

func GrantTraffic(mac, ip string) error {
	cmd := exec.Command(shared.APIConfig.Ebtables, "-A", "WIR", "-p", "ip", "--ip-src", ip, "-s", mac, "-j", "ACCEPT")

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("granting traffic (%s / %s): %s", mac, ip, err)
	}

	return nil
}

func DenyTraffic(mac, ip string) error {
	cmd := exec.Command(shared.APIConfig.Ebtables, "-D", "WIR", "-p", "ip", "--ip-src", ip, "-s", mac, "-j", "ACCEPT")

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("denying traffic (%s / %s): %s", mac, ip, err)
	}

	return nil
}
