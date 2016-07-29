package api

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/utils"
)

type QemuImage struct {
	shared.ImageInfo
}

func (i *QemuImage) Create(info shared.ImageInfo) error {
	url, err := url.Parse(info.Source)
	if err != nil {
		return errors.InvalidURL
	}

	if url.Scheme != "file" && url.Scheme != "scp" && url.Scheme != "http" {
		return errors.UnsupportedProto
	}

	path := fmt.Sprintf("%s/qemu/%s.qcow2", shared.APIConfig.ImagePath, info.Name)

	err = os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()

	err = utils.FetchResource(url, f)
	if err != nil {
		return err
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	i.Name = info.Name
	i.Type = shared.TypeQemu
	i.Source = abs
	i.Arch = info.Arch
	i.Distro = info.Distro
	i.Release = info.Release

	return nil
}

func (i *QemuImage) Info() shared.ImageInfo {
	return i.ImageInfo
}

func (i *QemuImage) Delete() error {
	// TODO: Implement
	return nil
}
