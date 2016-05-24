package machine

import (
	"fmt"
	"net"
	"os/exec"

	"github.com/milosgajdos83/tenus"
)

func NetCreateBridge(name string) error {
	br, err := tenus.BridgeFromName(name)
	if err != nil {
		br, err = tenus.NewBridgeWithName(name)
		if err != nil {
			return fmt.Errorf("Create bridge: %s", err)
		}
	}

	if err = br.SetLinkUp(); err != nil {
		return fmt.Errorf("Create bridge: %s", err)
	}

	return nil
}

func NetCreateTAP(name string) error {
	cmd := exec.Command("ip", "tuntap", "add", "dev", name, "mode", "tap")

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Create TAP: %s", err)
	}

	return nil
}

func NetDeleteTAP(name string) error {
	cmd := exec.Command("ip", "link", "del", "dev", name)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Delete TAP: %s", err)
	}

	return nil
}

func NetBridgeAddIf(brs, ifaces string) error {
	br, err := tenus.BridgeFromName(brs)
	if err != nil {
		return fmt.Errorf("Addif: %s", err)
	}

	iface, err := net.InterfaceByName(ifaces)
	if err != nil {
		return fmt.Errorf("Addif: %s", err)
	}

	if err = br.AddSlaveIfc(iface); err != nil {
		return fmt.Errorf("Addif: %s", err)
	}

	return nil
}
