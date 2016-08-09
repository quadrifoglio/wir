package api

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/quadrifoglio/wir/net"
	"github.com/quadrifoglio/wir/shared"
)

type Machine interface {
	json.Marshaler

	Create(img shared.Image, info shared.MachineInfo) error
	Update(info shared.MachineInfo) error
	Delete() error

	Sysprep(os, hostname, root string) error

	Start() error
	Stop() error

	Info() *shared.MachineInfo
	Type() string
	State() shared.MachineState
	Stats() (shared.MachineStats, error)

	Clone(name string) error

	ListVolumes() ([]shared.Volume, error)
	CreateVolume(shared.Volume) error
	DeleteVolume(name string) error

	ListCheckpoints() ([]string, error)
	HasCheckpoint(name string) bool
	CreateCheckpoint(name string) error
	RestoreCheckpoint(name string) error
	DeleteCheckpoint(name string) error

	ListBackups() ([]shared.MachineBackup, error)
	CreateBackup() (shared.MachineBackup, error)
	RestoreBackup(name string) error
	DeleteBackup(name string) error
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

func MonitorNetwork(m Machine) {
	go func(m Machine) {
		for {
			if m.State() != shared.StateUp {
				break
			}

			for i, _ := range m.Info().Interfaces {
				a := net.MonitorInterface(MachineIfName(m, i), "rx")

				if a == net.MonitorCancel {
					break
				}
				if a == net.MonitorAlert {
					// TODO: Send email
				}
				if a == net.MonitorStop {
					// TODO: Send email

					log.Println("iface monitor %s: shuting down (to many pps)", MachineIfName(m, i))

					err := m.Stop()
					if err != nil {
						log.Println(err)
					}

					break
				}
			}

			time.Sleep(10 * time.Second)
		}
	}(m)
}

func MachineIfName(m Machine, num int) string {
	return fmt.Sprintf("veth%d.%d", m.Info().Index, num)
}
