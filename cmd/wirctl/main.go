package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/quadrifoglio/wir/shared"
)

var (
	// Global flags
	CRemote = kingpin.Flag("remote", "Remote API server (host:port)").Default("127.0.0.1:8000").String()

	// Image command
	CImageCommand = kingpin.Command("image", "Images manipulation actions")

	CImageList = CImageCommand.Command("list", "List all the images")

	// Image creation
	CImageCreate       = CImageCommand.Command("create", "Create a new image")
	CImageCreateName   = CImageCreate.Flag("name", "Image name").Required().String()
	CImageCreateType   = CImageCreate.Flag("type", "Image type (kvm, vz)").Required().String()
	CImageCreateSource = CImageCreate.Flag("source", "Image source (scheme://[userinfo@][host]/path)").Required().String()

	// Image update
	CImageUpdate     = CImageCommand.Command("update", "Update an image")
	CImageUpdateID   = CImageUpdate.Arg("id", "Image ID").Required().String()
	CImageUpdateName = CImageUpdate.Flag("name", "New image name").Required().String()

	// Image delete
	CImageDelete   = CImageCommand.Command("delete", "Delete an image")
	CImageDeleteID = CImageDelete.Arg("id", "Image ID").Required().String()
)

// Fatal displays the error and
// then exits with an error code
func Fatal(err error) {
	fmt.Println(err)
	os.Exit(1)
}

// GetRemote returns a RemoteDef structure
// corresponding to the requested remote server
func GetRemote() shared.RemoteDef {
	l := strings.Split(*CRemote, ":")
	if len(l) < 2 {
		Fatal(fmt.Errorf("Invalid remote, must be host:port"))
	}

	port, err := strconv.Atoi(l[1])
	if err != nil {
		Fatal(fmt.Errorf("Invalid remote port"))
	}

	return shared.RemoteDef{l[0], port}
}

func main() {
	switch kingpin.Parse() {
	case "image create":
		ImageCreate()
		break
	case "image list":
		ImageList()
		break
	case "image update":
		ImageUpdate()
		break
	case "image delete":
		ImageDelete()
		break
	default:
		Fatal(fmt.Errorf("Invalid command"))
		break
	}
}
