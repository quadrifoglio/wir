package machine

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/milosgajdos83/tenus"
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

func NetOpenTAP(name string) (*os.File, error) {
	f, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	var r ifreq
	r.flags = IFF_TAP

	copy(r.name[:], name)

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TUNSETIFF), uintptr(unsafe.Pointer(&r)))
	if errno != 0 {
		f.Close()
		return nil, errno
	}

	return f, nil
}

func NetInitEbtables(cmds string) error {
	cmd := exec.Command(cmds, "-L", "WIR")

	err := cmd.Run()
	if err != nil {
		cmd = exec.Command(cmds, "-N", "WIR", "-P", "DROP")

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("Creating WIR chain: %s", err)
		}

		cmd = exec.Command(cmds, "-A", "FORWARD", "-p", "ip", "-i", "v+", "-j", "WIR")

		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf(string(out))
			return fmt.Errorf("Adding forwarding rule to WIR chain: %s", err)
		}
	}

	return nil
}

func NetGrantTraffic(cmds, mac, ip string) error {
	cmd := exec.Command(cmds, "-A", "WIR", "-p", "ip", "--ip-src", ip, "-s", mac, "-j", "ACCEPT")

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Granting traffic: %s", err)
	}

	return nil
}

func NetDenyTraffic(cmds, mac, ip string) error {
	cmd := exec.Command(cmds, "-D", "WIR", "-p", "ip", "--ip-src", ip, "-s", mac, "-j", "ACCEPT")

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Denying traffic: %s", err)
	}

	return nil
}

func NetTAPPersist(f *os.File, persist bool) error {
	var i int
	if persist {
		i = 1
	} else {
		i = 0
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TUNSETPERSIST), uintptr(i))
	if errno != 0 {
		return errno
	}

	return nil
}

func NetCreateBridge(name string) error {
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
