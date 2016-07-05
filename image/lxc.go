package image

import (
	"net/url"
	"os"
	"path/filepath"

	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/utils"
)

func LxcCreate(name, src, basePath string) (Image, error) {
	var i Image

	url, err := url.Parse(src)
	if err != nil {
		return i, errors.InvalidURL
	}

	if url.Scheme != "file" && url.Scheme != "scp" && url.Scheme != "http" && url.Scheme != "https" {
		return i, errors.UnsupportedProto
	}

	path := basePath + "lxc/" + name

	err = os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return i, err
	}

	f, err := os.Create(path)
	if err != nil {
		return i, err
	}

	defer f.Close()

	err = utils.FetchResource(url, f)
	if err != nil {
		return i, err
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return i, err
	}

	os.Chmod(abs, 777)
	if err != nil {
		return i, err
	}

	i.Name = name
	i.Type = TypeLXC
	i.Source = abs

	return i, nil
}
