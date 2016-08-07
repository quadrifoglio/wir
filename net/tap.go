package net

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func OpenTAP(name string) (*os.File, error) {
	f, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open /dev/net/tun: %s", err)
	}

	var r ifreq
	r.flags = IFF_TAP

	copy(r.name[:], name)

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TUNSETIFF), uintptr(unsafe.Pointer(&r)))
	if errno != 0 {
		f.Close()
		return nil, fmt.Errorf("open tap: %s", errno)
	}

	return f, nil
}

func TAPPersist(f *os.File, persist bool) error {
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
