package image

import (
	"net/url"
	"os"
	"path/filepath"

	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/utils"
)

func VzCreate(name, src, basePath string) (Image, error) {
	var i Image

	url, err := url.Parse(src)
	if err != nil {
		return i, errors.InvalidURL
	}

	if url.Scheme != "file" && url.Scheme != "scp" && url.Scheme != "http" {
		return i, errors.UnsupportedProto
	}

	path := basePath + "vz/" + name + ".tar.gz"

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

	err = os.MkdirAll("/var/lib/vz/template/cache", 0777)
	if err != nil {
		return i, err
	}

	err = os.Symlink(abs, "/var/lib/vz/template/cache/"+filepath.Base(abs))
	if err != nil {
		return i, err
	}

	i.Name = name
	i.Type = TypeVz
	i.Source = abs

	return i, nil
}
