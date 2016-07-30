package api

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/lxc/go-lxc.v2"

	"github.com/quadrifoglio/wir/net"
	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/utils"
)

type LxcMachine struct {
	shared.MachineInfo
}

func (m *LxcMachine) Info() *shared.MachineInfo {
	return &m.MachineInfo
}

func (m *LxcMachine) Type() string {
	return shared.TypeLXC
}

func (m *LxcMachine) Create(img Image, info shared.MachineInfo) error {
	path := fmt.Sprintf("%s/lxc", shared.APIConfig.MachinePath)

	m.Name = info.Name
	m.Image = img.Info().Name
	m.Cores = info.Cores
	m.Memory = info.Memory

	err := os.MkdirAll(path, 0777)
	if err != nil {
		return err
	}

	var c *lxc.Container

	tar := fmt.Sprintf("%s/%s.tar.gz", path, m.Name)
	if _, err := os.Stat(tar); os.IsNotExist(err) {
		c, err = lxc.NewContainer(m.Name, path)
		if err != nil {
			return err
		}

		if err := c.SetLogFile(fmt.Sprintf("%s/%s/log.txt", path, m.Name)); err != nil {
			return err
		}

		c.SetVerbosity(lxc.Verbose)

		var opts lxc.TemplateOptions
		opts.Template = img.Info().Source
		/*opts.Distro = img.Distro
		opts.Release = img.Release
		opts.Arch = img.Arch*/
		opts.FlushCache = false
		opts.DisableGPGValidation = false

		if err = c.Create(opts); err != nil {
			return err
		}
	} else {
		err = utils.UntarDirectory(tar, path)
		if err != nil {
			return fmt.Errorf("failed to create container from archive: %s", err)
		}

		c, err = lxc.NewContainer(m.Name, path)
		if err != nil {
			return err
		}

		if err := c.SetLogFile(fmt.Sprintf("%s/%s/log.txt", path, m.Name)); err != nil {
			return err
		}

		c.SetVerbosity(lxc.Verbose)

		err = os.Remove(tar)
		if err != nil {
			return fmt.Errorf("failed to remove container archive: %s", err)
		}

		if err := c.SetConfigItem("lxc.rootfs", fmt.Sprintf("%s/%s/rootfs", path, m.Name)); err != nil {
			return fmt.Errorf("failed to change rootfs config: %s", err)
		}
	}

	if err := c.SetConfigItem("lxc.console", "none"); err != nil {
		return fmt.Errorf("failed to change lxc.console config: %s", err)
	}
	if err := c.SetConfigItem("lxc.tty", "0"); err != nil {
		return fmt.Errorf("failed to change lxc.tty config: %s", err)
	}
	if err := c.SetConfigItem("lxc.cgroup.devices.deny", "c 5:1 rwm"); err != nil {
		return fmt.Errorf("failed to change lxc.cgroup.devices.deny config: %s", err)
	}

	if err := c.SaveConfigFile(fmt.Sprintf("%s/%s/config", path, m.Name)); err != nil {
		return fmt.Errorf("failed to save config: %s", err)
	}

	return SetupMachineNetwork(m, info.Network)
}

func (m *LxcMachine) Update(info shared.MachineInfo) error {
	if info.Cores != 0 && info.Cores != m.Cores {
		m.Cores = info.Cores
	}

	if info.Memory != 0 && info.Cores != m.Cores {
		m.Memory = info.Cores
	}

	return UpdateMachineNetwork(m, info.Network)
}

func (m *LxcMachine) Delete() error {
	path := fmt.Sprintf("%s/lxc", shared.APIConfig.MachinePath)

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return err
	}

	c.SetVerbosity(lxc.Verbose)

	err = c.Destroy()
	if err != nil {
		return err
	}

	return nil
}

func (m *LxcMachine) Sysprep(os, hostname, root string) error {
	path := fmt.Sprintf("%s/lxc", shared.APIConfig.MachinePath)

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return err
	}

	if err := c.SetConfigItem("lxc.utsname", hostname); err != nil {
		return err
	}

	err = utils.ChangeRootPassword(fmt.Sprintf("%s/%s/rootfs/etc/shadow", path, m.Name), root)
	if err != nil {
		return err
	}

	return nil
}

