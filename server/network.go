package server

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/rs/xid"

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
	if len(def.Name) > 12 {
		return fmt.Errorf("Network name must be less than 12 characters long")
	}

	if !system.InterfaceExists(NetworkNicName(def.Name)) {
		err := system.CreateBridge(NetworkNicName(def.Name))
		if err != nil {
			return err
		}
	}

	if len(def.GatewayIface) > 0 {
		err := system.SetInterfaceMaster(def.GatewayIface, NetworkNicName(def.Name))
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

	err = system.SetInterfaceMaster(MachineNicName(machineId, n), NetworkNicName(def.Network))
	if err != nil {
		return err
	}

	go func() {
		pps, err := system.GetInterfacePPS(MachineNicName(machineId, n), "tx")
		if err != nil {
			log.Printf("Monitor failed for machine %s: %s\n", machineId, err)
			return
		}

		if pps > 90000 {
			log.Printf("Monitor DDOS alert for machine %s: shutting down interface\n", machineId)

			err := system.DownInterface(MachineNicName(machineId, n))
			if err != nil {
				log.Println("Monitor failed to shut down interface after DDOS alter. Machine", machineId)
			}

			return
		}

		time.Sleep(10 * time.Second)
	}()

	return nil
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
			// Not loging: too much messages
			//log.Printf("DHCP: can't get machine with mac address '%s': %s\n", msg.ClientMAC, err)
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
				log.Printf("DHCP: invalid network address '%s' for network '%s'\n", netw.CIDR, netw.Name)
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

	ms, err := DBMachineListOnNetwork(netw.Name)
	if err != nil {
		return ip, err

	}

	// Pack all the IPs in use on the network into an array
	ips := make([]string, 0)
	for _, m := range ms {
		for _, i := range m.Interfaces {
			if i.Network == netw.Name {
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

	return ip, fmt.Errorf("No lease available")
}

// NetworkNicName returns the bridge interface name
// coresponding to the specified network name
func NetworkNicName(name string) string {
	return fmt.Sprintf("net%s%s", strings.ToUpper(name[:1]), name[1:])
}

// MachineNicName returns the interface name coresponding
// to the n-th interface of the specified machine ID
func MachineNicName(id string, n int) string {
	uid, err := xid.FromString(id)
	if err != nil {
		return ""
	}

	c := make([]byte, 4)
	t := make([]byte, 8)

	binary.LittleEndian.PutUint32(c, uint32(uid.Counter()))
	binary.LittleEndian.PutUint64(t, uint64(uid.Time().Unix()))

	// The resulting ID is 'nic'
	// + 2  bytes of the original counter (32bits)
	// + 2  bytes of the original timestamp (64bits)
	// + .n the interface index

	return fmt.Sprintf("nic%s%s.%d", hex.EncodeToString(c[:2]), hex.EncodeToString(t[:2]), n)
}
