package config

import (
	"encoding/json"
	"os"
)

type APIConfig struct {
	NodeID      byte // (0-255)
	Address     string
	Port        int
	BridgeIface string

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
	API APIConfig
)

func ReadAPIConfig(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	err = json.NewDecoder(f).Decode(&API)
	if err != nil {
		return err
	}

	return nil
}
