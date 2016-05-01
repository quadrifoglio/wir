package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
)

const (
	BackendQemu   = "qemu"
	BackendOpenVz = "openvz"

	DriveHDD   = "hdd"
	DriveCDROM = "cdrom"

	VmStateDown = 0
	VmStateUp   = 1

	VmNetworkNAT    = "nat"
	VmNetworkBridge = "bridge"
)

type VmBackend string
type VmDriveType string

type VmDrive struct {
	Type string `json:"type"`
	File string `json:"file"`
}

type VmParams struct {
	Migration   bool   `json:"migration"` // True if the VM is being migrated
	Backend     string `json:"backend"`
	Cores       int    `json:"cores"`
	Memory      int    `json:"memory"`
	ImageID     int    `json:"image_id"`
	NetBridgeOn string `json:"net_bridge_on"`
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

func (vm *Vm) Migrate(dst string) (int, error) {
	img, err := DatabaseGetImage(vm.Params.ImageID)
	if err != nil {
		return 0, err
	}

	if img.State != ImgStateAvailable {
		return 0, fmt.Errorf("Can not migrate non-available image")
	}

	img.Path = fmt.Sprintf("scp://%s@%s:%s", Config.User, Config.Address, img.Path)

	data, err := json.Marshal(img)
	if err != nil {
		return 0, fmt.Errorf("Can not encode image: %s", err)
	}

	resp, err := http.Post(fmt.Sprintf("http://%s/image/create", dst), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return 0, fmt.Errorf("POST %s/image/create: %s", dst, err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Cat not create new image (http %d)", resp.StatusCode)
	}

	var newImg Image
	err = json.NewDecoder(resp.Body).Decode(&newImg)
	if err != nil {
		return 0, fmt.Errorf("Cat not decode migrated image: %s", err)
	}

	vm.Params.Migration = true
	vm.Params.ImageID = newImg.ID

	data, err = json.Marshal(vm.Params)
	if err != nil {
		return 0, fmt.Errorf("Can not encode vm: %s", err)
	}

	resp, err = http.Post(fmt.Sprintf("http://%s/vm/create", dst), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return 0, fmt.Errorf("POST %s/vm/create: %s", dst, err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Cat not create new vm (http %d)", resp.StatusCode)
	}

	var newVm Vm
	err = json.NewDecoder(resp.Body).Decode(&newVm)
	if err != nil {
		return 0, fmt.Errorf("Can not decode new vm: %s", err)
	}

	for _, d := range vm.Drives {
		path := fmt.Sprintf("%s/%d/%s", Config.DrivesDir, newVm.ID, filepath.Base(d.File))
		cmd := exec.Command("scp", "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", d.File, path)

		err = cmd.Run()
		if err != nil {
			return 0, fmt.Errorf("Can not scp vm drive: %s", err)
		}
	}

	// TODO: Delete local VM from database and remove local files
	return newVm.ID, nil
}
