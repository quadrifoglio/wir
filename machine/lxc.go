package machine

import (
	"os"
	"path/filepath"

	"github.com/quadrifoglio/wir/image"
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

	c, err := lxc.NewContainer(name, path)
	if err != nil {
		return m, err
	}

	c.SetVerbosity(lxc.Verbose)

	if err := c.SetLogFile(path + "/" + m.Name + "/log.txt"); err != nil {
		return m, err
	}

	// TODO: Uncomment
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

	if len(m.Network.BridgeOn) > 0 {
		err := NetCreateBridge("wir0")
		if err != nil {
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

	err = c.SetMemoryLimit(lxc.ByteSize(m.Memory) * lxc.MB)
	if err != nil {
		return err
	}

	err = c.Start()
	if err != nil {
		return err
	}

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
