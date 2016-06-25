package inter

import (
	"fmt"
	"path/filepath"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/machine"
)

func MigrateQemu(basePath string, m machine.Machine, i image.Image, src, dst client.Remote) error {
	_, err := client.GetImage(dst, m.Image)
	if err != nil {
		r := client.ImageRequest{
			i.Name,
			i.Type,
			fmt.Sprintf("scp://%s/%s", src.Addr, i.Source),
			i.Arch,
			i.Distro,
			i.Release,
		}

		_, err = client.CreateImage(dst, r)
		if err != nil {
			return fmt.Errorf("Remote image: %s", err)
		}
	}

	conf, err := client.GetConfig(dst)
	if err != nil {
		return fmt.Errorf("Remote configuration: %s", err)
	}

	srcPath := basePath + "qemu/" + m.Name + ".qcow2"
	dstPath := conf.MachinePath + "qemu/" + m.Name + ".qcow2"

	err = RemoteMkdir(dst, filepath.Dir(dstPath))
	if err != nil {
		return fmt.Errorf("MkdirRemote: %s", err)
	}

	err = SCP(srcPath, dst, dstPath)
	if err != nil {
		return fmt.Errorf("SCP to remote: %s", err)
	}

	r := client.MachineRequest{m.Name, i.Name, m.Cores, m.Memory, m.Network}

	_, err = client.CreateMachine(dst, r)
	if err != nil {
		return fmt.Errorf("Remote machine: %s", err)
	}

	return nil
}
