package inter

import (
	"fmt"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/machine"
)

func MigrateQemu(basePath string, m machine.Machine, i image.Image, src, dst client.Remote) error {
	_, err := client.GetImage(dst, m.Image)
	if err != nil {
		r := client.ImageRequest{i.Name, i.Type, fmt.Sprintf("scp://%s/%s", src.Addr, i.Source)}

		_, err = client.CreateImage(dst, r)
		if err != nil {
			return fmt.Errorf("Remote image: %s", err)
		}
	}

	// TODO: Determine the remote's image directory
	path := basePath + "qemu/" + m.Name + ".img"

	err = Scp(path, dst.Addr, path)
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
