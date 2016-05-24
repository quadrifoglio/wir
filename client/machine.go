package client

type MachineRequest struct {
	Name        string
	Image       string
	Cores       int
	Memory      int
	NetBridgeOn string // Interface to bridge on
}
