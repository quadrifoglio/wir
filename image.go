package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	ImgStateLoading   = "loading"
	ImgStateAvailable = "available"
	ImgStateError     = "error"
)

type Image struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
	Type  string `json:"type"`
	Path  string `json:"path"`
}

func ImageLoadFile(img *Image) error {
	if _, err := os.Stat(img.Path); os.IsNotExist(err) {
		return ErrImageFileNotFound
	}

	img.State = ImgStateAvailable

	return nil
}

func ImageLoadHTTP(img *Image) error {
	id, err := DatabaseFreeImageId()
	if err != nil {
		return err
	}

	err = os.MkdirAll(fmt.Sprintf("%s/%d", Config.ImagesDir, id), 0755)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%d/%s", Config.ImagesDir, id, filepath.Base(img.Path))
	url := img.Path

	out, err := os.Create(path)
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	log.Println(fmt.Sprintf("Started downloading (HTTP) image %d (%s)", id, url))
	go func() {
		defer out.Close()
		defer resp.Body.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			img.State = ImgStateError
			log.Println(fmt.Sprintf("Error downloading (HTTP) image %d (%s): %s", id, url, err))
			return
		} else {
			log.Println(fmt.Sprintf("Downloaded (HTTP) image %d", id))
			img.State = ImgStateAvailable
		}

		err = DatabaseUpdateImage(img)
		if err != nil {
			log.Println(fmt.Sprintf("Error saving image %d data (%s): %s", id, url, err))
			return
		}
	}()

	img.Path = path
	return nil
}

func ImageLoadSCP(img *Image) error {
	id, err := DatabaseFreeImageId()
	if err != nil {
		return err
	}

	err = os.MkdirAll(fmt.Sprintf("%s/%d", Config.ImagesDir, id), 0755)
	if err != nil {
		return err
	}

	i := strings.Index(img.Path, ":")
	if i == -1 {
		return fmt.Errorf("Invalid scp path")
	}

	src := img.Path
	path := fmt.Sprintf("%s/%d/%s", Config.ImagesDir, id, filepath.Base(img.Path[i+1:]))
	cmd := exec.Command("scp", "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", img.Path, path)

	err = cmd.Start()
	if err != nil {
		return err
	}

	log.Println(fmt.Sprintf("Started downloading (SCP) image %d (%s)", id, src))
	go func() {
		err := cmd.Wait()
		if err != nil {
			img.State = ImgStateError
			log.Println(fmt.Sprintf("Error downloading (SCP) image %d (%s): %s", id, src, err))
		} else {
			img.Path = path
			img.State = ImgStateAvailable

			log.Println(fmt.Sprintf("Downloaded (SCP) image %d", id))
		}

		err = DatabaseUpdateImage(img)
		if err != nil {
			log.Println(fmt.Sprintf("Error saving image %d data (%s): %s", id, src, err))
		}
	}()

	return nil
}
