package machine

import (
	"strconv"
)

const (
	StateDown = 0
	StateUp   = 1
)

type NetworkMode struct {
	BridgeOn string `json:",omitempty"`
}

type BackendQemu struct {
	PID int
}

type BackendVz struct {
	CTID string
}

type Machine struct {
	Name    string
	Index   uint64
	Type    int
	Image   string
	State   int
	Cores   int
	Memory  int
	Network NetworkMode `json:",omitempty"`

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
