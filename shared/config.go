package shared

import (
	"encoding/json"
	"fmt"
	"os"
)

type APIConfigStruct struct {
	NodeID     byte // (0-255)
	AdminEmail string

	Address string
	Port    int

	EnableKVM      bool
	StorageBackend string
	ZfsPool        string

	BridgeIface      string
	EnableNetMonitor bool
	PPSAlert         uint64
	PPSStop          uint64

	DatabaseFile  string
	ImagePath     string
	MigrationPath string
	MachinePath   string
}

var (
	APIConfig APIConfigStruct
)

func ReadAPIConfig(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	err = json.NewDecoder(f).Decode(&APIConfig)
	if err != nil {
		return err
	}

	return nil
}

func MachinePath(backend string) string {
	return fmt.Sprintf("%s/%s", APIConfig.MachinePath, backend)
}

func IsStorage(s string) bool {
	return APIConfig.StorageBackend == s
}
