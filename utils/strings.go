package utils

import (
	"bytes"
	"io"
	"strconv"
	"strings"
)

// ReadLine reads a single line (max n characters) from the
// specified reader and returns it as a string
func ReadLine(r io.Reader, n int) (string, error) {
	var buf bytes.Buffer

	for i := 0; i < n; i++ {
		c := make([]byte, 1)

		_, err := r.Read(c)
		if err != nil {
			return "", err
		}

		if c[0] == '\n' {
			break
		}

		buf.WriteString(string(c))
	}

	return buf.String(), nil
}

// UintTokens takes a line containing numeric space-separated fields
// and converts them into a uint64 array
func UintTokens(s string) ([]uint64, error) {
	valsStr := strings.Fields(s)
	vals := make([]uint64, len(valsStr))

	for i, v := range valsStr {
		vNum, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, err
		}

		vals[i] = vNum
	}

	return vals, nil
}

// OneLine transforms the input byte sequence into
// a one-line string
func OneLine(b []byte) string {
	return strings.Replace(strings.Replace(strings.TrimSpace(string(b)), "\n", ". ", -1), "\"", "'", -1)
}
