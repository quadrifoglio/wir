// +build cgo
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

/*
#include <unistd.h>
*/
import "C"

type CpuUsage struct {
	Idle    int
	NonIdle int
	Total   int
}

func GetClockTicks() int {
	return int(C.sysconf(C._SC_CLK_TCK))
}

func GetCpuUsage() (CpuUsage, error) {
	var usage CpuUsage

	file, err := os.Open("/proc/stat")
	if err != nil {
		return usage, fmt.Errorf("get cpu usage: %s", err)
	}

	defer file.Close()

	r := bufio.NewReader(file)
	line, _, err := r.ReadLine()
	if err != nil {
		return usage, fmt.Errorf("get cpu usage: %s", err)
	}

	valuesStr := strings.Split(string(line[5:]), " ")
	values := make([]int, len(valuesStr))

	for i, v := range valuesStr {
		vv, err := strconv.Atoi(v)
		if err != nil {
			return usage, fmt.Errorf("get cpu usage: %s", err)
		}

		values[i] = vv
		usage.Total += vv
	}

	usage.Idle = values[3] + values[4]
	usage.NonIdle = values[0] + values[1] + values[2] + values[5] + values[6] + values[7]

	return usage, nil
}

func GetCpuUsagePercentage() (float32, error) {
	u1, err := GetCpuUsage()
	if err != nil {
		return 0, fmt.Errorf("get cpu usage percentage: %s", err)
	}

	time.Sleep(50 * time.Millisecond)

	u2, err := GetCpuUsage()
	if err != nil {
		return 0, fmt.Errorf("get cpu usage percentage: %s", err)
	}

	return float32((u2.Total-u1.Total)-(u2.Idle-u1.Idle)) / float32(u2.Total-u1.Total) * 100, nil
}

func GetProcessCpuStats(pid int) (int, int, error) {
	file, err := os.Open(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, 0, fmt.Errorf("get process cpu stats: %s", err)
	}

	defer file.Close()

	r := bufio.NewReader(file)
	line, _, err := r.ReadLine()
	if err != nil {
		return 0, 0, fmt.Errorf("get process cpu stats: %s", err)
	}

	values := strings.Split(string(line), " ")

	utime, err := strconv.Atoi(values[13])
	if err != nil {
		return 0, 0, fmt.Errorf("get process cpu stats: %s", err)
	}

	stime, err := strconv.Atoi(values[14])
	if err != nil {
		return 0, 0, fmt.Errorf("get process cpu stats: %s", err)
	}

	return utime, stime, nil
}

func GetRamUsage() (uint64, uint64, error) {
	var mib uint64 = 1048576

	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, fmt.Errorf("get ram usage: %s", err)
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
			return 0, 0, fmt.Errorf("get ram usage: /proc/meminfo: %s", err)
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
		return 0, 0, fmt.Errorf("get ram usage: /proc/meminfo: missing values")
	}

	return (total - (free + buffers + cached)) / mib, total / mib, nil
}

func GetProcessRamUsage(pid int) (uint64, error) {
	var result uint64

	f, err := os.Open(fmt.Sprintf("/proc/%d/smaps", pid))
	if err != nil {
		return 0, fmt.Errorf("get process ram usage: %s", err)
	}

	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		l := s.Text()

		if n := strings.Index(l, "Size:"); n != -1 {
			str := strings.Split(strings.Trim(l[n+6:], " "), " ")[0]

			mem, err := strconv.Atoi(str)
			if err != nil {
				return 0, fmt.Errorf("get process ram usage: %s", err)
			}

			result = uint64(mem / 1000)
			break
		}
	}

	return result, nil
}

func GetFreeSpace(dir string) (uint64, error) {
	var gib uint64 = 1073741824
	var s syscall.Statfs_t

	err := syscall.Statfs(dir, &s)
	if err != nil {
		return 0, fmt.Errorf("get free space: %s", err)
	}

	return (s.Bavail * uint64(s.Bsize)) / gib, nil
}
