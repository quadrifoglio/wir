package api

import (
	"encoding/json"
	"strconv"

	"github.com/quadrifoglio/wir/net"
	"github.com/quadrifoglio/wir/shared"
)

type Machine interface {
	json.Marshaler

	Create(img Image, info shared.MachineInfo) error
	Update(info shared.MachineInfo) error
	Delete() error

	Sysprep(os, hostname, root string) error

	Start() error
	Stop() error

	Info() *shared.MachineInfo
	Type() string
	State() shared.MachineState
	Stats() (shared.MachineStats, error)

	HasCheckpoint() bool
	CreateCheckpoint() error
	RestoreCheckpoint() error
	DeleteCheckpoint() error
}

type Machines []Machine

func (ms Machines) Len() int {
	return len(ms)
}

func (ms Machines) Less(i, j int) bool {
	return ms[i].Info().Index < ms[j].Info().Index
}

func (ms Machines) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}

func SetupMachineNetwork(m Machine, network shared.MachineNetwork) error {
	mi := m.Info()
	mi.Network = network

	var err error
	if len(mi.Network.Mode) > 0 {
		if len(mi.Network.MAC) == 0 {
			mi.Network.MAC, err = net.GenerateMAC(shared.APIConfig.NodeID)
			if err != nil {
				return err
			}
		}

		err = net.GrantBasic(shared.APIConfig.Ebtables, mi.Network.MAC)
		if err != nil {
			return err
		}

		if len(mi.Network.Mode) > 0 && len(mi.Network.IP) > 0 {
			err := net.GrantTraffic(shared.APIConfig.Ebtables, mi.Network.MAC, mi.Network.IP)
			if err != nil {
				return err
			}
		}
	}

	return err
}

func UpdateMachineNetwork(m Machine, network shared.MachineNetwork) error {
	mi := m.Info()

	if len(network.Mode) > 0 && network.Mode != mi.Network.Mode {
		mi.Network.Mode = network.Mode
	}

	if len(network.MAC) > 0 && network.MAC != mi.Network.MAC {
		net.DenyTraffic(shared.APIConfig.Ebtables, mi.Network.MAC, mi.Network.IP) // Not handling errors: can fail if no ip was previously registered

		mi.Network.MAC = network.MAC
		err := net.GrantTraffic(shared.APIConfig.Ebtables, mi.Network.MAC, mi.Network.IP)
		if err != nil {
			return err
		}
	}

	if len(network.IP) > 0 && network.IP != mi.Network.IP {
		net.DenyTraffic(shared.APIConfig.Ebtables, mi.Network.MAC, mi.Network.IP)

		mi.Network.IP = network.IP

		err := net.GrantTraffic(shared.APIConfig.Ebtables, mi.Network.MAC, mi.Network.IP)
		if err != nil {
			return err
		}
	}

	return nil
}

func MachineIfName(m Machine) string {
	return "veth" + strconv.Itoa(int(m.Info().Index)) + ".0"
}
