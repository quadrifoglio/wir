package utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func GetCpuUsage() (int, error) {
	values := make([][]int, 2)

	for i := 0; i < 2; i++ {
		file, err := os.Open("/proc/stat")
		if err != nil {
			return 0, fmt.Errorf("failed to get cpu stats: %s", err)
		}

		r := bufio.NewReader(file)
		line, _, err := r.ReadLine()
		if err != nil {
			file.Close()
			return 0, fmt.Errorf("failed to get cpu stats: %s", err)
		}

		file.Close()

		valuesStr := strings.Split(string(line[5:]), " ")
		values[i] = make([]int, len(valuesStr))

		for j, v := range valuesStr {
			vv, err := strconv.Atoi(v)
			if err != nil {
				return 0, fmt.Errorf("failed to get cpu stats: %s", err)
			}

			values[i][j] = vv
		}

		time.Sleep(50 * time.Millisecond)
	}

	prevIdle := values[0][3] + values[0][4]
	idle := values[1][3] + values[1][4]

	prevNonIdle := values[0][0] + values[0][1] + values[0][2] + values[0][5] + values[0][6] + values[0][7]
	nonIdle := values[1][0] + values[1][1] + values[1][2] + values[1][5] + values[1][6] + values[1][7]

	prevTotal := prevIdle + prevNonIdle
	total := idle + nonIdle

	percentage := float32((total-prevTotal)-(idle-prevIdle)) / float32(total-prevTotal) * 100
	return int(percentage), nil
}

func GetRamUsage() (uint64, uint64, error) {
	var mib uint64 = 1048576

	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}

	defer f.Close()

	var c bool = true
	var total uint64
	var free uint64
	var buffers uint64
	var cached uint64

	s := bufio.NewScanner(f)
	for s.Scan() && c {
		l := s.Text()
		n := strings.Index(l, ":")
		key := l[:n]
		vstr := strings.Split(strings.Trim(l[n+1:], " "), " ")[0]

		value, err := strconv.ParseUint(vstr, 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to parse /proc/meminfo: %s")
		}

		switch key {
		case "MemTotal":
			total = value * 1000
			break
		case "MemFree":
			free = value * 1000
			break
		case "Buffers":
			buffers = value * 1000
			break
		case "Cached":
			cached = value * 1000
			c = false
			break
		}
	}

	if total == 0 || free == 0 || buffers == 0 || cached == 0 {
		return 0, 0, fmt.Errorf("failed to parse /proc/meminfo: missing values")
	}

	return (total - (free + buffers + cached)) / mib, total / mib, nil
}

func GetFreeSpace(dir string) (uint64, error) {
	var gib uint64 = 1073741824
	var s syscall.Statfs_t

	err := syscall.Statfs(dir, &s)
	if err != nil {
		return 0, fmt.Errorf("failed to get free space: %s", err)
	}

	return (s.Bavail * uint64(s.Bsize)) / gib, nil
}
