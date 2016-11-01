package client

import (
	"fmt"
	"github.com/quadrifoglio/wir/shared"
)

// CheckpointCreate send a checkpoint creation request to the specified remote and
// returns the newly created checkpoint information
func CheckpointCreate(r shared.RemoteDef, machineId string, req shared.CheckpointDef) (shared.CheckpointDef, error) {
	var chk shared.CheckpointDef

	resp, err := PostJson(r, fmt.Sprintf("/machines/%s/checkpoints", machineId), req)
	if err != nil {
		return chk, err
	}

	err = DecodeJson(resp, &chk)
	if err != nil {
		return chk, err
	}

	return chk, nil
}

// CheckpointList fetches all the checkpoints from the specified
// server and returns them as a array
func CheckpointList(r shared.RemoteDef, machineId string) ([]shared.CheckpointDef, error) {
	var chks []shared.CheckpointDef

	resp, err := Get(r, fmt.Sprintf("/machines/%s/checkpoints", machineId))
	if err != nil {
		return nil, err
	}

	err = DecodeJson(resp, &chks)
	if err != nil {
		return nil, err
	}

	return chks, nil
}

// CheckpointResttore send a checkpoint restore request
// to the specified remote
func CheckpointRestore(r shared.RemoteDef, machineId, name string) error {
	resp, err := Get(r, fmt.Sprintf("/machines/%s/checkpoints/%s/restore", machineId, name))
	if err != nil {
		return err
	}

	err = CheckResponse(resp)
	if err != nil {
		return err
	}

	return nil
}

// CheckpointDelete send a checkpoint delete request
// to the specified remote
func CheckpointDelete(r shared.RemoteDef, machineId, name string) error {
	resp, err := Delete(r, fmt.Sprintf("/machines/%s/checkpoints/%s", machineId, name))
	if err != nil {
		return err
	}

	err = CheckResponse(resp)
	if err != nil {
		return err
	}

	return nil
}
