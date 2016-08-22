package client

import (
	"encoding/json"
	"fmt"

	"github.com/quadrifoglio/wir/shared"
)

func ListNetworks(target shared.Remote) ([]shared.Network, error) {
	type Response struct {
		ResponseBase
		Content []shared.Network
	}

	var r Response

	data, err := apiRequest(target, "GET", "/networks", nil)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return nil, fmt.Errorf("json: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func CreateNetwork(target shared.Remote, i shared.Network) (shared.Network, error) {
	type Response struct {
		ResponseBase
		Content shared.Network
	}

	var r Response

	data, err := json.Marshal(i)
	if err != nil {
		return r.Content, fmt.Errorf("json: %s", err)
	}

	data, err = apiRequest(target, "POST", "/networks", data)
	if err != nil {
		return r.Content, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return r.Content, fmt.Errorf("json: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func GetNetwork(target shared.Remote, name string) (shared.Network, error) {
	type Response struct {
		ResponseBase
		Content shared.Network
	}

	var r Response

	data, err := apiRequest(target, "GET", "/networks/"+name, nil)
	if err != nil {
		return r.Content, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return r.Content, fmt.Errorf("json: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func DeleteNetwork(target shared.Remote, name string) error {
	type Response struct {
		ResponseBase
		Content interface{}
	}

	var r Response

	data, err := apiRequest(target, "DELETE", "/networks/"+name, nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return fmt.Errorf("json: %s", err)
	}

	return apiError(r.ResponseBase)
}
