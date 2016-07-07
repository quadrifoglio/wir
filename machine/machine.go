package machine

import (
	"strconv"
)

const (
	StateDown = 0
	StateUp   = 1

	NetworkModeNone   = ""
	NetworkModeBridge = "bridge"
)

type NetworkSetup struct {
	Mode string
}

type BackendQemu struct {
	PID int
}

type BackendVz struct {
	CTID string
}

type Stats struct {
	CPU     float32
	RAMUsed uint64
	RAMFree uint64
}

type Machine struct {
	Name    string
	Index   uint64
	Type    string
	Image   string
	State   int
	Cores   int
	Memory  int
	Network NetworkSetup `json:",omitempty"`

	Qemu BackendQemu
	Vz   BackendVz
}

type Machines []Machine

func (ms Machines) Len() int {
	return len(ms)
}

func (ms Machines) Less(i, j int) bool {
	return ms[i].Index < ms[j].Index
}

func (ms Machines) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}

func (m Machine) IfName() string {
	return "veth" + strconv.Itoa(int(m.Index)) + ".0"
}

func StateToString(s int) string {
	switch s {
	case StateDown:
		return "down"
	case StateUp:
		return "up"
	default:
		return "unknown"
	}
}

func StringToState(s string) int {
	switch s {
	case "down":
		return StateDown
	case "up":
		return StateUp
	default:
		return StateDown
	}
}
