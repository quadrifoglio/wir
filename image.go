package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	ImgStateLoading   = "loading"
	ImgStateAvailable = "available"
)

type Image struct {
	ID    int    `json:"id"`
	State string `json:"state"`
	Type  string `json:"type"`
	Path  string `json:"path"`
}

func ImageLoadFile(img *Image) error {
	if _, err := os.Stat(img.Path); os.IsNotExist(err) {
		return ErrImageFileNotFound
	}

	return nil
}

func ImageLoadHTTP(img *Image) error {
	id, err := DatabaseFreeImageId()
	if err != nil {
		return err
	}

	err = os.MkdirAll(fmt.Sprintf("%s/%d", ImagesDir, id), 0755)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%d/%s", ImagesDir, id, filepath.Base(img.Path))
	url := img.Path

	out, err := os.Create(path)
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	log.Println(fmt.Sprintf("Downloading image %d (%s)", id, url))
	go func() {
		defer out.Close()
		defer resp.Body.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			log.Println(fmt.Sprintf("Downloading image %d (%s): %s", id, url, err))
			return
		}

		log.Println(fmt.Sprintf("Downloaded image %d", id))
		img.State = ImgStateAvailable

		err = DatabaseUpdateImage(img)
		if err != nil {
			log.Println(fmt.Sprintf("Downloading image %d (%s): %s", id, url, err))
			return
		}
	}()

	img.Path = path
	return nil
}
