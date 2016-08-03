package client

import (
	"encoding/json"
	"fmt"

	"github.com/quadrifoglio/wir/shared"
)

func CreateBackup(target shared.Remote, machine string) (shared.MachineBackup, error) {
	type Response struct {
		ResponseBase
		Content shared.MachineBackup
	}

	var r Response

	data, err := apiRequest(target, "POST", fmt.Sprintf("/machines/%s/backups", machine), nil)
	if err != nil {
		return r.Content, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return r.Content, fmt.Errorf("json: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func ListBackups(target shared.Remote, machine string) ([]shared.MachineBackup, error) {
	type Response struct {
		ResponseBase
		Content []shared.MachineBackup
	}

	var r Response

	data, err := apiRequest(target, "GET", fmt.Sprintf("/machines/%s/backups", machine), nil)
	if err != nil {
		return r.Content, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return r.Content, fmt.Errorf("json: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func RestoreBackup(target shared.Remote, machine, name string) error {
	var r ResponseBase

	data, err := apiRequest(target, "RESTORE", fmt.Sprintf("/machines/%s/backups/%s", machine, name), nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return fmt.Errorf("json: %s", err)
	}

	return apiError(r)
}

func DeleteBackup(target shared.Remote, machine, name string) error {
	var r ResponseBase

	data, err := apiRequest(target, "DELETE", fmt.Sprintf("/machines/%s/backups/%s", machine, name), nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return fmt.Errorf("json: %s", err)
	}

	return apiError(r)
}
