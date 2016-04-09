package main

const (
	BackendQemu   = "qemu"
	BackendOpenVz = "openvz"

	StateDown = 0
	StateUp   = 1
)

type VmBackend string

type VmParams struct {
	Backend string `json:"backend"`
	Cores   int    `json:"cores"`
	Memory  int    `json:"memory"`
}

type VmState int

type Vm struct {
	ID     int               `json:"id"`
	Params VmParams          `json:"params"`
	State  VmState           `json:"state"`
	Attrs  map[string]string `json:"-"`
}

func VmGetAll() ([]Vm, error) {
	vms, err := DatabaseListVms()
	if err != nil {
		return nil, err
	}

	return vms, nil
}

func VmGet(id int) (Vm, error) {
	return DatabaseGetVmByID(id)
}

func VmCreate(p *VmParams) (Vm, error) {
	var vm Vm
	vm.Params = *p

	return vm, DatabaseInsertVm(&vm)
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
