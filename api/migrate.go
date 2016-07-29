package api

import (
	"fmt"
	"path/filepath"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/utils"
)

func MigrateImage(i Image, src, dst shared.Remote) error {
	im := i.Info()

	_, err := client.GetImage(dst, im.Name)
	if err == nil {
		return nil
	}

	r := shared.ImageInfo{
		im.Name,
		im.Type,
		fmt.Sprintf("scp://%s/%s", src.Addr, im.Source),
		im.Arch,
		im.Distro,
		im.Release,
		im.MainPartition,
	}

	_, err = client.CreateImage(dst, r)
	if err != nil {
		return fmt.Errorf("failed to create remote image: %s", err)
	}

	return nil
}

func MigrateQemu(m Machine, i Image, src, dst shared.Remote) error {
	mi := m.Info()

	if m.State() == shared.StateUp {
		err := m.Stop()
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

	srcPath := fmt.Sprintf("%s/qemu/%s.qcow2", shared.APIConfig.MachinePath, mi.Name)
	dstPath := fmt.Sprintf("%s/qemu/%s.qcow2", conf.MachinePath, mi.Name)

	err = utils.MakeRemoteDirectories(dst, filepath.Dir(dstPath))
	if err != nil {
		return fmt.Errorf("falied to make remote dirs: %s", err)
	}

	err = utils.SCP(srcPath, dst, dstPath)
	if err != nil {
		return fmt.Errorf("failed to scp to remote: %s", err)
	}

	_, err = client.CreateMachine(dst, *mi)
	if err != nil {
		return fmt.Errorf("failed to create remote machine: %s", err)
	}

	return nil
}

func LiveMigrateQemu(m Machine, i Image, src, dst shared.Remote) error {
	err := m.CreateCheckpoint()
	if err != nil {
		return fmt.Errorf("failed to checkpoint: %s", err)
	}

	err = MigrateQemu(m, i, src, dst)
	if err != nil {
		return err
	}

	err = client.StartMachine(dst, m.Info().Name)
	if err != nil {
		return fmt.Errorf("failed to start remote machine: %s", err)
	}

	return nil
}

func MigrateLxc(m Machine, i Image, src, dst shared.Remote) error {
	mi := m.Info()

	if m.State() == shared.StateUp {
		err := m.Stop()
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

	srcFolder := fmt.Sprintf("%s/lxc/%s", shared.APIConfig.MachinePath, mi.Name)

	err = utils.TarDirectory(srcFolder, fmt.Sprintf("/tmp/%s.tar.gz", mi.Name))
	if err != nil {
		return err
	}

	srcPath := fmt.Sprintf("/tmp/%s.tar.gz", mi.Name)
	dstPath := fmt.Sprintf("%s/lxc/%s.tar.gz", conf.MachinePath, mi.Name)

	err = utils.MakeRemoteDirectories(dst, filepath.Dir(dstPath))
	if err != nil {
		return fmt.Errorf("falied to make remote dirs: %s", err)
	}

	err = utils.SCP(srcPath, dst, dstPath)
	if err != nil {
		return fmt.Errorf("failed to scp to remote: %s", err)
	}

	_, err = client.CreateMachine(dst, *mi)
	if err != nil {
		return fmt.Errorf("failed to create remote machine: %s", err)
	}

	err = m.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop local machine: %s", err)
	}

	return nil
}

func LiveMigrateLxc(m Machine, i Image, src, dst shared.Remote) error {
	err := m.CreateCheckpoint()
	if err != nil {
		return fmt.Errorf("failed to checkpoint: %s", err)
	}

	err = MigrateLxc(m, i, src, dst)
	if err != nil {
		return err
	}

	err = client.StartMachine(dst, m.Info().Name)
	if err != nil {
		return fmt.Errorf("failed to start remote macine: %s", err)
	}

	return nil
}
