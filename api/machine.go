package api

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

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

	ListBackups() ([]string, error)
	CreateBackup(name string) error
	RestoreBackup(name string) error
	DeleteBackup(name string) error

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

		err = net.GrantTraffic(mi.Network.MAC, "0.0.0.0")
		if err != nil {
			return err
		}

		if len(mi.Network.IP) > 0 {
			err := net.GrantTraffic(mi.Network.MAC, mi.Network.IP)
			if err != nil {
				return err
			}
		}
	}

	return err
}

func CheckMachineNetwork(m Machine) error {
	mi := m.Info()
	netw := mi.Network

	if netw.Mode == shared.NetworkModeNone || len(netw.MAC) == 0 {
		return nil
	}

	is, err := net.IsGranted(netw.MAC, "0.0.0.0")
	if err != nil {
		return err
	}

	if !is {
		err := net.GrantTraffic(netw.MAC, "0.0.0.0")
		if err != nil {
			return err
		}
	}

	if len(netw.IP) > 0 {
		is, err = net.IsGranted(netw.MAC, netw.IP)
		if err != nil {
			return err
		}

		if !is {
			err := net.GrantTraffic(netw.MAC, netw.IP)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func UpdateMachineNetwork(m Machine, network shared.MachineNetwork) error {
	mi := m.Info()

	if len(network.Mode) > 0 && network.Mode != mi.Network.Mode {
		mi.Network.Mode = network.Mode
	}

	if len(network.MAC) > 0 && network.MAC != mi.Network.MAC {
		net.DenyTraffic(mi.Network.MAC, mi.Network.IP) // Not handling errors: can fail if no ip was previously registered

		mi.Network.MAC = network.MAC
		err := net.GrantTraffic(mi.Network.MAC, mi.Network.IP)
		if err != nil {
			return err
		}
	}

	if len(network.IP) > 0 && network.IP != mi.Network.IP {
		net.DenyTraffic(mi.Network.MAC, mi.Network.IP)

		mi.Network.IP = network.IP

		err := net.GrantTraffic(mi.Network.MAC, mi.Network.IP)
		if err != nil {
			return err
		}
	}

	return nil
}

func MonitorNetwork(m Machine) {
	go func(m Machine) {
		for {
			a := net.MonitorInterface(MachineIfName(m), "rx")

			if m.State() != shared.StateUp {
				break
			}

			if a == net.MonitorCancel {
				break
			}
			if a == net.MonitorAlert {
				// TODO: Send email
			}
			if a == net.MonitorStop {
				// TODO: Send email

				log.Println("iface monitor %s: shuting down (to many pps)", MachineIfName(m))

				err := m.Stop()
				if err != nil {
					log.Println(err)
				}

				break
			}

			time.Sleep(10 * time.Second)
		}
	}(m)
}

func MachineIfName(m Machine) string {
	return "veth" + strconv.Itoa(int(m.Info().Index)) + ".0"
}
