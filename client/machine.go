package client

import (
	"encoding/json"
	"fmt"

	"github.com/quadrifoglio/wir/machine"
)

type MachineRequest struct {
	Image   string
	Cores   int
	Memory  int
	Network machine.NetworkMode
}

func ListMachines() ([]machine.Machine, error) {
	type Response struct {
		ResponseBase
		Content []machine.Machine
	}

	var r Response

	data, err := apiRequest("GET", "/machines", nil)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return nil, fmt.Errorf("JSON: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func CreateMachine(i MachineRequest) (machine.Machine, error) {
	type Response struct {
		ResponseBase
		Content machine.Machine
	}

	var r Response

	data, err := json.Marshal(i)
	if err != nil {
		return r.Content, fmt.Errorf("JSON: %s", err)
	}

	data, err = apiRequest("POST", "/machines", data)
	if err != nil {
		return r.Content, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return r.Content, fmt.Errorf("JSON: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func GetMachine(id string) (machine.Machine, error) {
	type Response struct {
		ResponseBase
		Content machine.Machine
	}

	var r Response

	data, err := apiRequest("GET", "/machines/"+id, nil)
	if err != nil {
		return r.Content, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return r.Content, fmt.Errorf("JSON: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func StartMachine(id string) error {
	type Response struct {
		ResponseBase
		Content machine.Machine
	}

	var r Response

	data, err := apiRequest("START", "/machines/"+id, nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return fmt.Errorf("JSON: %s", err)
	}

	return apiError(r.ResponseBase)
}

func StopMachine(id string) error {
	type Response struct {
		ResponseBase
		Content machine.Machine
	}

	var r Response

	data, err := apiRequest("STOP", "/machines/"+id, nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return fmt.Errorf("JSON: %s", err)
	}

	return apiError(r.ResponseBase)
}

func DeleteMachine(id string) error {
	type Response struct {
		ResponseBase
		Content interface{}
	}

	var r Response

	data, err := apiRequest("DELETE", "/machines/"+id, nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return fmt.Errorf("JSON: %s", err)
	}

	return apiError(r.ResponseBase)
}
