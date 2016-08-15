package api

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/quadrifoglio/wir/net"
	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/utils"
)

type Machine interface {
	// Used for sending a representaion of the machine
	// To the clinets
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

	ListInterfaces() []shared.NetworkDevice
	CreateInterface(iface shared.NetworkDevice) (shared.NetworkDevice, error)
	UpdateInterface(index int, iface shared.NetworkDevice) (shared.NetworkDevice, error)
	DeleteInterface(index int) error

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

			var alerted bool

			for i, _ := range m.ListInterfaces() {
				a := net.MonitorInterface(MachineIfName(m, i), "rx")

				if a == net.MonitorCancel {
					break
				}
				if a == net.MonitorAlert {
					msg := fmt.Sprintf("Machine %s has reached the 'alert' level number of pps.", m.Info().Name)

					err := utils.SendAlertMail(msg)
					if err != nil {
						log.Printf("iface monitor %s: %s\n", MachineIfName(m, i), err)
					}

					alerted = true
				}
				if a == net.MonitorStop {
					log.Printf("iface monitor %s: shuting down (to many pps)\n", MachineIfName(m, i))

					err := m.Stop()
					if err != nil {
						log.Println(err)
					}

					msg := fmt.Sprintf("Machine %s has reached the 'stop' level number of pps. Shutting down.", m.Info().Name)
					err = utils.SendAlertMail(msg)
					if err != nil {
						log.Printf("iface monitor %s: %s\n", MachineIfName(m, i), err)
					}

					break
				}
			}

			if alerted {
				time.Sleep(15 * time.Minute)
				alerted = false
			} else {
				time.Sleep(10 * time.Second)
			}
		}
	}(m)
}

func MachineIfName(m Machine, num int) string {
	return fmt.Sprintf("veth%d.%d", m.Info().Index, num)
}