func (m *LxcMachine) Start() error {
	path := fmt.Sprintf("%s/lxc", shared.APIConfig.MachinePath)

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return err
	}

	c.SetVerbosity(lxc.Verbose)

	if m.Network.Mode == shared.NetworkModeBridge {
		if err := c.SetConfigItem("lxc.network.type", "veth"); err != nil {
			return err
		}
		if err := c.SetConfigItem("lxc.network.veth.pair", MachineIfName(m)); err != nil {
			return err
		}
		if err := c.SetConfigItem("lxc.network.flags", "up"); err != nil {
			return err
		}
		if err := c.SetConfigItem("lxc.network.link", "wir0"); err != nil {
			return err
		}
		if err := c.SetConfigItem("lxc.network.hwaddr", m.Network.MAC); err != nil {
			return err
		}

		err = CheckMachineNetwork(m)
		if err != nil {
			return err
		}
	}

	chk := fmt.Sprintf("%s/%s/checkpoint", path, m.Name)
	if _, err := os.Stat(chk); err == nil {
		err = c.Restore(lxc.RestoreOptions{chk, true})
		if err != nil {
			return fmt.Errorf("failed to restore: %s", err)
		}
	} else {
		err = c.Start()
		if err != nil {
			return err
		}
	}

	if shared.APIConfig.EnableNetMonitor && m.Network.Mode != shared.NetworkModeNone {
		go func(m *LxcMachine) {
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

	/*err = c.SetMemoryLimit(lxc.ByteSize(m.Memory) * lxc.MB)
	if err != nil {
		return err
	}*/

	return nil
}

func (m *LxcMachine) Stop() error {
	path := fmt.Sprintf("%s/lxc", shared.APIConfig.MachinePath)

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return err
	}

	c.SetVerbosity(lxc.Verbose)

	err = c.Stop()
	if err != nil {
		return err
	}

	return nil
}

func (m *LxcMachine) State() shared.MachineState {
	path := fmt.Sprintf("%s/lxc", shared.APIConfig.MachinePath)

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		log.Println("state: %s", err)
	}

	s := c.State()
	if s == lxc.RUNNING {
		return shared.StateUp
	}

	return shared.StateDown
}

func (m *LxcMachine) Stats() (shared.MachineStats, error) {
	var stats shared.MachineStats

	path := fmt.Sprintf("%s/lxc", shared.APIConfig.MachinePath)

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return stats, err
	}

	c1, err := c.CPUStats()
	if err != nil {
		return stats, err
	}

	c1time := c1["user"] + c1["system"]

	s1, err := utils.GetCpuUsage()
	if err != nil {
		return stats, err
	}

	time.Sleep(50 * time.Millisecond)

	c2, err := c.CPUStats()
	if err != nil {
		return stats, err
	}

	c2time := c2["user"] + c2["system"]

	s2, err := utils.GetCpuUsage()
	if err != nil {
		return stats, err
	}

	stats.CPU = (float32(c2time-c1time) / float32(s2.Total-s1.Total)) * 100

	mem, err := c.MemoryUsage()
	if err != nil {
		return stats, err
	}

	stats.RAMUsed = uint64(mem * lxc.MB)
	stats.RAMFree = uint64(m.Memory) - stats.RAMUsed

	return stats, nil
}

func (m *LxcMachine) HasCheckpoint() bool {
	path := fmt.Sprintf("%s/lxc/%s/checkpoint", shared.APIConfig.MachinePath, m.Name)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func (m *LxcMachine) CreateCheckpoint() error {
	path := fmt.Sprintf("%s/lxc", shared.APIConfig.MachinePath)

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return err
	}

	c.SetVerbosity(lxc.Verbose)

	err = c.Freeze()
	if err != nil {
		return err
	}

	path = fmt.Sprintf("%s/%s/checkpoint", path, m.Name)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0777)
		if err != nil {
			return err
		}
	} else {
		err := utils.ClearFolder(path)
		if err != nil {
			return err
		}
	}

	err = c.Checkpoint(lxc.CheckpointOptions{path, false, true})
	if err != nil {
		return err
	}

	return nil
}

func (m *LxcMachine) RestoreCheckpoint() error {
	path := fmt.Sprintf("%s/lxc", shared.APIConfig.MachinePath)

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return err
	}

	c.SetVerbosity(lxc.Verbose)

	if !m.HasCheckpoint() {
		return fmt.Errorf("checkpoint does not exists")
	}

	err = c.Restore(lxc.RestoreOptions{path, true})
	if err != nil {
		return err
	}

	return nil
}

func (m *LxcMachine) DeleteCheckpoint() error {
	path := fmt.Sprintf("%s/lxc/%s/checkpoint", shared.APIConfig.MachinePath, m.Name)
	return os.RemoveAll(path)
}

func (m *LxcMachine) MarshalJSON() ([]byte, error) {
	type mdr struct {
		LxcMachine

		Type  string
		State shared.MachineState
	}

	return json.Marshal(mdr{*m, m.Type(), m.State()})
}
