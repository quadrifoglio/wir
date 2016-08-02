package shared

import (
	"encoding/json"
	"os"
)

type APIConfigStruct struct {
	NodeID     byte // (0-255)
	AdminEmail string

	Address string
	Port    int

	StorageBackend string
	ZfsPool        string

	BridgeIface      string
	EnableNetMonitor bool
	PPSAlert         uint64
	PPSStop          uint64

	EnableKVM bool
	Ebtables  string `json:"EbtablesCommand"`
	QemuImg   string `json:"QemuImgCommand"`
	QemuNbd   string `json:"QemuNbdCommand"`
	Qemu      string `json:"QemuCommand"`
	Vzctl     string `json:"VzctlCommand"`

	DatabaseFile string
	ImagePath    string
	MachinePath  string
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
