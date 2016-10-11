package system

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// MemoryUsage returns respectively the currently used memory and the
// total memory available on the system
// It does so by parsing /proc/meminfo
func MemoryUsage() (uint64, uint64, error) {
	var total uint64
	var free uint64
	var buffers uint64
	var cached uint64

	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}

	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		fields := strings.Fields(s.Text())

		if len(fields) < 2 {
			return 0, 0, fmt.Errorf("Invalid /proc/meminfo file")
		}

		if fields[0] == "MemTotal:" {
			v, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return 0, 0, err
			}

			total = v
		}
		if fields[0] == "MemFree:" {
			v, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return 0, 0, err
			}

			free = v
		}
		if fields[0] == "Buffers:" {
			v, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return 0, 0, err
			}

			buffers = v
		}
		if fields[0] == "Cached:" {
			v, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return 0, 0, err
			}

			cached = v
		}
	}

	if err := s.Err(); err != nil {
		return 0, 0, err
	}

	return total - (free + cached + buffers), total, nil
}
