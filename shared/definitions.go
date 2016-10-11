package shared

// IndexDef represents the data returned by
// the Index HTTP handler
type IndexDef struct {
	Hostname    string  // System hostname
	CpuUsage    float32 // Current CPU Usage in percent
	MemoryUsage uint64  // Currently used memory in KiB
	MemoryTotal uint64  // Total memory available to the system in KiB
}
