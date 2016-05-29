package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	apiURL = "http://127.0.0.1:1997"
)

type ResponseBase struct {
	Status  int
	Message string
}

func apiRequest(method, url string, body []byte) ([]byte, error) {
	var c http.Client

	req, err := http.NewRequest(method, apiURL+url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("HTTP: %s", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP: %s", err)
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("I/O: %s", err)
	}

	return data, nil
}

func apiError(r ResponseBase) error {
	if r.Status != 200 {
		return fmt.Errorf("Error response from API: %s", r.Message)
	}

	return nil
}
