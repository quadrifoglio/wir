package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/quadrifoglio/wir/global"
)

type APIConfig struct {
	Address      string
	DatabaseFile string
	ImagePath    string
	MachinePath  string
}

type ResponseBase struct {
	Status  int
	Message string
}

func GetConfig(target global.Remote) (APIConfig, error) {
	type Response struct {
		ResponseBase
		Content struct {
			Configuration APIConfig
		}
	}

	var r Response

	data, err := apiRequest(target, "GET", "/", nil)
	if err != nil {
		return APIConfig{}, err
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		return APIConfig{}, fmt.Errorf("json: %s", err)
	}

	return r.Content.Configuration, apiError(r.ResponseBase)
}

func apiRequest(target global.Remote, method, url string, body []byte) ([]byte, error) {
	var c http.Client

	req, err := http.NewRequest(method, fmt.Sprintf("http://%s:%d%s", target.Addr, target.APIPort, url), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("http: %s", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %s", err)
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("i/o: %s", err)
	}

	return data, nil
}

func apiError(r ResponseBase) error {
	if r.Status != 200 {
		return fmt.Errorf("error response from api: %s", r.Message)
	}

	return nil
}
