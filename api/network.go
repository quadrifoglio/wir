package api

import (
	"encoding/binary"
	"fmt"
	"log"
	gonet "net"

	"github.com/quadrifoglio/go-dhcp"

	"github.com/quadrifoglio/wir/net"
	"github.com/quadrifoglio/wir/shared"
)

func InitNetworks() {
	nets, err := DBListNetworks()
	if err != nil {
		log.Println("init networks:", err)
		return
	}

	for _, netw := range nets {
		err := StartNetwork(netw)
		if err != nil {
			log.Println("init networks:", err)
		}
	}

	server, err := dhcp.NewServer()
	if err != nil {
		log.Println("init networks:", err)
		return
	}

	server.HandleDiscover(HandleDHCPDiscover)
	server.HandleRequest(HandleDHCPRequest)

	err = server.ListenAndServe()
	if err != nil {
		log.Println("init networks:", err)
		return
	}
}

func StartNetwork(netw shared.Network) error {
	br := net.BridgeName(netw.Name)

	if !net.BridgeExists(br) {
		err := net.CreateBridge(br)
		if err != nil {
			return fmt.Errorf("start network: %s", err)
		}

		if len(netw.Gateway) > 0 {
			err = net.AddBridgeIf(br, netw.Gateway)
			if err != nil {
				return fmt.Errorf("start network: %s", err)
			}
		}
	}

	return nil
}

func HandleDHCPDiscover(s *dhcp.Server, id uint32, mac gonet.HardwareAddr) {
	m, err := DBGetMachineByMAC(mac.String())
	if err != nil {
		log.Println("dhcp discover: no nic with address", mac)
		return
	}

	for _, nic := range m.ListInterfaces() {
		if len(nic.Network) == 0 {
			continue
		}
		if nic.MAC != mac.String() {
			continue
		}
		if len(nic.IP) == 0 {
			// TODO: Pick an IP address from pool
			log.Println("dhcp discover: no ip assigned to nic")
			return
		}

		netw, err := DBGetNetwork(nic.Network)
		if err != nil {
			log.Println("dhcp discover: network", nic.Network, "not found")
			return
		}

		srv := []byte{0, 0, 0, 0}
		leaseTime := make([]byte, 4)
		binary.BigEndian.PutUint32(leaseTime, 86400) // 1 day lease

		message := dhcp.NewMessage(dhcp.DHCPTypeOffer, id, srv, gonet.ParseIP(nic.IP).To4(), mac)
		message.SetOption(dhcp.OptionSubnetMask, net.ParseMask(netw.Mask))
		message.SetOption(dhcp.OptionRouter, gonet.ParseIP(netw.Router).To4())
		message.SetOption(dhcp.OptionServerIdentifier, srv)
		message.SetOption(dhcp.OptionIPAddressLeaseTime, leaseTime)

		s.BroadcastPacket(message.GetFrame())
	}
}

func HandleDHCPRequest(s *dhcp.Server, id uint32, mac gonet.HardwareAddr, requestedIp gonet.IP) {
	m, err := DBGetMachineByMAC(mac.String())
	if err != nil {
		log.Println("dhcp request: no nic with address", mac)
		return
	}

	for _, nic := range m.ListInterfaces() {
		if len(nic.Network) == 0 {
			continue
		}
		if nic.MAC != mac.String() {
			continue
		}
		if len(nic.IP) == 0 {
			// TODO: Pick an IP address from pool
			log.Println("dhcp request: no ip assigned to nic")
			return
		}
		if !gonet.ParseIP(nic.IP).Equal(requestedIp) {
			// TODO: Send DHCP NACK
			log.Println("dhcp request: invalid requested ip")
			return
		}

		netw, err := DBGetNetwork(nic.Network)
		if err != nil {
			log.Println("dhcp request: network", nic.Network, "not found")
			return
		}

		srv := []byte{0, 0, 0, 0}
		leaseTime := make([]byte, 4)
		binary.BigEndian.PutUint32(leaseTime, 86400) // 1 day lease

		message := dhcp.NewMessage(dhcp.DHCPTypeACK, id, srv, gonet.ParseIP(nic.IP).To4(), mac)
		message.SetOption(dhcp.OptionSubnetMask, net.ParseMask(netw.Mask))
		message.SetOption(dhcp.OptionRouter, gonet.ParseIP(netw.Router).To4())
		message.SetOption(dhcp.OptionServerIdentifier, srv)
		message.SetOption(dhcp.OptionIPAddressLeaseTime, leaseTime)

		s.BroadcastPacket(message.GetFrame())
	}
}
