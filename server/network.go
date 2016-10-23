package server

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/quadrifoglio/go-dhcp"

	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/system"
	"github.com/quadrifoglio/wir/utils"
)

// StartNetworks is called when the daemon starts
// to initialize all the existing networks
func StartNetworks() error {
	nets, err := DBNetworkList()
	if err != nil {
		return err
	}

	for _, net := range nets {
		err := CreateNetwork(net)
		if err != nil {
			return err
		}
	}

	err = system.EbtablesSetup()
	if err != nil {
		return err
	}

	go StartNetworkDHCP()
	return nil
}

// CreateNetwork creates the specified network
// by using Linux bridges
func CreateNetwork(def shared.NetworkDef) error {
	if !system.InterfaceExists(NetworkNicName(def.ID)) {
		err := system.CreateBridge(NetworkNicName(def.ID))
		if err != nil {
			return err
		}
	}

	if len(def.GatewayIface) > 0 {
		err := system.SetInterfaceMaster(def.GatewayIface, NetworkNicName(def.ID))
		if err != nil {
			return err
		}
	}

	return nil
}

// AttachInterfaceToNetwork attaches the n-th interface of the specified machine ID
// to the specified network ID, and sets up the traffic controle rules
func AttachInterfaceToNetwork(machineId string, n int, def shared.InterfaceDef) error {
	err := system.EbtablesFlush(MachineNicName(machineId, n))
	if err != nil {
		return err
	}

	err = system.EbtablesAllowTraffic(MachineNicName(machineId, n), def.MAC, def.IP)
	if err != nil {
		return err
	}

	return system.SetInterfaceMaster(MachineNicName(machineId, n), NetworkNicName(def.Network))
}

// DeleteNetwork deletes the specified network
// and the associated bridge
func DeleteNetwork(id string) error {
	return system.DeleteInterface(NetworkNicName(id))
}

// StartNetworkDHCP starts an internal DHCP server to handle
// DHCP requests from machines attached to the DHCP-enabled networks
func StartNetworkDHCP() error {
	handler := func(s *dhcp.Server, msg dhcp.Message) {
		machine, err := DBMachineGetByMAC(msg.ClientMAC.String())
		if err != nil {
			log.Printf("DHCP: can't get machine with mac address '%s': %s\n", msg.ClientMAC, err)
			return
		}

		for _, nic := range machine.Interfaces {
			if len(nic.Network) == 0 {
				continue
			}
			if nic.MAC != msg.ClientMAC.String() {
				continue
			}

			netw, err := DBNetworkGet(nic.Network)
			if err != nil {
				log.Printf("DHCP: can't get network '%s': %s\n", nic.Network, err)
				return
			}

			if !netw.DHCP.Enabled {
				continue
			}

			_, netAddr, err := net.ParseCIDR(netw.CIDR)
			if err != nil {
				log.Printf("DHCP: invalid network address '%s' for network '%s'\n", netw.CIDR, netw.ID)
				return
			}

			srv := []byte{0, 0, 0, 0}
			leaseTime := make([]byte, 4)
			binary.BigEndian.PutUint32(leaseTime, 86400) // 1 day lease

			var t byte
			if msg.Type == dhcp.DHCPTypeDiscover {
				t = dhcp.DHCPTypeOffer
			} else if msg.Type == dhcp.DHCPTypeRequest {
				t = dhcp.DHCPTypeACK
			}

			response := dhcp.NewMessage(t, msg.TransactionID, srv, net.ParseIP(nic.IP).To4(), msg.ClientMAC)
			response.SetOption(dhcp.OptionSubnetMask, netAddr.Mask)
			response.SetOption(dhcp.OptionRouter, net.ParseIP(netw.DHCP.Router).To4())
			response.SetOption(dhcp.OptionServerIdentifier, srv)
			response.SetOption(dhcp.OptionIPAddressLeaseTime, leaseTime)

			s.BroadcastPacket(response.GetFrame())
			break
		}
	}

	server, err := dhcp.NewServer()
	if err != nil {
		return err
	}

	server.HandleFunc(handler)
	return server.ListenAndServe()
}

// NetworkFreeLease returns the first available IP address
// in the specified network
func NetworkFreeLease(netw shared.NetworkDef) (net.IP, error) {
	var ip net.IP

	ms, err := DBMachineListOnNetwork(netw.ID)
	if err != nil {
		return ip, err

	}

	// Pack all the IPs in use on the network into an array
	ips := make([]string, 0)
	for _, m := range ms {
		for _, i := range m.Interfaces {
			if i.Network == netw.ID {
				ips = append(ips, i.IP)
			}
		}
	}

	// Parse the CIDR to get the network address and mask
	_, netAddr, err := net.ParseCIDR(netw.CIDR)
	if err != nil {
		return ip, err
	}

	ip = net.ParseIP(netw.DHCP.StartIP).To4()
	ip.Mask(netAddr.Mask)

	// Try to find an available IP
	for i := 0; i < netw.DHCP.NumIP; i++ {
		if !utils.SliceContainsStr(ip.To4().String(), ips) {
			return ip.To4(), nil
		}

		utils.IncrementIP(ip)
	}

	return ip, fmt.Errorf("no lease available")
}

// NetworkNicName returns the bridge interface name
// coresponding to the specified network ID
func NetworkNicName(id string) string {
	return fmt.Sprintf("net%s%s", strings.ToUpper(id[:1]), id[1:])
}

// MachineNicName returns the interface name coresponding
// to the n-th interface of the specified machine ID
func MachineNicName(id string, n int) string {
	return fmt.Sprintf("nic%s%s.%d", strings.ToUpper(id[:1]), id[1:], n)
}
