package main

const (
	BackendQemu   = "qemu"
	BackendOpenVz = "openvz"
)

type Backend string

type VmParams struct {
	Backend Backend `json:"backend"`
}

type Vm struct {
	ID     int `json:"id"`
	Params VmParams
}

func VmGetAll() ([]Vm, error) {
	return nil, nil
}

func VmGet(id int) (Vm, error) {
	var vm Vm
	vm.ID = id

	return vm, nil
}

func VmCreate(p *VmParams) (Vm, error) {
	var vm Vm
	vm.Params.Backend = p.Backend

	return vm, nil
}

func (vm *Vm) Start() error {
	switch vm.Params.Backend {
	case BackendQemu:
		return QemuStart(vm)
	case BackendOpenVz:
		return OpenVzStart(vm)
	}

	return ErrInvalidBackend
}

func (vm *Vm) Stop() error {
	switch vm.Params.Backend {
	case BackendQemu:
		return QemuStop(vm)
	case BackendOpenVz:
		return OpenVzStop(vm)
	}

	return ErrInvalidBackend
}
