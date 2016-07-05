package net

import (
	"fmt"
	"os/exec"
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

func GrantTraffic(cmds, mac, ip string) error {
	cmd := exec.Command(cmds, "-A", "WIR", "-p", "ip", "--ip-src", ip, "-s", mac, "-j", "ACCEPT")

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Granting traffic: %s", err)
	}

	return nil
}

func DenyTraffic(cmds, mac, ip string) error {
	cmd := exec.Command(cmds, "-D", "WIR", "-p", "ip", "--ip-src", ip, "-s", mac, "-j", "ACCEPT")

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Denying traffic: %s", err)
	}

	return nil
}
