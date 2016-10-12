package system

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/quadrifoglio/wir/utils"
)

// FetchURL fetches the specified resource and
// stores it in the 'dst' file path
func FetchURL(src, dst string) error {
	dir := filepath.Dir(dst)

	if !utils.FileExists(dir) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	if strings.Contains(src, "//") { // This is a URL
		url, err := url.Parse(src)
		if err != nil {
			return err
		}

		switch url.Scheme {
		case "file":
			return CopyFile(src[7:], dst)
		case "http":
		case "https":
			return DownloadHttp(url.String(), dst)
		default:
			break
		}
	} else {
		err := CopyFile(src, dst)
		if err != nil {
			return err
		}
	}

	return fmt.Errorf("Invalid source path (must be URL or file path)")
}

// CopyFile copies the specified 'src' file
// into the 'dst' file path
func CopyFile(src, dst string) error {
	cmd := exec.Command("cp", src, dst)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", utils.OneLine(out))
	}

	return nil
}

// DownloadHttp downloads the specified HTTP ressource
// and stores it in the 'dst' file path
func DownloadHttp(url, dst string) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
