package client

import (
	"encoding/json"
	"fmt"

	"github.com/quadrifoglio/wir/shared"
)

func ListImages(target shared.Remote) ([]shared.Image, error) {
	type Response struct {
		ResponseBase
		Content []shared.Image
	}

	var r Response

	data, err := apiRequest(target, "GET", "/images", nil)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return nil, fmt.Errorf("json: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func CreateImage(target shared.Remote, i shared.Image) (shared.Image, error) {
	type Response struct {
		ResponseBase
		Content shared.Image
	}

	var r Response

	data, err := json.Marshal(i)
	if err != nil {
		return r.Content, fmt.Errorf("json: %s", err)
	}

	data, err = apiRequest(target, "POST", "/images", data)
	if err != nil {
		return r.Content, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return r.Content, fmt.Errorf("json: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func GetImage(target shared.Remote, name string) (shared.Image, error) {
	type Response struct {
		ResponseBase
		Content shared.Image
	}

	var r Response

	data, err := apiRequest(target, "GET", "/images/"+name, nil)
	if err != nil {
		return r.Content, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return r.Content, fmt.Errorf("json: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func DeleteImage(target shared.Remote, name string) error {
	type Response struct {
		ResponseBase
		Content interface{}
	}

	var r Response

	data, err := apiRequest(target, "DELETE", "/images/"+name, nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return fmt.Errorf("json: %s", err)
	}

	return apiError(r.ResponseBase)
}
