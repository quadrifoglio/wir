package image

import (
	"log"
	"net/url"
	"os"

	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/utils"
)

func QemuFetch(name, src, basePath string) error {
	url, err := url.Parse(src)
	if err != nil {
		return errors.InvalidURL
	}

	if url.Scheme != "file" && url.Scheme != "scp" && url.Scheme != "http" {
		return errors.UnsupportedProto
	}

	path := basePath + "qemu/" + name + ".img"

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()

	err = utils.FetchResource(url, f)
	if err != nil {
		return err
	}

	log.Printf("Fetched image %s to %s\n", name, path)
	return nil
}
