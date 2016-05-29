package image

const (
	TypeUnknown = 0
	TypeQemu    = 1
	TypeDocker  = 2
	TypeVz      = 3
)

type Image struct {
	Name   string
	Type   int
	Source string
}

func TypeToString(t int) string {
	switch t {
	case TypeQemu:
		return "qemu"
	case TypeDocker:
		return "docker"
	case TypeVz:
		return "openvz"
	default:
		return "unknown"
	}
}

func StringToType(t string) int {
	switch t {
	case "qemu":
		return TypeQemu
	case "docker":
		return TypeDocker
	case "openvz":
		return TypeVz
	default:
		return TypeUnknown
	}
}
