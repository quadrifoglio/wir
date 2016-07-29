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
	MonitorOK     = 0
	MonitorCancel = 1
	MonitorAlert  = 2
	MonitorStop   = 3
)

func MonitorInterface(iface, direction string) int {
	var err error
	var tx1, tx2 uint64

	if !InterfaceExists(iface) {
		return MonitorCancel
	}

	if tx1, err = ifaceStat(iface, direction); err != nil {
		log.Println(err)
		return MonitorCancel
	}

	time.Sleep(1 * time.Second)

	if tx2, err = ifaceStat(iface, direction); err != nil {
		log.Println(err)
		return MonitorCancel
	}

	txPPS := tx2 - tx1

	if txPPS >= global.APIConfig.PPSStop {
		return MonitorStop
	} else if txPPS >= global.APIConfig.PPSAlert {
		return MonitorAlert
	}

	return MonitorOK
}

func ifaceStat(iface, stat string) (uint64, error) {
	d, err := ioutil.ReadFile(fmt.Sprintf("/sys/class/net/%s/statistics/%s_packets", iface, stat))
	if err != nil {
		return 0, fmt.Errorf("iface monitor %s (%s): %s\n", iface, stat, err)
	}

	i, err := strconv.Atoi(string(d[:len(d)-1]))
	if err != nil {
		return 0, fmt.Errorf("iface monitor %s (%s): %s\n", iface, stat, err)
	}

	return uint64(i), nil
}
