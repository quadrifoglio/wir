package client

import (
	"encoding/json"
	"fmt"

	"github.com/quadrifoglio/wir/machine"
)

type MachineRequest struct {
	Name    string
	Image   string
	Cores   int
	Memory  int
	Network machine.NetworkMode
}

func ListMachines(target Remote) ([]machine.Machine, error) {
	type Response struct {
		ResponseBase
		Content []machine.Machine
	}

	var r Response

	data, err := apiRequest(target, "GET", "/machines", nil)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return nil, fmt.Errorf("JSON: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func CreateMachine(target Remote, i MachineRequest) (machine.Machine, error) {
	type Response struct {
		ResponseBase
		Content machine.Machine
	}

	var r Response

	data, err := json.Marshal(i)
	if err != nil {
		return r.Content, fmt.Errorf("JSON: %s", err)
	}

	data, err = apiRequest(target, "POST", "/machines", data)
	if err != nil {
		return r.Content, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return r.Content, fmt.Errorf("JSON: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func GetMachine(target Remote, name string) (machine.Machine, error) {
	type Response struct {
		ResponseBase
		Content machine.Machine
	}

	var r Response

	data, err := apiRequest(target, "GET", "/machines/"+name, nil)
	if err != nil {
		return r.Content, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return r.Content, fmt.Errorf("JSON: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func StartMachine(target Remote, name string) error {
	type Response struct {
		ResponseBase
		Content machine.Machine
	}

	var r Response

	data, err := apiRequest(target, "START", "/machines/"+name, nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return fmt.Errorf("JSON: %s", err)
	}

	return apiError(r.ResponseBase)
}

func StopMachine(target Remote, name string) error {
	type Response struct {
		ResponseBase
		Content machine.Machine
	}

	var r Response

	data, err := apiRequest(target, "STOP", "/machines/"+name, nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return fmt.Errorf("JSON: %s", err)
	}

	return apiError(r.ResponseBase)
}

func DeleteMachine(target Remote, name string) error {
	type Response struct {
		ResponseBase
		Content interface{}
	}

	var r Response

	data, err := apiRequest(target, "DELETE", "/machines/"+name, nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return fmt.Errorf("JSON: %s", err)
	}

	return apiError(r.ResponseBase)
}
