package shared

const (
	StateDown = 0
	StateUp   = 1

	NetworkModeNone   = ""
	NetworkModeBridge = "bridge"
)

type MachineNetwork struct {
	Mode string
	MAC  string
	IP   string
}

type MachineInfo struct {
	Index   uint64
	Name    string
	Image   string
	Cores   int
	Memory  int
	Disk    uint64
	Network MachineNetwork `json:",omitempty"`
}

type MachineStats struct {
	CPU     float32
	RAMUsed uint64
	RAMFree uint64
}

type MachineState int

func StateToString(s MachineState) string {
	switch s {
	case StateDown:
		return "down"
	case StateUp:
		return "up"
	default:
		return "unknown"
	}
}

func StringToState(s string) MachineState {
	switch s {
	case "down":
		return StateDown
	case "up":
		return StateUp
	default:
		return StateDown
	}
}
