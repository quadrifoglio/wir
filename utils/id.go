package utils

import (
	"encoding/binary"
	"encoding/hex"
	"time"
)

func UniqueID() string {
	b := make([]byte, 8)
	nano := time.Now().UnixNano()
	binary.LittleEndian.PutUint64(b, uint64(nano))

	return hex.EncodeToString(b)
}
