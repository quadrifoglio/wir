package net

import (
	"crypto/rand"
	"fmt"
	"net"
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