package utils

import (
	"strings"
)

func OneLine(in []byte) string {
	str := strings.TrimSpace(string(in))
	return strings.Replace(str, "\n", ". ", -1)
}
