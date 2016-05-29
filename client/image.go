package client

import (
	"encoding/json"
	"fmt"

	"github.com/quadrifoglio/wir/image"
)

type ImageRequest struct {
	Name   string
	Type   int
	Source string
}

func ListImages() ([]image.Image, error) {
	type Response struct {
		ResponseBase
		Content []image.Image
	}

	var r Response

	data, err := apiRequest("GET", "/images", nil)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return nil, fmt.Errorf("JSON: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func CreateImage(i ImageRequest) (image.Image, error) {
	type Response struct {
		ResponseBase
		Content image.Image
	}

	var r Response

	data, err := json.Marshal(i)
	if err != nil {
		return r.Content, fmt.Errorf("JSON: %s", err)
	}

	data, err = apiRequest("POST", "/images", data)
	if err != nil {
		return r.Content, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return r.Content, fmt.Errorf("JSON: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func GetImage(name string) (image.Image, error) {
	type Response struct {
		ResponseBase
		Content image.Image
	}

	var r Response

	data, err := apiRequest("GET", "/images/"+name, nil)
	if err != nil {
		return r.Content, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return r.Content, fmt.Errorf("JSON: %s", err)
	}

	return r.Content, apiError(r.ResponseBase)
}

func DeleteImage(name string) error {
	type Response struct {
		ResponseBase
		Content interface{}
	}

	var r Response

	data, err := apiRequest("DELETE", "/images/"+name, nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return fmt.Errorf("JSON: %s", err)
	}

	return apiError(r.ResponseBase)
}
