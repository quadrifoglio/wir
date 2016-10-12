package utils

import (
	"encoding/binary"
	"encoding/hex"
	"math/rand"
)

func RandID() string {
	buf := make([]byte, 4)
	r := rand.Uint32()

	binary.PutUvarint(buf, uint64(r))

	return hex.EncodeToString(buf)
}
