package client

import (
	"encoding/json"
	"fmt"

	"github.com/quadrifoglio/wir/shared"
)

func CreateInterface(target shared.Remote, machine, mode, mac, ip string) (shared.NetworkDevice, error) {
	type Response struct {
		ResponseBase
		Content shared.NetworkDevice
	}

	var r Response

	data, err := json.Marshal(shared.NetworkDevice{mode, mac, ip})
	if err != nil {
		return r.Content, fmt.Errorf("json: %s", err)
	}

	data, err = apiRequest(target, "POST", fmt.Sprintf("/machines/%s/interfaces", machine), data)
	if err != nil {
		return r.Content, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return r.Content, fmt.Errorf("json: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func UpdateInterface(target shared.Remote, machine string, index int, mode, mac, ip string) (shared.NetworkDevice, error) {
	type Response struct {
		ResponseBase
		Content shared.NetworkDevice
	}

	var r Response

	data, err := json.Marshal(shared.NetworkDevice{mode, mac, ip})
	if err != nil {
		return r.Content, fmt.Errorf("json: %s", err)
	}

	data, err = apiRequest(target, "POST", fmt.Sprintf("/machines/%s/interfaces/%d", machine, index), data)
	if err != nil {
		return r.Content, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return r.Content, fmt.Errorf("json: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func ListInterfaces(target shared.Remote, name string) ([]shared.NetworkDevice, error) {
	type Response struct {
		ResponseBase
		Content []shared.NetworkDevice
	}

	var r Response

	data, err := apiRequest(target, "GET", fmt.Sprintf("/machines/%s/interfaces", name), nil)
	if err != nil {
		return r.Content, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return r.Content, fmt.Errorf("json: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func DeleteInterface(target shared.Remote, machine string, index int) error {
	var r ResponseBase

	data, err := apiRequest(target, "DELETE", fmt.Sprintf("/machines/%s/interfaces/%d", machine, index), nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return fmt.Errorf("json: %s", err)
	}

	return apiError(r)
}
