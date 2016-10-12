package server

import (
	"fmt"
)

var (
	GlobalNodeID byte

	GlobalImagePath   string
	GlobalVolumePath  string
	GlobalMachinePath string
)

// Init initializes the parameters
// of the server
func Init(nodeId byte, db string, img, vol, machine string) error {
	GlobalNodeID = nodeId
	GlobalImagePath = img
	GlobalVolumePath = vol
	GlobalMachinePath = machine

	return InitDatabase(db)
}

// ImageFile returns the path of the file
// for the specified image name
func ImageFile(name string) string {
	return fmt.Sprintf("%s/%s/img.data", GlobalImagePath, name)
}

// VolumeFile returns the path of the file
// for the specified volume name
func VolumeFile(name string) string {
	return fmt.Sprintf("%s/%s/volume.data", GlobalVolumePath, name)
}

// MachinePath returns the current folder
// for the specified machine name
func MachinePath(name string) string {
	return fmt.Sprintf("%s/%s", GlobalMachinePath, name)
}
