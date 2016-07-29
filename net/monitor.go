package net

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/quadrifoglio/wir/global"
)

const (
	MonitorOK    = 0
	MonitorAlert = 1
	MonitorStop  = 2
)

func MonitorInterface(iface string) int {
	var tx1, tx2 uint64

	if !InterfaceExists(iface) {
		return MonitorStop
	}

	if tx1 = ifaceStat(iface, "tx"); tx1 == 0 {
		return MonitorOK
	}

	time.Sleep(1 * time.Second)

	if tx2 = ifaceStat(iface, "tx"); tx2 == 0 {
		return MonitorOK
	}

	txPPS := tx2 - tx1

	if txPPS >= global.APIConfig.PPSStop {
		return MonitorStop
	} else if txPPS >= global.APIConfig.PPSAlert {
		return MonitorAlert
	}

	return MonitorOK
}

func ifaceStat(iface, stat string) uint64 {
	d, err := ioutil.ReadFile(fmt.Sprintf("/sys/class/net/%s/statistics/%s_packets"))
	if err != nil {
		log.Println("iface monitor %s (%s): %s", iface, stat, err)
		return 0
	}

	i, err := strconv.Atoi(string(d))
	if err != nil {
		log.Println("iface monitor %s (%s): %s", iface, stat, err)
		return 0
	}

	return uint64(i)
}
