package api

import (
	"log"

	"github.com/quadrifoglio/wir/net"
	"github.com/quadrifoglio/wir/shared"
)

func StartNetworkDHCP(netw shared.Network) error {
	cb := func(mac, ip string) {
		m, err := DBGetMachineByMAC(mac)
		if err != nil {
			log.Println("DHCP callback: ", err)
			return
		}

		for i, iface := range m.ListInterfaces() {
			if iface.MAC == mac {
				iface.IP = ip

				_, err := m.UpdateInterface(i, iface)
				if err != nil {
					log.Println("DHCP callback: ", err)
					return
				}
			}
		}
	}

	// TODO: Do not hardcode parameters
	opts := net.DHCPOptions{"255.255.255.0", "192.168.137.1", "8.8.8.8", nil}
	serv := net.NewDHCPServer("192.168.137.10", "192.168.137.11", 10, opts, cb)

	return serv.Start(net.BridgeName(netw.Name))
}
