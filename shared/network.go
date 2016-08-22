package shared

type Network struct {
	Name string

	Gateway string // Name of the physical interface serving as a gateway, if any
	Router  string // IP address of the router, if any
	Addr    string // Network address
	Mask    int    `json:",omitempty"` // Netmask
}
