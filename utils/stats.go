package utils

import (
	"bufio"
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
			return 0, err
		}

		r := bufio.NewReader(file)
		line, _, err := r.ReadLine()
		if err != nil {
			file.Close()
			return 0, err
		}

		file.Close()

		valuesStr := strings.Split(string(line[5:]), " ")
		values[i] = make([]int, len(valuesStr))

		for j, v := range valuesStr {
			vv, err := strconv.Atoi(v)
			if err != nil {
				return 0, err
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
	s := syscall.Sysinfo_t{}

	err := syscall.Sysinfo(&s)
	if err != nil {
		return 0, 0, err
	}

	return (s.Totalram - s.Freeram), s.Totalram, nil
}
