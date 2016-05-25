package machine

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"unsafe"

	"github.com/milosgajdos83/tenus"
)

const (
	IFF_TUN   = 0x0001
	IFF_TAP   = 0x0002
	IFF_NO_PI = 0x1000
)

type ifreq struct {
	name  [0x10]byte
	flags uint16
	osef  [0x16]byte
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
		return fmt.Errorf("Create bridge: %s", err)
	}

	return nil
}

func NetCreateTAP(name string) error {
	f, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return err
	}

	defer f.Close()

	var r ifreq
	r.flags = IFF_TAP

	copy(r.name[:], name[:14])

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TUNSETIFF), uintptr(unsafe.Pointer(&r)))
	if errno != 0 {
		return errno
	}

	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TUNSETPERSIST), 1)
	if errno != 0 {
		return errno
	}

	return nil
}

func NetDeleteTAP(name string) error {
	f, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return err
	}

	defer f.Close()

	var r ifreq
	r.flags = IFF_TAP

	copy(r.name[:], name[:14])

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TUNSETIFF), uintptr(unsafe.Pointer(&r)))
	if errno != 0 {
		return errno
	}

	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TUNSETPERSIST), 0)
	if errno != 0 {
		return errno
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
