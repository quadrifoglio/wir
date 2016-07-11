package inter

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/quadrifoglio/wir/client"
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

func MigrateQemu(basePath string, m machine.Machine, i image.Image, src, dst client.Remote) error {
	err := MigrateImage(i, src, dst)
	if err != nil {
		return err
	}

	conf, err := client.GetConfig(dst)
	if err != nil {
		return fmt.Errorf("failed to get remote config: %s", err)
	}

	srcPath := basePath + "qemu/" + m.Name + ".qcow2"
	dstPath := conf.MachinePath + "qemu/" + m.Name + ".qcow2"

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

	return nil
}

func MigrateLxc(basePath string, m machine.Machine, i image.Image, src, dst client.Remote) error {
	basePath = basePath + "lxc"

	err := MigrateImage(i, src, dst)
	if err != nil {
		return err
	}

	conf, err := client.GetConfig(dst)
	if err != nil {
		return fmt.Errorf("Remote configuration: %s", err)
	}

	tarball := fmt.Sprintf("%s/%s.tar.gz", basePath, m.Name)
	dstPath := conf.MachinePath + "lxc/" + m.Name
	cmd := exec.Command("tar", "--numeric-owner", "-czvf", tarball, "-C", basePath, m.Name)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("tar: %s", err)
	}

	err = RemoteMkdir(dst, filepath.Dir(dstPath))
	if err != nil {
		return fmt.Errorf("MkdirRemote: %s", err)
	}

	err = SCP(tarball, dst, dstPath+".tar.gz")
	if err != nil {
		return fmt.Errorf("SCP to remote: %s", err)
	}

	r := client.MachineRequest{m.Name, m.Image, m.Cores, m.Memory, m.Network}

	_, err = client.CreateMachine(dst, r)
	if err != nil {
		return fmt.Errorf("Remote machine: %s", err)
	}

	return nil
}
