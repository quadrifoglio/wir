package utils

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"

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
	rand.Seed(time.Now().Unix())

	f := rand.Uint32()
	buf := make([]byte, 4)

	binary.LittleEndian.PutUint32(buf, f)

	str := fmt.Sprintf("52:54:%02x:%02x:%02x:%02x", prefix, buf[0], buf[1], buf[2])
	return str, nil
}
