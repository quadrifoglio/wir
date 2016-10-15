package server

import (
	"fmt"
	"strings"

	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/system"
)

// NetworkIface returns the bridge interface name
// coresponding to the specified network ID
func NetworkIface(id string) string {
	return fmt.Sprintf("wir%s%s", strings.ToUpper(id[:1]), id[1:])
}

// MachineIface returns the interface name coresponding
// to the n-th interface of the specified machine ID
func MachineIface(id string, n int) string {
	return fmt.Sprintf("wir%s%s.%d", strings.ToUpper(id[:1]), id[1:], n)
}

// CreateNetwork creates the specified network
// by using Linux bridges
func CreateNetwork(def shared.NetworkDef) error {
	err := system.CreateBridge(NetworkIface(def.ID))
	if err != nil {
		return err
	}

	if len(def.GatewayIface) > 0 {
		err := system.SetInterfaceMaster(def.GatewayIface, NetworkIface(def.ID))
		if err != nil {
			return err
		}
	}

	return nil
}

// AttachInterfaceToNetwork attaches the n-th interface of the specified machine ID
// to the specified network ID
func AttachInterfaceToNetwork(machineId string, n int, netId string) error {
	return system.SetInterfaceMaster(MachineIface(machineId, n), NetworkIface(netId))
}

// DeleteNetwork deletes the specified network
// and the associated bridge
func DeleteNetwork(id string) error {
	return system.DeleteInterface(NetworkIface(id))
}
