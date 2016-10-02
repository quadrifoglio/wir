package net

import (
	"crypto/rand"
	"fmt"
	"net"
	"regexp"
	"strconv"
)

func InterfaceExists(iface string) bool {
	_, err := net.InterfaceByName(iface)
	if err != nil {
		return false
	}

	return true
}

func GenerateMAC(nodeId byte) (string, error) {
	buf := make([]byte, 3)

	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}

	str := fmt.Sprintf("52:54:%02x:%02x:%02x:%02x", nodeId, buf[0], buf[1], buf[2])
	return str, nil
}

func ParseMask(mask string) net.IPMask {
	r := regexp.MustCompile("^([0-9]{1,3})\\.([0-9]{1,3})\\.([0-9]{1,3})\\.([0-9]{1,3})$")

	valsArr := r.FindAllStringSubmatch(mask, -1)
	if len(valsArr) < 1 {
		return nil
	}

	vals := valsArr[0]
	if len(vals) != 5 {
		return nil
	}

	bs := make([]byte, 4)
	for i := 1; i <= 4; i++ {
		b, err := strconv.ParseUint(vals[i], 10, 8)
		if err != nil {
			return nil
		}

		bs[i-1] = byte(b)
	}

	return net.IPMask(bs)
}
