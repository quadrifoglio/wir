package machine

import (
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/utils"
)

func QemuCreate(img *image.Image, cores, memory int) (Machine, error) {
	var m Machine
	m.ID = utils.UniqueID()
	m.Image = img.Name
	m.Cores = cores
	m.Memory = memory

	return m, nil
}
