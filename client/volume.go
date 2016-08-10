package client

import (
	"encoding/json"
	"fmt"

	"github.com/quadrifoglio/wir/shared"
)

func CreateVolume(target shared.Remote, machine, name string, size uint64) (shared.Volume, error) {
	type Response struct {
		ResponseBase
		Content shared.Volume
	}

	var r Response

	data, err := json.Marshal(shared.Volume{name, size})
	if err != nil {
		return r.Content, fmt.Errorf("json: %s", err)
	}

	data, err = apiRequest(target, "POST", fmt.Sprintf("/machines/%s/volumes", machine), data)
	if err != nil {
		return r.Content, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return r.Content, fmt.Errorf("json: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func ListVolumes(target shared.Remote, machine string) ([]shared.Volume, error) {
	type Response struct {
		ResponseBase
		Content []shared.Volume
	}

	var r Response

	data, err := apiRequest(target, "GET", fmt.Sprintf("/machines/%s/volumes", machine), nil)
	if err != nil {
		return r.Content, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return r.Content, fmt.Errorf("json: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func DeleteVolume(target shared.Remote, machine, name string) error {
	var r ResponseBase

	data, err := apiRequest(target, "DELETE", fmt.Sprintf("/machines/%s/volumes/%s", machine, name), nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return fmt.Errorf("json: %s", err)
	}

	return apiError(r)
}
