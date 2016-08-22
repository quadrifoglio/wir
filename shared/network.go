package shared

type Network struct {
	Name    string
	Bridge  string // Name of the underlying bridge interface
	Gateway string // Name of the physical interface serving as a gateway, if any
}
