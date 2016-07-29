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

type LxcImage struct {
	shared.ImageInfo
}

func (i *LxcImage) Create(info shared.ImageInfo) error {
	url, err := url.Parse(info.Source)
	if err != nil {
		return errors.InvalidURL
	}

	if url.Scheme != "file" && url.Scheme != "scp" && url.Scheme != "http" && url.Scheme != "https" {
		return errors.UnsupportedProto
	}

	path := fmt.Sprintf("%s/lxc/%s", shared.APIConfig.ImagePath, info.Name)

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

	os.Chmod(abs, 777)
	if err != nil {
		return err
	}

	i.Name = info.Name
	i.Type = shared.TypeLXC
	i.Source = abs
	i.Arch = info.Arch
	i.Distro = info.Distro
	i.Release = info.Release

	return nil
}

func (i *LxcImage) Info() shared.ImageInfo {
	return i.ImageInfo
}

func (i *LxcImage) Delete() error {
	// TODO: Implement
	return nil
}
