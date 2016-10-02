package shared

type Network struct {
	Name    string
	Gateway string // Name of the physical interface serving as a gateway, if any
	Addr    string // Network address
	Mask    string // Netmask

	UseDHCP bool   // Enable internal DHCP for this network
	Router  string // DHCP: IP address of the router, if any
	StartIP string // DHCP: First IP to be leased
	NumIP   int    // DHCP: Number of IP to lease, starting from StartIP
}
