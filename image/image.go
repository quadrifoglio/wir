package image

const (
	TypeUnknown = 0
	TypeQemu    = 1
	TypeDocker  = 2
	TypeVz      = 3
)

func TypeExists(t int) bool {
	return t > 0 && t <= 3
}
