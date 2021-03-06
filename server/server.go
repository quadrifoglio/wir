package server

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/quadrifoglio/wir/utils"
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

	if !utils.FileExists(filepath.Dir(db)) {
		err := os.MkdirAll(filepath.Dir(db), 0755)
		if err != nil {
			return err
		}
	}

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

// MachineMonitorPath returns the path to the
// machine's monitor device
func MachineMonitorPath(id string) string {
	return fmt.Sprintf("%s/monitor.sock", MachinePath(id))
}
