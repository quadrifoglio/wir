package utils

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
)

// RandID generates a 32 bits identifier
// in which the first byte will have the value of 'prefix'
func RandID(prefix byte) string {
	b := make([]byte, 9)
	nano := time.Now().UnixNano()
	binary.LittleEndian.PutUint64(b[1:], uint64(nano))

	b[0] = prefix
	return hex.EncodeToString(b)
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
