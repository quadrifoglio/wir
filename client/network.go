package client

import (
	"fmt"
	"github.com/quadrifoglio/wir/shared"
)

// NetworkCreate send an network creation request to the specified remote and
// returns the newly created network information
func NetworkCreate(r shared.RemoteDef, req shared.NetworkDef) (shared.NetworkDef, error) {
	var netw shared.NetworkDef

	resp, err := PostJson(r, "/networks", req)
	if err != nil {
		return netw, err
	}

	err = DecodeJson(resp, &netw)
	if err != nil {
		return netw, err
	}

	return netw, nil
}

// NetworkList fetches all the networks from the specified
// server and returns them as an array
func NetworkList(r shared.RemoteDef) ([]shared.NetworkDef, error) {
	var netws []shared.NetworkDef

	resp, err := GetJson(r, "/networks")
	if err != nil {
		return nil, err
	}

	err = DecodeJson(resp, &netws)
	if err != nil {
		return nil, err
	}

	return netws, nil
}

// NetworkGet fetches the networks from the specified
// server and returns it
func NetworkGet(r shared.RemoteDef, id string) (shared.NetworkDef, error) {
	var netw shared.NetworkDef

	resp, err := GetJson(r, fmt.Sprintf("/networks/%s", id))
	if err != nil {
		return netw, err
	}

	err = DecodeJson(resp, &netw)
	if err != nil {
		return netw, err
	}

	return netw, nil
}

// NetworkUpdate send an network update request to the specified remote and
// returns the new network information
func NetworkUpdate(r shared.RemoteDef, id string, req shared.NetworkDef) (shared.NetworkDef, error) {
	var netw shared.NetworkDef

	resp, err := PostJson(r, fmt.Sprintf("/networks/%s", id), req)
	if err != nil {
		return netw, err
	}

	err = DecodeJson(resp, &netw)
	if err != nil {
		return netw, err
	}

	return netw, nil
}

// NetworkDelete send an network delete request
// to the specified remote
func NetworkDelete(r shared.RemoteDef, id string) error {
	resp, err := Delete(r, fmt.Sprintf("/networks/%s", id))
	if err != nil {
		return err
	}

	err = CheckResponse(resp)
	if err != nil {
		return err
	}

	return nil
}
