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

	err := InitDatabase(db)
	if err != nil {
		return err
	}

	err = StartNetworks()
	if err != nil {
		return err
	}

	return nil
}

// ImageFile returns the path of the file
// for the specified image name
func ImageFile(id string) string {
	return fmt.Sprintf("%s/%s/img.data", GlobalImagePath, id)
}

// VolumeFile returns the path of the file
// for the specified volume name
func VolumeFile(id string) string {
	return fmt.Sprintf("%s/%s/volume.data", GlobalVolumePath, id)
}

// MachinePath returns the current folder
// for the specified machine name
func MachinePath(id string) string {
	return fmt.Sprintf("%s/%s", GlobalMachinePath, id)
}

// MachineDisk returns the path of the disk file
// for the specified machine name
func MachineDisk(id string) string {
	return fmt.Sprintf("%s/disk.data", MachinePath(id))
}
