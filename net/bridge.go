package net

import (
	gonet "net"

	"fmt"
	"github.com/milosgajdos83/tenus"
)

func CreateBridge(name string) error {
	br, err := tenus.BridgeFromName(name)
	if err != nil {
		br, err = tenus.NewBridgeWithName(name)
		if err != nil {
			return fmt.Errorf("Create bridge: %s", err)
		}
	}

	if err = br.SetLinkUp(); err != nil {
		return fmt.Errorf("Bridge up: %s", err)
	}

	return nil
}

func BridgeAddIf(brs, ifaces string) error {
	br, err := tenus.BridgeFromName(brs)
	if err != nil {
		return fmt.Errorf("Addif: %s", err)
	}

	iface, err := gonet.InterfaceByName(ifaces)
	if err != nil {
		return fmt.Errorf("Addif: %s", err)
	}

	if err = br.AddSlaveIfc(iface); err != nil {
		return fmt.Errorf("Addif: %s", err)
	}

	return nil
}
