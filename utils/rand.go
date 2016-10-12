package utils

import (
	"encoding/binary"
	"encoding/hex"
	"math/rand"
)

// RandID generates a random 32 bits identifier
// in which the first byte will have the value of 'prefix'
func RandID(prefix byte) string {
	buf := make([]byte, 8)
	r := rand.Uint32()

	binary.PutUvarint(buf, uint64(r))
	buf[0] = prefix

	return hex.EncodeToString(buf[:4])
}
