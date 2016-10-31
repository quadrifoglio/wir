package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/quadrifoglio/wir/shared"
)

// RemoteDefURL returns the compltement HTTP URL
// corresponding to the specified remote and path
func RemoteDefURL(r shared.RemoteDef, path string) string {
	return fmt.Sprintf("http://%s:%d%s", r.Host, r.Port, path)
}

// Get sends an HTTP GET request to the remote
// and returns the HTTP response
func Get(r shared.RemoteDef, path string) (*http.Response, error) {
	resp, err := http.Get(RemoteDefURL(r, path))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// PostJson sends an HTTP POST request containing
// the specified data encoded as JSON and returns the response
func PostJson(r shared.RemoteDef, path string, req interface{}) (*http.Response, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(RemoteDefURL(r, path), "encoding/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Delete sends an HTTP DELETE request to the remote
// and returns an error if any occurred
func Delete(r shared.RemoteDef, path string) (*http.Response, error) {
	var c http.Client

	req, err := http.NewRequest("DELETE", RemoteDefURL(r, path), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CheckResponse checks the HTTP response and
// returns an error message if it is unsuccessful
func CheckResponse(resp *http.Response) error {
	type errResp struct {
		Error string
	}

	if resp.StatusCode != http.StatusOK {
		var e errResp

		err := json.NewDecoder(resp.Body).Decode(&e)
		if err != nil {
			return fmt.Errorf("Remote: HTTP %d", resp.Status)
		}

		return fmt.Errorf("Remote: HTTP %d: %s", resp.StatusCode, e.Error)
	}

	return nil
}

// DecodeJson checks the specified HTTP response
// and if it it a success, it parses the JSON into 'data'
func DecodeJson(resp *http.Response, data interface{}) error {
	err := CheckResponse(resp)
	if err != nil {
		return err
	}

	return json.NewDecoder(resp.Body).Decode(data)
}
