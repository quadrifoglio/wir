package shared

const (
	TypeUnknown = "unknown"
	TypeQemu    = "qemu"
	TypeLXC     = "lxc"
)

type Image struct {
	Name   string
	Type   string
	Source string

	// Optional information
	Desc          string
	MainPartition int
}
