package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/utils"
)

func validateMachine(req shared.MachineDef) (error, int) {
	if len(req.Name) == 0 {
		return fmt.Errorf("Missing 'Name'"), 400
	}

	if len(req.Image) > 0 {
		if !DBImageExists(req.Image) {
			return fmt.Errorf("Image not found"), 404
		}
	} else {
		return fmt.Errorf("Missing 'Image'"), 400
	}

	if req.Cores == 0 {
		return fmt.Errorf("'Cores' can't be 0"), 400
	}
	if req.Memory == 0 {
		return fmt.Errorf("'Memory' can't be 0"), 400
	}

	for _, v := range req.Volumes {
		if !DBVolumeExists(v) {
			return fmt.Errorf("Volume '%s' not found", v), 404
		}
	}

	for _, i := range req.Interfaces {
		if len(i.Network) == 0 {
			return fmt.Errorf("Missing 'Network' for interface"), 400
		}
		if !DBNetworkExists(i.Network) {
			return fmt.Errorf("Network '%s' not found", i.Network), 404
		}

		if len(i.MAC) > 0 {
			_, err := net.ParseMAC(i.MAC)
			if err != nil {
				return fmt.Errorf("Invalid 'MAC' for interface"), 400
			}

			if !DBIsMACFree(i.MAC) {
				return fmt.Errorf("MAC address is already in use"), 400
			}
		}
		if len(i.IP) > 0 {
			ip := net.ParseIP(i.IP)
			if ip == nil {
				return fmt.Errorf("Invalid 'IP' for interface"), 400
			}
		}
	}

	return nil, 200
}

func HandleMachineCreate(w http.ResponseWriter, r *http.Request) {
	var req shared.MachineDef

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ErrorResponse(w, r, err, 400)
		return
	}

	err, status := validateMachine(req)
	if err != nil {
		ErrorResponse(w, r, err, status)
		return
	}

	for {
		req.ID = utils.RandID(GlobalNodeID)
		if !DBMachineExists(req.ID) {
			break
		}
	}

	img, err := DBImageGet(req.Image)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	if img.Type == shared.BackendKVM {
		err := MachineKvmCreate(req, img)
		if err != nil {
			ErrorResponse(w, r, err, 500)
			return
		}
	} else {
		ErrorResponse(w, r, fmt.Errorf("Unsupported backend"), 400)
		return
	}

	err = DBMachineCreate(req)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, req)
}

func HandleMachineList(w http.ResponseWriter, r *http.Request) {
	machines, err := DBMachineList()
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, machines)
}

func HandleMachineGet(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	if !DBMachineExists(id) {
		ErrorResponse(w, r, fmt.Errorf("Machine not found"), 404)
		return
	}

	machine, err := DBMachineGet(id)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, machine)
}

func HandleMachineUpdate(w http.ResponseWriter, r *http.Request) {
	var req shared.MachineDef

	v := mux.Vars(r)
	id := v["id"]

	if !DBMachineExists(id) {
		ErrorResponse(w, r, fmt.Errorf("Machine not found"), 404)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ErrorResponse(w, r, err, 400)
		return
	}

	req.ID = id

	err, status := validateMachine(req)
	if err != nil {
		ErrorResponse(w, r, err, status)
		return
	}

	err = DBMachineUpdate(req)
	if err != nil {
		ErrorResponse(w, r, err, 404)
		return
	}

	SuccessResponse(w, r, req)
}

func HandleMachineDelete(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	if !DBMachineExists(id) {
		ErrorResponse(w, r, fmt.Errorf("Machine not found"), 404)
		return
	}

	err := DBMachineDelete(id)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, nil)
}

func HandleMachineStart(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	if !DBMachineExists(id) {
		ErrorResponse(w, r, fmt.Errorf("Machine not found"), 404)
		return
	}

	err := MachineKvmStart(id)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, nil)
}

func HandleMachineStop(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	if !DBMachineExists(id) {
		ErrorResponse(w, r, fmt.Errorf("Machine not found"), 404)
		return
	}

	err := MachineKvmStop(id)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, nil)
}
