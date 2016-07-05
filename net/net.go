package net

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
