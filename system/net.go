package system

import (
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

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

// DownInterface shuts down the
// specified network interface
func DownInterface(name string) error {
	cmd := exec.Command("ip", "link", "set", "down", "dev", name)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ip link set down: %s", utils.OneLine(out))
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

// EbtablesClear removes the rules created by this program
// so they don't accumulate at each startup
func EbtablesClear() {
	cmd := exec.Command("ebtables", "-D", "FORWARD", "-p", "arp", "-j", "ACCEPT")
	cmd.Run()

	cmd = exec.Command("ebtables", "-D", "FORWARD", "-p", "ip", "--ip-src", "0.0.0.0", "-j", "ACCEPT")
	cmd.Run()
}

// EbtablesSetup runs the basic commads in order for
// ebtables traffic control to work
func EbtablesSetup() error {
	EbtablesClear()

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

// GetInterfacePPS returns the number of packets transmitted
// by second in the specified direction for a given interface
func GetInterfacePPS(iface, direction string) (uint64, error) {
	var err error
	var tx1, tx2 uint64

	if !InterfaceExists(iface) {
		return 0, fmt.Errorf("Interface does not exist")
	}

	if tx1, err = GetInterfaceStat(iface, fmt.Sprintf("%s_packets", direction)); err != nil {
		return 0, err
	}

	time.Sleep(1 * time.Second)

	if tx2, err = GetInterfaceStat(iface, fmt.Sprintf("%s_packets", direction)); err != nil {
		return 0, err
	}

	return tx2 - tx1, nil
}

// GetInterfaceStat returns the required stats for
// the specified interface from /sys
func GetInterfaceStat(iface, stat string) (uint64, error) {
	d, err := ioutil.ReadFile(fmt.Sprintf("/sys/class/net/%s/statistics/%s", iface, stat))
	if err != nil {
		return 0, fmt.Errorf("iface monitor %s (%s): %s\n", iface, stat, err)
	}

	i, err := strconv.Atoi(string(d[:len(d)-1]))
	if err != nil {
		return 0, fmt.Errorf("iface monitor %s (%s): %s\n", iface, stat, err)
	}

	return uint64(i), nil
}
