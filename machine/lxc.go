package machine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/utils"
	"gopkg.in/lxc/go-lxc.v2"
)

func LxcCreate(basePath, name string, img *image.Image, cores, memory int) (Machine, error) {
	var m Machine
	m.Name = name
	m.Type = img.Type
	m.Image = img.Name
	m.Cores = cores
	m.Memory = memory
	m.State = StateDown

	path, err := filepath.Abs(basePath + "lxc/")
	if err != nil {
		return m, err
	}

	err = os.MkdirAll(path, 0777)
	if err != nil {
		return m, err
	}

	tarball := fmt.Sprintf("%s/%s.tar.gz", path, name)
	if _, err := os.Stat(tarball); os.IsNotExist(err) {
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
	} else {
		cmd := exec.Command("tar", "--numeric-owner", "-xzvf", tarball, "-C", path)

		err := cmd.Run()
		if err != nil {
			return m, fmt.Errorf("tar: %s", err)
		}
	}

	return m, nil
}

func LxcStart(basePath string, m *Machine) error {
	path, err := filepath.Abs(basePath + "lxc/")
	if err != nil {
		return err
	}

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

func LxcCheck(basePath string, m *Machine) error {
	path, err := filepath.Abs(basePath + "lxc/")
	if err != nil {
		return err
	}

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

func LxcStats(basePath string, m *Machine) (Stats, error) {
	var stats Stats

	path, err := filepath.Abs(basePath + "lxc/")
	if err != nil {
		return stats, err
	}

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

func LxcStop(basePath string, m *Machine) error {
	path, err := filepath.Abs(basePath + "lxc/")
	if err != nil {
		return err
	}

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

func LxcDelete(basePath string, m *Machine) error {
	path, err := filepath.Abs(basePath + "lxc/")
	if err != nil {
		return err
	}

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
