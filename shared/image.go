package shared

const (
	TypeUnknown = "unknown"
	TypeQemu    = "qemu"
	TypeLXC     = "lxc"
)

type ImageInfo struct {
	Name   string
	Type   string
	Source string

	// Optional information
	Arch    string
	Distro  string
	Release string

	MainPartition int
}
