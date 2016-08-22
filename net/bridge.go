package net

import (
	gonet "net"

	"fmt"
	"strings"

	"github.com/milosgajdos83/tenus"
)

func BridgeName(name string) string {
	if len(name) >= 12 {
		name = name[:12]
	}

	return fmt.Sprintf("wir%s%s", strings.ToUpper(name[:1]), name[1:])
}

func CreateBridge(name string) error {
	_, err := tenus.BridgeFromName(name)
	if err == nil {
		return fmt.Errorf("create bridge: already exists")
	}

	br, err := tenus.NewBridgeWithName(name)
	if err != nil {
		return fmt.Errorf("create bridge: %s", err)
	}

	if err = br.SetLinkUp(); err != nil {
		return fmt.Errorf("create bridge: %s", err)
	}

	return nil
}

func AddBridgeIf(brs, ifaces string) error {
	br, err := tenus.BridgeFromName(brs)
	if err != nil {
		return fmt.Errorf("add bridge if: %s", err)
	}

	iface, err := gonet.InterfaceByName(ifaces)
	if err != nil {
		return fmt.Errorf("add bridge if: %s", err)
	}

	if err = br.AddSlaveIfc(iface); err != nil {
		return fmt.Errorf("add bridge if: %s", err)
	}

	return nil
}

func DeleteBridge(name string) error {
	err := tenus.DeleteLink(name)
	if err != nil {
		return fmt.Errorf("delete bridge: %s", err)
	}

	return nil
}
