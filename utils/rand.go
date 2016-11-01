package utils

import (
	"fmt"
	"math/rand"

	"github.com/rs/xid"
)

// RandID generates a unique 12 bytes identifier and returns
// it as a 20 chars string (Base32 hex encoded)
func RandID() string {
	return xid.New().String()
}

// RandMAC generates a random MAC address with
// the third byte as prefix
func RandMAC(prefix byte) (string, error) {
	buf := make([]byte, 3)

	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}

	str := fmt.Sprintf("52:54:%02x:%02x:%02x:%02x", prefix, buf[0], buf[1], buf[2])
	return str, nil
}
