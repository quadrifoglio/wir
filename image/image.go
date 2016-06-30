package image

const (
	TypeUnknown = "unknown"
	TypeQemu    = "qemu"
	TypeVz      = "openvz"
	TypeLXC     = "lxc"
)

type Image struct {
	Name   string
	Type   string
	Source string

	// Optional information
	Arch    string
	Distro  string
	Release string
}
