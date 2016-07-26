package image

import (
	"github.com/quadrifoglio/wir/errors"
)

const (
	TypeUnknown = "unknown"
	TypeQemu    = "qemu"
	TypeVz      = "openvz"
	TypeLXC     = "lxc"
)

type Image struct {
	Name          string
	Type          string
	Source        string
	MainPartition int

	// Optional information
	Arch    string
	Distro  string
	Release string
}

func Create(t, name, source, arch, distro, release string, mainPart int) (Image, error) {
	var err error
	var img Image

	switch t {
	case TypeQemu:
		img, err = QemuCreate(name, source)
		break
	case TypeVz:
		img, err = VzCreate(name, source)
		break
	case TypeLXC:
		img, err = LxcCreate(name, source)
		break
	default:
		err = errors.InvalidImageType
		break
	}

	if err != nil {
		return img, err
	}

	img.MainPartition = mainPart
	img.Arch = arch
	img.Distro = distro
	img.Release = release

	return img, nil
}
