package net

import (
	"github.com/quadrifoglio/wir/shared"
)

func SetupInterface(iface *shared.NetDev) error {
	var err error

	if len(iface.Mode) > 0 {
		if len(iface.MAC) == 0 {
			iface.MAC, err = GenerateMAC(shared.APIConfig.NodeID)
			if err != nil {
				return err
			}
		}

		err = GrantTraffic(iface.MAC, "0.0.0.0")
		if err != nil {
			return err
		}

		if len(iface.IP) > 0 {
			err := GrantTraffic(iface.MAC, iface.IP)
			if err != nil {
				return err
			}
		}
	}

	return err
}

func CheckInterface(iface shared.NetDev) error {
	if iface.Mode == shared.NetworkModeNone || len(iface.MAC) == 0 {
		return nil
	}

	is, err := IsGranted(iface.MAC, "0.0.0.0")
	if err != nil {
		return err
	}

	if !is {
		err := GrantTraffic(iface.MAC, "0.0.0.0")
		if err != nil {
			return err
		}
	}

	if len(iface.IP) > 0 {
		is, err = IsGranted(iface.MAC, iface.IP)
		if err != nil {
			return err
		}

		if !is {
			err := GrantTraffic(iface.MAC, iface.IP)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func DeleteInterface(iface shared.NetDev) error {
	if len(iface.MAC) > 0 {
		err := DenyTraffic(iface.MAC, "0.0.0.0")
		if err != nil {
			return err
		}

		if len(iface.IP) > 0 {
			err := DenyTraffic(iface.MAC, iface.IP)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
