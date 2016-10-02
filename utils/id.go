package utils

import (
	"encoding/binary"
	"encoding/hex"
	"time"
)

func UniqueID(nodeId byte) string {
	b := make([]byte, 9)
	nano := time.Now().UnixNano()
	binary.LittleEndian.PutUint64(b[1:], uint64(nano))

	b[0] = nodeId
	return hex.EncodeToString(b)
}
