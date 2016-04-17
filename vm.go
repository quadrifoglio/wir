package main

import (
	"log"
)

const (
	BackendQemu   = "qemu"
	BackendOpenVz = "openvz"

	DriveHDD   = "hdd"
	DriveCDROM = "cdrom"

	VmStateDown = 0
	VmStateUp   = 1
)

type VmBackend string
type VmDriveType string

type VmDrive struct {
	Type string `json:"type"`
	File string `json:"file"`
}

type VmParams struct {
	Backend string `json:"backend"`
	Cores   int    `json:"cores"`
	Memory  int    `json:"memory"`
	ImageID int    `json:"image_id"`
}

type VmState int

type Vm struct {
	ID     int               `json:"id"`
	State  VmState           `json:"state"`
	Params VmParams          `json:"params"`
	Drives []VmDrive         `json:"drives"`
	Attrs  map[string]string `json:"-"`
}

func VmGetAll() ([]Vm, error) {
	vms, err := DatabaseListVms()
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		err = vm.Status()
		if err != nil {
			return nil, err
		}
	}

	return vms, nil
}

func VmGet(id int) (Vm, error) {
	vm, err := DatabaseGetVmByID(id)
	if err != nil {
		log.Println(err)
		return vm, ErrVmNotFound
	}

	err = vm.Status()
	if err != nil {
		return vm, err
	}

	return vm, nil
}

func VmCreate(p *VmParams) (Vm, error) {
	var err error
	var vm Vm
	vm.Params = *p

	switch vm.Params.Backend {
	case BackendQemu:
		err = QemuSetupImage(&vm)
		break
	case BackendOpenVz:
		err = OpenVzSetupImage(&vm)
		break
	}

	if err != nil {
		return vm, err
	}

	return vm, DatabaseInsertVm(&vm)
}

func (vm *Vm) Status() error {
	switch vm.Params.Backend {
	case BackendQemu:
		return QemuStatus(vm)
	case BackendOpenVz:
		return OpenVzStatus(vm)
	}

	return ErrInvalidBackend
}

func (vm *Vm) Start() error {
	if vm.State == VmStateUp {
		return ErrRunning
	}

	switch vm.Params.Backend {
	case BackendQemu:
		return QemuStart(vm)
	case BackendOpenVz:
		return OpenVzStart(vm)
	}

	return ErrInvalidBackend
}

func (vm *Vm) Stop() error {
	if vm.State != VmStateUp {
		return ErrNotRunning
	}

	switch vm.Params.Backend {
	case BackendQemu:
		return QemuStop(vm)
	case BackendOpenVz:
		return OpenVzStop(vm)
	}

	return ErrInvalidBackend
}
