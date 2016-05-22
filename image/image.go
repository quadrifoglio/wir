package image

const (
	TypeUnknown = 0
	TypeQemu    = 1
	TypeDocker  = 2
	TypeVz      = 3
)

type Image struct {
	Name string
	Type int
	Path string
}
