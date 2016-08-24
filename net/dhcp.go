package net

import (
	"math/rand"
	"net"
	"time"

	dhcp "github.com/krolaw/dhcp4"
)

type DHCPLease struct {
	NIC    string
	Expiry time.Time
}

type DHCPOptions struct {
	Mask   string
	Router string
	DNS    string

	StaticAssocs map[string]string // MAC to IP
}

type DHCPServer struct {
	IP         net.IP
	StartIP    net.IP
	LeaseRange int

	StaticAssocs map[string]string

	LeaseDuration time.Duration
	Leases        map[int]DHCPLease

	Options dhcp.Options

	Callback DHCPCallback
}

type DHCPCallback func(string, string)

func NewDHCPServer(addr string, startIp string, leaseRange int, opts DHCPOptions, cb DHCPCallback) DHCPServer {
	var serv DHCPServer
	serv.IP = net.ParseIP(addr)
	serv.StartIP = net.ParseIP(startIp)
	serv.LeaseRange = leaseRange
	serv.LeaseDuration = 2 * time.Hour
	serv.Leases = make(map[int]DHCPLease, 10)
	serv.StaticAssocs = opts.StaticAssocs

	serv.Options = dhcp.Options{
		dhcp.OptionSubnetMask:       net.IPMask(net.ParseIP(opts.Mask)),
		dhcp.OptionRouter:           []byte(net.ParseIP(opts.Router)),
		dhcp.OptionDomainNameServer: []byte(net.ParseIP(opts.DNS)),
	}

	serv.Callback = cb

	return serv
}

func (s *DHCPServer) Start(iface string) error {
	return dhcp.ListenAndServeIf(iface, s)
}

func (s *DHCPServer) ServeDHCP(p dhcp.Packet, typ dhcp.MessageType, options dhcp.Options) dhcp.Packet {
	if typ == dhcp.Discover {
		var ip net.IP

		nic := p.CHAddr().String()

		if v, ok := s.StaticAssocs[nic]; ok {
			ip = net.ParseIP(v)
		} else {
			leaseId := s.previousLeaseID(nic)

			if leaseId == -1 { // If there is no previous lease for that nic
				leaseId = s.freeLeaseID()
			}

			ip = dhcp.IPAdd(s.StartIP, leaseId)
		}

		return dhcp.ReplyPacket(p, dhcp.Offer, s.IP, ip, s.LeaseDuration,
			s.Options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
	}

	if typ == dhcp.Request {
		if server, ok := options[dhcp.OptionServerIdentifier]; ok && !net.IP(server).Equal(s.IP) {
			return nil // Message not for this dhcp server
		}

		reqIP := net.IP(options[dhcp.OptionRequestedIPAddress])
		if reqIP == nil {
			reqIP = net.IP(p.CIAddr())
		}

		if len(reqIP) == 4 && !reqIP.Equal(net.IPv4zero) {
			if leaseNum := dhcp.IPRange(s.StartIP, reqIP) - 1; leaseNum >= 0 && leaseNum < s.LeaseRange {
				if l, exists := s.Leases[leaseNum]; !exists || l.NIC == p.CHAddr().String() {
					s.Leases[leaseNum] = DHCPLease{p.CHAddr().String(), time.Now().Add(s.LeaseDuration)}

					s.Callback(p.CHAddr().String(), reqIP.String())

					return dhcp.ReplyPacket(p, dhcp.ACK, s.IP, reqIP, s.LeaseDuration,
						s.Options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
				}
			}
		}

		return dhcp.ReplyPacket(p, dhcp.NAK, s.IP, nil, 0, nil)
	}

	if typ == dhcp.Release || typ == dhcp.Decline {
		nic := p.CHAddr().String()

		for i, v := range s.Leases {
			if v.NIC == nic {
				delete(s.Leases, i)
				break
			}
		}
	}

	return nil
}

func (s *DHCPServer) previousLeaseID(nic string) int {
	var index int = -1

	for i, v := range s.Leases {
		if v.NIC == nic {
			index = i
			break
		}
	}

	return index
}

func (s *DHCPServer) freeLeaseID() int {
	now := time.Now()
	b := rand.Intn(s.LeaseRange)

	for _, v := range [][]int{[]int{b, s.LeaseRange}, []int{0, b}} {
		for i := v[0]; i < v[1]; i++ {
			if l, ok := s.Leases[i]; !ok || l.Expiry.Before(now) {
				return i
			}
		}
	}

	return -1
}
