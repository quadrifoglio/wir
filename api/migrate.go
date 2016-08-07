package api

import (
	"fmt"
	"os"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/utils"
)

func MigrateImage(i shared.Image, src, dst shared.Remote) error {
	_, err := client.GetImage(dst, i.Name)
	if err == nil {
		return nil
	}

	r := shared.Image{
		i.Name,
		i.Type,
		fmt.Sprintf("scp://%s/%s", src.Addr, i.Source),
		i.Desc,
		i.MainPartition,
	}

	_, err = client.CreateImage(dst, r)
	if err != nil {
		return fmt.Errorf("failed to create remote image: %s", err)
	}

	return nil
}

func MigrateMachine(m Machine, i shared.Image, src, dst shared.Remote) error {
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

	mi := m.Info()
	basePath := fmt.Sprintf("%s/%s/%s", shared.APIConfig.MachinePath, i.Type, mi.Name)
	srcPath := fmt.Sprintf("/tmp/%s.tar.gz", mi.Name)
	dstPath := fmt.Sprintf("%s/%s.tar.gz", conf.MigrationPath, mi.Name)

	err = utils.TarDirectory(basePath, srcPath)
	if err != nil {
		return err
	}

	err = utils.SCP(srcPath, dst, dstPath)
	if err != nil {
		return err
	}

	os.Remove(srcPath)

	_, err = client.CreateMachine(dst, *mi)
	if err != nil {
		return fmt.Errorf("failed to create remote machine: %s", err)
	}

	return nil
}

func LiveMigrateMachine(m Machine, i shared.Image, src, dst shared.Remote) error {
	err := m.CreateCheckpoint("wird_migration")
	if err != nil {
		return fmt.Errorf("failed to checkpoint: %s", err)
	}

	err = MigrateMachine(m, i, src, dst)
	if err != nil {
		return err
	}

	err = client.StartMachine(dst, m.Info().Name)
	if err != nil {
		return fmt.Errorf("failed to start remote macine: %s", err)
	}

	return nil
}
