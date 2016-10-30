package client

import (
	"fmt"
	"github.com/quadrifoglio/wir/shared"
)

// VolumeCreate send an volume creation request to the specified remote and
// returns the newly created volume information
func VolumeCreate(r shared.RemoteDef, req shared.VolumeDef) (shared.VolumeDef, error) {
	var vol shared.VolumeDef

	resp, err := PostJson(r, "/volumes", req)
	if err != nil {
		return vol, err
	}

	err = DecodeJson(resp, &vol)
	if err != nil {
		return vol, err
	}

	return vol, nil
}

// VolumeList fetches all the volumes from the specified
// server and returns them as an array
func VolumeList(r shared.RemoteDef) ([]shared.VolumeDef, error) {
	var vols []shared.VolumeDef

	resp, err := GetJson(r, "/volumes")
	if err != nil {
		return nil, err
	}

	err = DecodeJson(resp, &vols)
	if err != nil {
		return nil, err
	}

	return vols, nil
}

// VolumeGet fetches the volumes from the specified
// server and returns it
func VolumeGet(r shared.RemoteDef, id string) (shared.VolumeDef, error) {
	var vol shared.VolumeDef

	resp, err := GetJson(r, fmt.Sprintf("/volumes/%s", id))
	if err != nil {
		return vol, err
	}

	err = DecodeJson(resp, &vol)
	if err != nil {
		return vol, err
	}

	return vol, nil
}

// VolumeUpdate send an volume update request to the specified remote and
// returns the new volume information
func VolumeUpdate(r shared.RemoteDef, id string, req shared.VolumeDef) (shared.VolumeDef, error) {
	var vol shared.VolumeDef

	resp, err := PostJson(r, fmt.Sprintf("/volumes/%s", id), req)
	if err != nil {
		return vol, err
	}

	err = DecodeJson(resp, &vol)
	if err != nil {
		return vol, err
	}

	return vol, nil
}

// VolumeDelete send an volume delete request
// to the specified remote
func VolumeDelete(r shared.RemoteDef, id string) error {
	resp, err := Delete(r, fmt.Sprintf("/volumes/%s", id))
	if err != nil {
		return err
	}

	err = CheckResponse(resp)
	if err != nil {
		return err
	}

	return nil
}
