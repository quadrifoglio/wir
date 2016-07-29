package machine

import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/lxc/go-lxc.v2"

	"github.com/quadrifoglio/wir/global"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/net"
	"github.com/quadrifoglio/wir/utils"
)

func LxcCreate(m *Machine, name string, img image.Image, cores, memory int) error {
	m.Name = name
	m.Type = img.Type
	m.Image = img.Name
	m.Cores = cores
	m.Memory = memory
	m.State = StateDown

	path := fmt.Sprintf("%s/lxc", global.APIConfig.MachinePath)

	err := os.MkdirAll(path, 0777)
	if err != nil {
		return err
	}

	var c *lxc.Container

	tar := fmt.Sprintf("%s/%s.tar.gz", path, name)
	if _, err := os.Stat(tar); os.IsNotExist(err) {
		c, err = lxc.NewContainer(name, path)
		if err != nil {
			return err
		}

		if err := c.SetLogFile(fmt.Sprintf("%s/%s/log.txt", path, m.Name)); err != nil {
			return err
		}

		c.SetVerbosity(lxc.Verbose)

		var opts lxc.TemplateOptions
		opts.Template = img.Source
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

		c, err = lxc.NewContainer(name, path)
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

		if err := c.SetConfigItem("lxc.rootfs", fmt.Sprintf("%s/%s/rootfs", path, name)); err != nil {
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

	if err := c.SaveConfigFile(fmt.Sprintf("%s/%s/config", path, name)); err != nil {
		return fmt.Errorf("failed to save config: %s", err)
	}

	return nil
}

func LxcStart(m *Machine) error {
	path := fmt.Sprintf("%s/lxc", global.APIConfig.MachinePath)

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return err
	}

	c.SetVerbosity(lxc.Verbose)

	if m.Network.Mode == NetworkModeBridge {
		if err := c.SetConfigItem("lxc.network.type", "veth"); err != nil {
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

		if global.APIConfig.EnableNetMonitor {
			go func(m *Machine) {
				a := net.MonitorInterface(m.IfName())

				m.Check()

				if m.State != StateUp {
					return
				}

				if a == net.MonitorStop {
					return
				}
				if a == net.MonitorAlert {
					// TODO: Send email
				}
				if a == net.MonitorStop {
					// TODO: Send email

					err := LxcStop(m)
					if err != nil {
						log.Println(err)
					}

					return
				}

				time.Sleep(60 * time.Second)
			}(m)
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

	/*err = c.SetMemoryLimit(lxc.ByteSize(m.Memory) * lxc.MB)
	if err != nil {
		return err
	}*/

	return nil
}

func LxcLinuxSysprep(m *Machine, hostname, root string) error {
	path := fmt.Sprintf("%s/lxc", global.APIConfig.MachinePath)

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return err
	}

	c.SetVerbosity(lxc.Verbose)

	if err := c.SetConfigItem("lxc.utsname", hostname); err != nil {
		return err
	}

	err = utils.ChangeRootPassword(fmt.Sprintf("%s/%s/rootfs/etc/shadow", path, m.Name), root)
	if err != nil {
		return err
	}

	return nil
}

func LxcCheck(m *Machine) error {
	path := fmt.Sprintf("%s/lxc", global.APIConfig.MachinePath)

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return err
	}

	c.SetVerbosity(lxc.Verbose)

	s := c.State()
	if s == lxc.RUNNING {
		m.State = StateUp
	} else {
		m.State = StateDown
	}

	return nil
}

func LxcStats(m *Machine) (Stats, error) {
	var stats Stats

	path := fmt.Sprintf("%s/lxc", global.APIConfig.MachinePath)

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

func LxcCheckpoint(m *Machine) error {
	path := fmt.Sprintf("%s/lxc", global.APIConfig.MachinePath)

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

func LxcStop(m *Machine) error {
	path := fmt.Sprintf("%s/lxc", global.APIConfig.MachinePath)

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

func LxcDelete(m *Machine) error {
	path := fmt.Sprintf("%s/lxc", global.APIConfig.MachinePath)

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
