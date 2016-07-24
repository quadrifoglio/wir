package inter

import (
	"fmt"
	"path/filepath"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/config"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/machine"
)

func MigrateImage(i image.Image, src, dst client.Remote) error {
	_, err := client.GetImage(dst, i.Name)
	if err == nil {
		return nil
	}

	r := client.ImageRequest{
		i.Name,
		i.Type,
		fmt.Sprintf("scp://%s/%s", src.Addr, i.Source),
		i.MainPartition,
		i.Arch,
		i.Distro,
		i.Release,
	}

	_, err = client.CreateImage(dst, r)
	if err != nil {
		return fmt.Errorf("failed to create remote image: %s", err)
	}

	return nil
}

func MigrateQemu(m machine.Machine, i image.Image, src, dst client.Remote) error {
	err := MigrateImage(i, src, dst)
	if err != nil {
		return err
	}

	conf, err := client.GetConfig(dst)
	if err != nil {
		return fmt.Errorf("failed to get remote config: %s", err)
	}

	srcPath := fmt.Sprintf("%s/qemu/%s.qcow2", config.API.MachinePath, m.Name)
	dstPath := fmt.Sprintf("%s/qemu/%s.qcow2", conf.MachinePath, m.Name)

	err = RemoteMkdir(dst, filepath.Dir(dstPath))
	if err != nil {
		return fmt.Errorf("falied to make remote dirs: %s", err)
	}

	err = SCP(srcPath, dst, dstPath)
	if err != nil {
		return fmt.Errorf("failed to scp to remote: %s", err)
	}

	r := client.MachineRequest{m.Name, i.Name, m.Cores, m.Memory, m.Network}

	_, err = client.CreateMachine(dst, r)
	if err != nil {
		return fmt.Errorf("failed to create remote machine: %s", err)
	}

	err = machine.QemuDelete(&m)
	if err != nil {
		return fmt.Errorf("failed to delete local machine: %s", err)
	}

	return nil
}

func LiveMigrateQemu(m machine.Machine, i image.Image, src, dst client.Remote) error {
	err := machine.QemuCheckpoint(&m)
	if err != nil {
		return fmt.Errorf("failed to snapshot: %s", err)
	}

	return MigrateQemu(m, i, src, dst)
}
