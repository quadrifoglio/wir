package machine

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"gopkg.in/lxc/go-lxc.v2"

	"github.com/quadrifoglio/wir/config"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/utils"
)

func LxcCreate(name string, img *image.Image, cores, memory int) (Machine, error) {
	var m Machine
	m.Name = name
	m.Type = img.Type
	m.Image = img.Name
	m.Cores = cores
	m.Memory = memory
	m.State = StateDown

	path := config.API.MachinePath + "lxc"

	err := os.MkdirAll(path, 0777)
	if err != nil {
		return m, err
	}

	tarball := fmt.Sprintf("%s/%s.tar.gz", path, name)
	checkpoint := fmt.Sprintf("%s/%s.checkpoint.tar.gz", path, name)

	if _, err := os.Stat(tarball); !os.IsNotExist(err) {
		cmd := exec.Command("tar", "--numeric-owner", "-xzvf", tarball, "-C", path)

		err := cmd.Run()
		if err != nil {
			return m, fmt.Errorf("tar: %s", err)
		}
	} else if _, err := os.Stat(checkpoint); !os.IsNotExist(err) {
		// TODO: Uncompress checkpoint tarball

		err := LxcRestore(&m)
		if err != nil {
			return m, err
		}
	} else {
		c, err := lxc.NewContainer(name, path)
		if err != nil {
			return m, err
		}

		c.SetVerbosity(lxc.Verbose)

		if err := c.SetLogFile(path + "/" + m.Name + "/log.txt"); err != nil {
			return m, err
		}

		var opts lxc.TemplateOptions
		opts.Template = img.Source
		opts.Distro = img.Distro
		opts.Release = img.Release
		opts.Arch = img.Arch
		opts.FlushCache = false
		opts.DisableGPGValidation = false

		if err := c.Create(opts); err != nil {
			return m, err
		}

	}

	return m, nil
}

func LxcStart(m *Machine) error {
	path := config.API.MachinePath + "lxc"

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return err
	}

	c.SetVerbosity(lxc.Verbose)

	if m.Network.Mode == NetworkModeBridge {
		if err := c.SetConfigItem("lxc.network.hwaddr", m.Network.MAC); err != nil {
			return err
		}
		if err := c.SetConfigItem("lxc.network.type", "veth"); err != nil {
			return err
		}
		if err := c.SetConfigItem("lxc.network.flags", "up"); err != nil {
			return err
		}
		if err := c.SetConfigItem("lxc.network.link", "wir0"); err != nil {
			return err
		}
	}

	err = c.Start()
	if err != nil {
		return err
	}

	/*err = c.SetMemoryLimit(lxc.ByteSize(m.Memory) * lxc.MB)
	if err != nil {
		return err
	}*/

	return nil
}

func LxcLinuxSysprep(m *Machine, hostname, root string) error {
	path := config.API.MachinePath + "lxc"

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
	path := config.API.MachinePath + "lxc"

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

	path := config.API.MachinePath + "lxc"

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
	path := config.API.MachinePath + "lxc"

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return err
	}

	c.SetVerbosity(lxc.Verbose)

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

func LxcRestore(m *Machine) error {
	path := config.API.MachinePath + "lxc"

	c, err := lxc.NewContainer(m.Name, path)
	if err != nil {
		return err
	}

	c.SetVerbosity(lxc.Verbose)

	path = fmt.Sprintf("%s/%s/checkpoint", path, m.Name)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		err = c.Restore(lxc.RestoreOptions{path, true})
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("restore: no checkpoint")
	}

	return nil
}

func LxcStop(m *Machine) error {
	path := config.API.MachinePath + "lxc"

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
	path := config.API.MachinePath + "lxc"

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
