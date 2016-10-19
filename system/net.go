package system

import (
	"fmt"
	"net"
	"os/exec"

	"github.com/quadrifoglio/wir/utils"
)

// InterfaceExists returns true if the
// specified interface exists on the system
func InterfaceExists(name string) bool {
	_, err := net.InterfaceByName(name)
	if err != nil {
		// Do net log the error: assuming the interface just does not exist
		return false
	}

	return true
}

// CreateBridge creates a new interface
// working as Linux bridge (switch equivalent)
func CreateBridge(name string) error {
	// Add the bridge with iproute2 (brctl is deprecated)
	cmd := exec.Command("ip", "link", "add", name, "type", "bridge")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ip link add: %s", utils.OneLine(out))
	}

	// Set up the bridge interface
	cmd = exec.Command("ip", "link", "set", "up", "dev", name)

	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ip link set up: %s", utils.OneLine(out))
	}

	return nil
}

// SetInterfaceMaster adds the specified interface
// to the master bridge
func SetInterfaceMaster(name, master string) error {
	cmd := exec.Command("ip", "link", "set", "dev", name, "master", master)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ip link set master: %s", utils.OneLine(out))
	}

	return nil
}

// DeleteInterface deletes the specified
// network inteface
func DeleteInterface(name string) error {
	cmd := exec.Command("ip", "link", "del", name)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ip link add: %s", utils.OneLine(out))
	}

	return nil
}
