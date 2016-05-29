package client

import (
	"github.com/quadrifoglio/wir/machine"
)

type MachineRequest struct {
	Image   string
	Cores   int
	Memory  int
	Network machine.NetworkMode
}
