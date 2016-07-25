package client

import (
	"encoding/json"
	"fmt"

	"github.com/quadrifoglio/wir/global"
	"github.com/quadrifoglio/wir/machine"
)

type MachineRequest struct {
	Name    string
	Image   string
	Cores   int
	Memory  int
	Network machine.NetworkSetup
}

type LinuxSysprep struct {
	Hostname   string
	RootPasswd string
}

func ListMachines(target global.Remote) ([]machine.Machine, error) {
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
		return nil, fmt.Errorf("json: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func CreateMachine(target global.Remote, i MachineRequest) (machine.Machine, error) {
	type Response struct {
		ResponseBase
		Content machine.Machine
	}

	var r Response

	data, err := json.Marshal(i)
	if err != nil {
		return r.Content, fmt.Errorf("json: %s", err)
	}

	data, err = apiRequest(target, "POST", "/machines", data)
	if err != nil {
		return r.Content, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return r.Content, fmt.Errorf("json: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func GetMachine(target global.Remote, name string) (machine.Machine, error) {
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
		return r.Content, fmt.Errorf("json: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func UpdateMachine(target global.Remote, name string, i MachineRequest) error {
	var r ResponseBase

	data, err := json.Marshal(i)
	if err != nil {
		return fmt.Errorf("json: %s", err)
	}

	data, err = apiRequest(target, "POST", "/machines/"+name, data)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return fmt.Errorf("json: %s", err)
	}

	return apiError(r)
}

func StartMachine(target global.Remote, name string) error {
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
		return fmt.Errorf("json: %s", err)
	}

	return apiError(r.ResponseBase)
}

func StopMachine(target global.Remote, name string) error {
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
		return fmt.Errorf("json: %s", err)
	}

	return apiError(r.ResponseBase)
}

func MigrateMachine(target global.Remote, name string, remote global.Remote, live bool) error {
	var r ResponseBase

	type request struct {
		Target global.Remote
		Live   bool
	}

	data, err := json.Marshal(request{remote, live})
	if err != nil {
		return fmt.Errorf("json: %s", err)
	}

	resp, err := apiRequest(target, "MIGRATE", "/machines/"+name, data)
	if err != nil {
		return err
	}

	err = json.Unmarshal(resp, &r)
	if err != nil {
		return fmt.Errorf("json: %s", err)
	}

	return apiError(r)
}

func DeleteMachine(target global.Remote, name string) error {
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
		return fmt.Errorf("json: %s", err)
	}

	return apiError(r.ResponseBase)
}
