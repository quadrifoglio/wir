package shared

const (
	BackendKVM = "kvm"
	BackendLXC = "lxc"
)

// RemoteDef represents an API server
// It is used in the 'client' package
type RemoteDef struct {
	Host string // IP address or host name of the API server
	Port int    // TCP port
}

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
		StartIP string `json:",omitempty"` // First IP to be leased
		NumIP   int    `json:",omitempty"` // Number of IP addresses to lease, starting from StartIP
		Router  string `json:",omitempty"` // IP address of the network router
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

// MachineFetchDef represents a machine fetch request sent to an API server
// It contains all the information needed to retreive the machine
// from the specified server
type MachineFetchDef struct {
	Remote     RemoteDef // The API server from which the machine should be fetched
	ID         string    // ID of the machine to be fetched
	KeepRemote bool      // Wether the distant machine should be kept on the remote
}

// MachineDef is the data structure used in communications
// with all the Machine* HTTP handlers (/machines)
type MachineDef struct {
	ID     string // 64 bit random unique identifier
	Name   string // Name of the machine
	Image  string // ID of the machine's image
	Cores  int    // Number of CPUs
	Memory uint64 // Memory in MiB
	Disk   uint64 // Disk size in bytes

	Volumes    []string       // IDs of the attached volumes
	Interfaces []InterfaceDef // List of network interfaces
}

// MachineStatusDef is the data structure used as a response
// to the MachineStatus HTTP handler (/machines/<id>/status)
type MachineStatusDef struct {
	Running   bool    // True if the machine is currently running
	CpuUsage  float32 // Percentage of the time the CPU is busy
	RamUsage  uint64  // Currently used RAM in MiB
	DiskUsage uint64  // Current size of the disk image in bytes
}

// KvmOptsDef is the data structure used to represent
// KVM-specific options in the machine HTTP handler (/machines/<id>/kvm)
type KvmOptsDef struct {
	PID   int    // The QEMU/KVM proccess ID
	CDRom string // Path to a disk image to insert into the machine as a CD-ROM

	VNC struct {
		Enabled bool   // Wether to use the VNC server
		Address string // Bind address of the VNC server (ip:port)
		Port    int    // Port number

		// TODO: Add parapeters (ssl, authentication...)
	}

	Linux struct {
		Hostname     string // Linux hostname
		RootPassword string // Linux root password in clear text
	}
}

// CheckpointDef is the data structure used in transactions with
// the checkpoint HTTP handler (/machines/<id>/checkpoints)
type CheckpointDef struct {
	Name      string // Name of the checkpoint
	Timestamp int64  // Timestamp of the checkpoint
}
