package server

import (
	"fmt"
)

var (
	GlobalNodeID byte

	GlobalImagePath   string
	GlobalMachinePath string
)

// Init initializes the parameters
// of the server
func Init(nodeId byte, database string, img, machine string) error {
	GlobalNodeID = nodeId
	GlobalImagePath = img
	GlobalMachinePath = machine

	return InitDatabase(database)
}

// ImageFile returns the path of the file
// for the specified image name
func ImageFile(name string) string {
	return fmt.Sprintf("%s/%s/img.data", GlobalImagePath, name)
}

// MachinePath returns the current path
// for the specified machine name
func MachinePath(name string) string {
	return fmt.Sprintf("%s/%s", GlobalMachinePath, name)
}
