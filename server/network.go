package server

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/utils"
)

// NetworkIface returns the interface name
// coresponding to the specified network ID
func NetworkIface(id string) string {
	return fmt.Sprintf("wir%s%s", strings.ToUpper(id[:1]), id[1:])
}

// CreateNetwork creates the specified network
// by using Linux bridges
func CreateNetwork(def shared.NetworkDef) error {
	cmd := exec.Command("ip", "link", "add", NetworkIface(def.ID), "type", "bridge")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ip link add: %s", utils.OneLine(out))
	}

	if len(def.GatewayIface) > 0 {
		cmd := exec.Command("ip", "link", "dev", def.GatewayIface, "master", NetworkIface(def.ID))

		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("ip link set master: %s", utils.OneLine(out))
		}
	}

	return nil
}

// DeleteNetwork deletes the specified network
// and the associated bridge
func DeleteNetwork(id string) error {
	cmd := exec.Command("ip", "link", "del", id)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ip link add: %s", utils.OneLine(out))
	}

	return nil
}
