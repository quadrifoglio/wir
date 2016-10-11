package shared

// IndexDef represents the data returned by
// the Index HTTP handler
type IndexDef struct {
	Hostname string  // System hostname
	CpuUsage float32 // Current CPU Usage in percent
}
