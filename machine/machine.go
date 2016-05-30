package machine

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

type Machine struct {
	ID      string
	Name    string
	Type    int
	Image   string
	State   int
	Cores   int
	Memory  int
	Network NetworkMode `json:",omitempty"`
	Qemu    BackendQemu
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
