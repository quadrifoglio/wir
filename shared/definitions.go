package shared

const (
	BackendKVM = "kvm"
	BackendLXC = "lxc"
)

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
	Type   string // Type of the image (kvm, lxc)
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

// VolumeDef is the data structure used in communications
// with all the Volume* HTTP handlers (/volumes)
type VolumeDef struct {
	ID   string // 64 bit random unique identifier
	Name string // Name of the volume
	Type string // Type of the volume (kvm, lxc)
	Size uint64 // Size of the volume in KiB
}

// InterfaceDef represents a network interface
// associated with a machine
type InterfaceDef struct {
	Network string // ID of the network to which the interface is attached
	MAC     string // MAC address of the interface
	IP      string // IP address of the interface in CIDR notation
}

// MachineDef is the data structure used in communications
// with all the Machine* HTTP handlers (/machines)
type MachineDef struct {
	ID     string // 64 bit random unique identifier
	Name   string // Name of the machine
	Image  string // ID of the machine's image
	Cores  int    // Number of CPUs
	Memory uint64 // Memory in MiB

	Volumes    []string       // IDs of the attached volumes
	Interfaces []InterfaceDef // List of network interfaces
}

// MachineStatusDef is the data structure used as a response
// to the MachineStatus HTTP handler (/machines/<id>/status)
type MachineStatusDef struct {
	Running bool // True if the machine is currently running
}
