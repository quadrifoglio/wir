package shared

// IndexDef represents the data returned by
// the Index HTTP handler (/)
type IndexDef struct {
	Hostname    string  // System hostname
	CpuUsage    float32 // Current CPU Usage in percent
	MemoryUsage uint64  // Currently used memory in KiB
	MemoryTotal uint64  // Total memory available to the system in KiB
}

// ImageDef is the data structure used in communications
// with all the Image* HTTP handlers (/images)
type ImageDef struct {
	ID     string // 64 bit random unique identifier
	Name   string // Name of the image
	Source string // Location of the image file (scheme://[user@]host/path or just file path)
}

// NetworkDef is the data structure used in communications
// with all the Network* HTTP handlers (/networks)
type NetworkDef struct {
	ID           string // 64 bit random unique identifier
	Name         string // Name of the image
	CIDR         string // CIDR notation of network address (a.b.c.d/mask)
	GatewayIface string // Name of a physical interface that should be part of the network (optional)

	DHCP struct {
		Enabled bool   // Wether internal DHCP is in use on this network
		StartIP string // First IP to be leased
		NumIP   int    // Number of IP addresses to lease, starting from StartIP
		Router  string // IP address of the network router
	}
}
