package net

import (
	"crypto/rand"
	"fmt"
)

func Init(ebtables, bridgeIf string) error {
	err := CreateBridge("wir0")
	if err != nil {
		return err
	}

	err = BridgeAddIf("wir0", bridgeIf)
	if err != nil {
		return err
	}

	err = InitEbtables(ebtables)
	if err != nil {
		return err
	}

	return nil
}

func GenerateMAC() (string, error) {
	buf := make([]byte, 6)

	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}

	buf[0] |= 2
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x\n", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5]), nil
}
