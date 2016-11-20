package system

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

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

// EbtablesSetup runs the basic commads in order for
// ebtables traffic control to work
func EbtablesSetup() error {
	// Set default policy
	cmd := exec.Command("ebtables", "-P", "FORWARD", "DROP")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ebtables -P FORWARD DROP: %s", utils.OneLine(out))
	}

	// Accept ARP
	cmd = exec.Command("ebtables", "-A", "FORWARD", "-p", "arp", "-j", "ACCEPT")

	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ebtables -A FORWARD -p arp -j ACCEPT: %s", utils.OneLine(out))
	}

	// Accept DHCP
	cmd = exec.Command("ebtables", "-A", "FORWARD", "-p", "ip", "--ip-src", "0.0.0.0", "-j", "ACCEPT")

	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ebtables -A FORWARD -p ip --ip-src 0.0.0.0 -j ACCEPT: %s", utils.OneLine(out))
	}

	return nil
}

// EbtablesAllowTraffic creates an ebtables rule to allow
// traffic from the specified MAC/IP addresses
func EbtablesAllowTraffic(iface, mac, ip string) error {
	cmd := exec.Command("ebtables", "-A", "FORWARD", "-p", "ip", "-i", iface, "--ip-src", ip, "-s", mac, "-j", "ACCEPT")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ebtables -A FORWARD -p ip -i %s --ip-src %s -s %s -j ACCEPT: %s", iface, ip, mac, utils.OneLine(out))
	}

	cmd = exec.Command("ebtables", "-A", "FORWARD", "-p", "ip", "-o", iface, "--ip-dst", ip, "-d", mac, "-j", "ACCEPT")

	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ebtables -A FORWARD -p ip -o %s --ip-dst %s -d %s -j ACCEPT: %s", iface, ip, mac, utils.OneLine(out))
	}

	return nil
}

// EbtablesFlush deletes all the rules related
// to the `iface` interface
func EbtablesFlush(iface string) error {
	// List all the rules
	cmd := exec.Command("ebtables", "-L", "FORWARD")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ebtables -L FORWARD: %s", utils.OneLine(out))
	}

	// Find the rules concerning iface
	var rules []string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, iface) {
			rules = append(rules, strings.TrimSuffix(line, " "))
		}
	}

	// Finally, delete them
	for _, rule := range rules {
		args := make([]string, 2)
		args[0] = "-D"
		args[1] = "FORWARD"
		args = append(args, strings.Split(rule, " ")...)

		cmd := exec.Command("ebtables", args...)

		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("ebtables -D FORWARD %s: %s", rule, utils.OneLine(out))
		}
	}

	return nil
}
