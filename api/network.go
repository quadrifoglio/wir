package api

import (
	"encoding/binary"
	"log"
	"net"

	"github.com/quadrifoglio/go-dhcp"
)

func InitNetworks() error {
	server, err := dhcp.NewServer()
	if err != nil {
		return err
	}

	server.HandleDiscover(HandleDHCPDiscover)
	server.HandleRequest(HandleDHCPRequest)

	return server.ListenAndServe()
}

func HandleDHCPDiscover(s *dhcp.Server, id uint32, mac net.HardwareAddr) {
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

		message := dhcp.NewMessage(dhcp.DHCPTypeOffer, id, srv, net.ParseIP(nic.IP), mac)
		message.SetOption(dhcp.OptionSubnetMask, net.CIDRMask(netw.Mask, 32))
		message.SetOption(dhcp.OptionRouter, net.ParseIP(netw.Router))
		message.SetOption(dhcp.OptionServerIdentifier, srv)
		message.SetOption(dhcp.OptionIPAddressLeaseTime, leaseTime)
	}
}

func HandleDHCPRequest(s *dhcp.Server, id uint32, mac net.HardwareAddr, requestedIp net.IP) {
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
		if !net.ParseIP(nic.IP).Equal(requestedIp) {
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

		message := dhcp.NewMessage(dhcp.DHCPTypeACK, id, srv, net.ParseIP(nic.IP), mac)
		message.SetOption(dhcp.OptionSubnetMask, net.CIDRMask(netw.Mask, 32))
		message.SetOption(dhcp.OptionRouter, net.ParseIP(netw.Router))
		message.SetOption(dhcp.OptionServerIdentifier, srv)
		message.SetOption(dhcp.OptionIPAddressLeaseTime, leaseTime)
	}
}
