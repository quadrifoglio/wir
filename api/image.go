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

func CreateImage(img *shared.Image) error {
	if img.Type != "qemu" && img.Type != "lxc" {
		return errors.InvalidImageType
	}

	url, err := url.Parse(img.Source)
	if err != nil {
		return errors.InvalidURL
	}

	if url.Scheme != "file" && url.Scheme != "scp" && url.Scheme != "http" {
		return errors.UnsupportedProto
	}

	path := fmt.Sprintf("%s/qemu/%s%s", shared.APIConfig.ImagePath, img.Name, filepath.Ext(img.Source))

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

	/*os.Chmod(abs, 777)
	if err != nil {
		return err
	}*/

	img.Source = abs
	return nil
}
