package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"github.com/quadrifoglio/wir/errors"
)

func FetchResource(url *url.URL, dst io.Writer) error {
	switch url.Scheme {
	case "file":
		return FetchFile(url.Path, dst)
	case "scp":
		return FetchScp(url.Host, url.User, url.Path, dst)
	case "http":
		return FetchFile(url.String(), dst)
	default:
		return errors.UnsupportedProto
	}
}

func FetchFile(path string, dst io.Writer) error {
	src, err := os.Open(path)
	if err != nil {
		return err
	}

	defer src.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	return nil
}

func FetchScp(host string, user *url.Userinfo, path string, dst io.Writer) error {
	srcf := fmt.Sprintf("%s@%s:%s", user.Username(), host, path)
	cmd := exec.Command("scp", "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", srcf, "/dev/stdout")

	cmd.Stdout = dst

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func FetchHTTP(path string, dst io.Writer) error {
	resp, err := http.Get(path)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	_, err = io.Copy(dst, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
