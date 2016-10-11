package system

import (
	"os"
	"time"

	"github.com/quadrifoglio/wir/utils"
)

// CpuUsage returns the current CPU usage in percentage
// calculated using the values in /proc/stat
func CpuUsage() (float32, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return -1, err
	}

	defer f.Close()

	line1, err := utils.ReadLine(f, 1024)
	if err != nil {
		return -1, err
	}

	vals1, err := utils.UintTokens(line1[5:]) // /proc/stat first line starts with 'cpu ', so remove it
	if err != nil {
		return -1, err
	}

	time.Sleep(200 * time.Millisecond)

	_, err = f.Seek(0, 0) // Go back to the begining of the file to read the same line
	if err != nil {
		return -1, err
	}

	line2, err := utils.ReadLine(f, 1024)
	if err != nil {
		return -1, err
	}

	vals2, err := utils.UintTokens(line2[5:]) // /proc/stat first line starts with 'cpu ', so remove it
	if err != nil {
		return -1, err
	}

	idle1 := vals1[3] + vals1[4]
	idle2 := vals2[3] + vals2[4]

	nonIdle1 := vals1[0] + vals1[1] + vals1[2] + vals1[5] + vals1[6] + vals1[7]
	nonIdle2 := vals2[0] + vals2[1] + vals2[2] + vals2[5] + vals2[6] + vals2[7]

	total1 := idle1 + nonIdle1
	total2 := idle2 + nonIdle2

	return float32((total2-total1)-(idle2-idle1)) / float32(total2-total1) * 100, nil
}
