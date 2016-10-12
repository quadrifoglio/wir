package utils

import (
	"encoding/binary"
	"encoding/hex"
	"math/rand"
)

func RandID(prefix byte) string {
	buf := make([]byte, 4)
	r := rand.Uint32()

	binary.PutUvarint(buf, uint64(r))
	buf[0] = prefix

	return hex.EncodeToString(buf)
}
