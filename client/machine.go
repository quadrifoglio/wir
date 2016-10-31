package client

import (
	"fmt"
	"github.com/quadrifoglio/wir/shared"
)

// MachineCreate send an mume creation request to the specified remote and
// returns the newly created mume information
func MachineCreate(r shared.RemoteDef, req shared.MachineDef) (shared.MachineDef, error) {
	var m shared.MachineDef

	resp, err := PostJson(r, "/machines", req)
	if err != nil {
		return m, err
	}

	err = DecodeJson(resp, &m)
	if err != nil {
		return m, err
	}

	return m, nil
}

// MachineList fetches all the machines from the specified
// server and returns them as an array
func MachineList(r shared.RemoteDef) ([]shared.MachineDef, error) {
	var ms []shared.MachineDef

	resp, err := Get(r, "/machines")
	if err != nil {
		return nil, err
	}

	err = DecodeJson(resp, &ms)
	if err != nil {
		return nil, err
	}

	return ms, nil
}

// MachineGet fetches the machines from the specified
// server and returns it
func MachineGet(r shared.RemoteDef, id string) (shared.MachineDef, error) {
	var m shared.MachineDef

	resp, err := Get(r, fmt.Sprintf("/machines/%s", id))
	if err != nil {
		return m, err
	}

	err = DecodeJson(resp, &m)
	if err != nil {
		return m, err
	}

	return m, nil
}

// MachineUpdate send an mume update request to the specified remote and
// returns the new mume information
func MachineUpdate(r shared.RemoteDef, id string, req shared.MachineDef) (shared.MachineDef, error) {
	var m shared.MachineDef

	resp, err := PostJson(r, fmt.Sprintf("/machines/%s", id), req)
	if err != nil {
		return m, err
	}

	err = DecodeJson(resp, &m)
	if err != nil {
		return m, err
	}

	return m, nil
}

// MachineDelete send an mume delete request
// to the specified remote
func MachineDelete(r shared.RemoteDef, id string) error {
	resp, err := Delete(r, fmt.Sprintf("/machines/%s", id))
	if err != nil {
		return err
	}

	err = CheckResponse(resp)
	if err != nil {
		return err
	}

	return nil
}

// MachineGetKvmOpts fetches the KVM-specific options
// of the machine
func MachineGetKvmOpts(r shared.RemoteDef, id string) (shared.KvmOptsDef, error) {
	var opts shared.KvmOptsDef

	resp, err := Get(r, fmt.Sprintf("/machines/%s/kvm", id))
	if err != nil {
		return opts, err
	}

	err = DecodeJson(resp, &opts)
	if err != nil {
		return opts, err
	}

	return opts, nil
}

// MachineSetKvmOpts updates the KVM-specified options of the machine
// and returns the updated options
func MachineSetKvmOpts(r shared.RemoteDef, id string, req shared.KvmOptsDef) (shared.KvmOptsDef, error) {
	var opts shared.KvmOptsDef

	resp, err := PostJson(r, fmt.Sprintf("/machines/%s/kvm", id), req)
	if err != nil {
		return opts, err
	}

	err = DecodeJson(resp, &opts)
	if err != nil {
		return opts, err
	}

	return opts, nil
}

// MachineStart sends a machine start request
// to the specified remote
func MachineStart(r shared.RemoteDef, id string) error {
	resp, err := Get(r, fmt.Sprintf("/machines/%s/start", id))
	if err != nil {
		return err
	}

	err = CheckResponse(resp)
	if err != nil {
		return err
	}

	return nil
}

// MachineStop sends a machine stop request
// to the specified remote
func MachineStop(r shared.RemoteDef, id string) error {
	resp, err := Get(r, fmt.Sprintf("/machines/%s/stop", id))
	if err != nil {
		return err
	}

	err = CheckResponse(resp)
	if err != nil {
		return err
	}

	return nil
}

// MachineStatus gets the machine status from
// the specified server and returns it
func MachineStatus(r shared.RemoteDef, id string) (shared.MachineStatusDef, error) {
	var status shared.MachineStatusDef

	resp, err := Get(r, fmt.Sprintf("/machines/%s/status", id))
	if err != nil {
		return status, err
	}

	err = DecodeJson(resp, &status)
	if err != nil {
		return status, err
	}

	return status, nil
}
