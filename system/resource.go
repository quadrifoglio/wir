package system

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

/*
#include <unistd.h>
#include <sys/types.h>
#include <pwd.h>
#include <stdlib.h>
*/
import "C"

// Uptime returns the amount of time that the system
// has been running for in seconds
func Uptime() (float64, error) {
	data, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}

	l := strings.Fields(string(data))
	if len(l) < 1 {
		return 0, fmt.Errorf("invalid /proc/uptime file")
	}

	uptime, err := strconv.ParseFloat(l[0], 64)
	if err != nil {
		return 0, err
	}

	return uptime, nil
}

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

// GetProcessRamUsage returns the current number or megabytes
// used by the specified process
func ProcessRamUsage(pid int) (uint64, error) {
	data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/statm", pid))
	if err != nil {
		return 0, err
	}

	l := strings.Fields(string(data))
	if len(l) < 1 {
		return 0, fmt.Errorf("invalid /proc/%d/statm", pid)
	}

	// Resident memory size in number of pages
	ram, err := strconv.ParseUint(l[1], 10, 64)
	if err != nil {
		return 0, err
	}

	return ram * 4096, nil

}

// GetProcessCpuUsage calculates the current CPU usage
// in percent of the specified process
func ProcessCpuUsage(pid int) (float32, error) {
	data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, fmt.Errorf("get process cpu usage: %s", err)
	}

	l := strings.Fields(string(data))
	if len(l) < 22 {
		return 0, fmt.Errorf("get process cpu usage: invalid /proc/%d/stat file", pid)
	}

	hz := float32(TicksPerSecond())

	uptime, err := Uptime()
	if err != nil {
		return 0, err
	}

	utime, err := strconv.ParseUint(l[13], 10, 64)
	if err != nil {
		return 0, err
	}

	stime, err := strconv.ParseUint(l[14], 10, 64)
	if err != nil {
		return 0, err
	}

	startTime, err := strconv.ParseUint(l[21], 10, 64)
	if err != nil {
		return 0, err
	}

	totalTime := float32(utime + stime)
	seconds := float32(uptime) - float32(startTime)/hz

	return 100 * ((totalTime / hz) / seconds), nil
}

// TicksPerSecond returns the number of CPU clock
// ticks per second
func TicksPerSecond() uint64 {
	return uint64(C.sysconf(C._SC_CLK_TCK))
}
