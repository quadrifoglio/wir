package inter

import (
	"fmt"
	"path/filepath"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/global"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/machine"
	"github.com/quadrifoglio/wir/utils"
)

func MigrateImage(i image.Image, src, dst global.Remote) error {
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

func MigrateQemu(m machine.Machine, i image.Image, src, dst global.Remote) error {
	m.Check()

	if m.State == machine.StateUp {
		err := machine.QemuStop(&m)
		if err != nil {
			return fmt.Errorf("failed to stop local machine: %s", err)
		}
	}

	err := MigrateImage(i, src, dst)
	if err != nil {
		return err
	}

	conf, err := client.GetConfig(dst)
	if err != nil {
		return fmt.Errorf("failed to get remote config: %s", err)
	}

	srcPath := fmt.Sprintf("%s/qemu/%s.qcow2", global.APIConfig.MachinePath, m.Name)
	dstPath := fmt.Sprintf("%s/qemu/%s.qcow2", conf.MachinePath, m.Name)

	err = utils.MakeRemoteDirectories(dst, filepath.Dir(dstPath))
	if err != nil {
		return fmt.Errorf("falied to make remote dirs: %s", err)
	}

	err = utils.SCP(srcPath, dst, dstPath)
	if err != nil {
		return fmt.Errorf("failed to scp to remote: %s", err)
	}

	r := client.MachineRequest{m.Name, i.Name, m.Cores, m.Memory, m.Network}

	_, err = client.CreateMachine(dst, r)
	if err != nil {
		return fmt.Errorf("failed to create remote machine: %s", err)
	}

	return nil
}

func LiveMigrateQemu(m machine.Machine, i image.Image, src, dst global.Remote) error {
	err := machine.QemuCheckpoint(&m)
	if err != nil {
		return fmt.Errorf("failed to checkpoint: %s", err)
	}

	err = MigrateQemu(m, i, src, dst)
	if err != nil {
		return err
	}

	err = client.StartMachine(dst, m.Name)
	if err != nil {
		return fmt.Errorf("failed to start remote machine: %s", err)
	}

	return nil
}

func MigrateLxc(m machine.Machine, i image.Image, src, dst global.Remote) error {
	m.Check()

	if m.State == machine.StateUp {
		err := machine.LxcStop(&m)
		if err != nil {
			return fmt.Errorf("failed to stop local machine: %s", err)
		}
	}

	err := MigrateImage(i, src, dst)
	if err != nil {
		return err
	}

	conf, err := client.GetConfig(dst)
	if err != nil {
		return fmt.Errorf("failed to get remote config: %s", err)
	}

	srcFolder := fmt.Sprintf("%s/lxc/%s", global.APIConfig.MachinePath, m.Name)

	err = utils.TarDirectory(srcFolder, fmt.Sprintf("/tmp/%s.tar.gz", m.Name))
	if err != nil {
		return err
	}

	srcPath := fmt.Sprintf("/tmp/%s.tar.gz", m.Name)
	dstPath := fmt.Sprintf("%s/lxc/%s.tar.gz", conf.MachinePath, m.Name)

	err = utils.MakeRemoteDirectories(dst, filepath.Dir(dstPath))
	if err != nil {
		return fmt.Errorf("falied to make remote dirs: %s", err)
	}

	err = utils.SCP(srcPath, dst, dstPath)
	if err != nil {
		return fmt.Errorf("failed to scp to remote: %s", err)
	}

	r := client.MachineRequest{m.Name, i.Name, m.Cores, m.Memory, m.Network}

	_, err = client.CreateMachine(dst, r)
	if err != nil {
		return fmt.Errorf("failed to create remote machine: %s", err)
	}

	err = machine.LxcStop(&m)
	if err != nil {
		return fmt.Errorf("failed to stop local machine: %s", err)
	}

	return nil
}

func LiveMigrateLxc(m machine.Machine, i image.Image, src, dst global.Remote) error {
	err := machine.LxcCheckpoint(&m)
	if err != nil {
		return fmt.Errorf("failed to checkpoint: %s", err)
	}

	err = MigrateLxc(m, i, src, dst)
	if err != nil {
		return err
	}

	err = client.StartMachine(dst, m.Name)
	if err != nil {
		return fmt.Errorf("failed to start remote macine: %s", err)
	}

	return nil
}
