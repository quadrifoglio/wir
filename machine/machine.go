package machine

const (
	StateDown = 0
	StateUp   = 1
)

type Machine struct {
	ID     string
	Type   int
	Image  string
	State  int
	Cores  int
	Memory int
	PID    int
}
