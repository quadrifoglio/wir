package client

import (
	"fmt"
	"github.com/quadrifoglio/wir/shared"
)

// ImageCreate send an image creation request to the specified remote and
// returns the newly created image information
func ImageCreate(r shared.RemoteDef, req shared.ImageDef) (shared.ImageDef, error) {
	var img shared.ImageDef

	resp, err := PostJson(r, "/images", req)
	if err != nil {
		return img, err
	}

	err = DecodeJson(resp, &img)
	if err != nil {
		return img, err
	}

	return img, nil
}

// ImageList fetches all the images from the specified
// server and returns them as an array
func ImageList(r shared.RemoteDef) ([]shared.ImageDef, error) {
	var imgs []shared.ImageDef

	resp, err := GetJson(r, "/images")
	if err != nil {
		return nil, err
	}

	err = DecodeJson(resp, &imgs)
	if err != nil {
		return nil, err
	}

	return imgs, nil
}

// ImageGet fetches the images from the specified
// server and returns it
func ImageGet(r shared.RemoteDef, id string) (shared.ImageDef, error) {
	var img shared.ImageDef

	resp, err := GetJson(r, fmt.Sprintf("/images/%s", id))
	if err != nil {
		return img, err
	}

	err = DecodeJson(resp, &img)
	if err != nil {
		return img, err
	}

	return img, nil
}

// ImageUpdate send an image update request to the specified remote and
// returns the new image information
func ImageUpdate(r shared.RemoteDef, id string, req shared.ImageDef) (shared.ImageDef, error) {
	var img shared.ImageDef

	resp, err := PostJson(r, fmt.Sprintf("/images/%s", id), req)
	if err != nil {
		return img, err
	}

	err = DecodeJson(resp, &img)
	if err != nil {
		return img, err
	}

	return img, nil
}

// ImageDelete send an image delete request
// to the specified remote
func ImageDelete(r shared.RemoteDef, id string) error {
	resp, err := Delete(r, fmt.Sprintf("/images/%s", id))
	if err != nil {
		return err
	}

	err = CheckResponse(resp)
	if err != nil {
		return err
	}

	return nil
}
