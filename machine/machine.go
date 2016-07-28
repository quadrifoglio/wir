package machine

import (
	"strconv"

	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/global"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/net"
)

const (
	StateDown = 0
	StateUp   = 1

	NetworkModeNone   = ""
	NetworkModeBridge = "bridge"
)

type NetworkSetup struct {
	Mode string
	MAC  string
	IP   string
}

type BackendQemu struct {
	PID int
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

func Create(freeIndex uint64, name string, img image.Image, cores, memory int, network NetworkSetup) (Machine, error) {
	var m Machine
	var err error

	switch img.Type {
	case image.TypeQemu:
		err = QemuCreate(&m, name, img, cores, memory)
		break
	case image.TypeLXC:
		err = LxcCreate(&m, name, img, cores, memory)
		break
	default:
		err = errors.InvalidImageType
		break
	}

	m.Network = network

	if len(m.Network.Mode) > 0 {
		if len(m.Network.MAC) == 0 {
			m.Network.MAC, err = net.GenerateMAC(global.APIConfig.NodeID)
			if err != nil {
				return m, err
			}
		}

		err = net.GrantBasic(global.APIConfig.Ebtables, m.Network.MAC)
		if err != nil {
			return m, err
		}

		if len(m.Network.Mode) > 0 && len(m.Network.IP) > 0 {
			err := net.GrantTraffic(global.APIConfig.Ebtables, m.Network.MAC, m.Network.IP)
			if err != nil {
				return m, err
			}
		}
	}

	return m, err
}

func (m *Machine) Update(cores, memory int, network NetworkSetup) error {
	if cores != 0 {
		m.Cores = cores
	}

	if memory != 0 {
		m.Memory = memory
	}

	if len(network.Mode) > 0 && network.Mode != m.Network.Mode {
		m.Network.Mode = network.Mode
	}

	if len(network.MAC) > 0 && network.MAC != m.Network.MAC {
		net.DenyTraffic(global.APIConfig.Ebtables, m.Network.MAC, m.Network.IP) // Not handling errors: can fail if no ip was previously registered

		m.Network.MAC = network.MAC
		err := net.GrantTraffic(global.APIConfig.Ebtables, m.Network.MAC, m.Network.IP)
		if err != nil {
			return err
		}
	}

	if len(network.IP) > 0 && network.IP != m.Network.IP {
		net.DenyTraffic(global.APIConfig.Ebtables, m.Network.MAC, m.Network.IP)

		m.Network.IP = network.IP

		err := net.GrantTraffic(global.APIConfig.Ebtables, m.Network.MAC, m.Network.IP)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Machine) Check() {
	switch m.Type {
	case image.TypeQemu:
		QemuCheck(m)
		break
	case image.TypeLXC:
		LxcCheck(m)
		break
	}
}

func (m *Machine) Start() error {
	var err error

	switch m.Type {
	case image.TypeQemu:
		err = QemuStart(m)
		break
	case image.TypeLXC:
		err = LxcStart(m)
		break
	default:
		err = errors.InvalidImageType
		break
	}

	if err != nil {
		return err
	}

	return nil
}

func (m *Machine) Stop() error {
	var err error

	m.Check()

	if m.State != StateUp {
		return errors.InvalidMachineState
	}

	switch m.Type {
	case image.TypeQemu:
		err = QemuStop(m)
		break
	case image.TypeLXC:
		err = LxcStop(m)
		break
	default:
		err = errors.InvalidImageType
		break
	}

	if err != nil {
		return err
	}

	return nil
}

func (m *Machine) Stats() (Stats, error) {
	var err error
	var stats Stats

	switch m.Type {
	case image.TypeQemu:
		stats, err = QemuStats(m)
		break
	case image.TypeLXC:
		stats, err = LxcStats(m)
		break
	default:
		err = errors.InvalidImageType
		break
	}

	if err != nil {
		return stats, err
	}

	return stats, nil
}

func (m *Machine) LinuxSysprep(mainPart int, hostname, rootPasswd string) error {
	var err error

	switch m.Type {
	case image.TypeQemu:
		err = QemuLinuxSysprep(m, mainPart, hostname, rootPasswd)
		break
	case image.TypeLXC:
		err = LxcLinuxSysprep(m, hostname, rootPasswd)
		break
	}

	if err != nil {
		return err
	}

	return nil
}

func (m *Machine) Delete() error {
	var err error

	m.Check()
	if m.State != StateDown {
		return errors.InvalidMachineState
	}

	switch m.Type {
	case image.TypeQemu:
		err = QemuDelete(m)
		break
	case image.TypeLXC:
		err = LxcDelete(m)
		break
	}

	if err != nil {
		return err
	}

	return nil
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
