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
	Type    int
	Image   string
	State   int
	Cores   int
	Memory  int
	Network NetworkMode `json:",omitempty"`
	Qemu    BackendQemu
}
