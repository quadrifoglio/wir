package utils

import (
	"net/url"

	"github.com/quadrifoglio/wir/errors"
)

func FetchResource(url *url.URL) ([]byte, error) {
	switch url.Scheme {
	case "file":
		return FetchFile(url.Path)
	case "scp":
		return FetchScp(url.Host, url.User, url.Path)
	case "http":
		return FetchFile(url.String())
	default:
		return nil, errors.UnsupportedProto
	}
}

func FetchFile(path string) ([]byte, error) {
	return nil, nil
}

func FetchScp(host string, user *url.Userinfo, path string) ([]byte, error) {
	return nil, nil
}

func FetchHTTP(path string) ([]byte, error) {
	return nil, nil
}
