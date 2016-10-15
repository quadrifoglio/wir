package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/system"
	"github.com/quadrifoglio/wir/utils"
)

// validateMachine validates the specified machine definition,
// modifying it if need be, and returns the coresponding http status code
func validateMachine(req *shared.MachineDef) (error, int) {
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

	for i, iface := range req.Interfaces {
		if len(iface.Network) == 0 {
			return fmt.Errorf("Missing 'Network' for interface"), 400
		}
		if !DBNetworkExists(iface.Network) {
			return fmt.Errorf("Network '%s' not found", iface.Network), 404
		}

		if len(iface.MAC) > 0 {
			_, err := net.ParseMAC(iface.MAC)
			if err != nil {
				return fmt.Errorf("Invalid 'MAC' for interface"), 400
			}

			if !DBIsMACFree(iface.MAC) {
				return fmt.Errorf("MAC address is already in use"), 400
			}
		} else {
			for {
				mac, err := utils.RandMAC(GlobalNodeID)
				if err != nil {
					return err, 500
				}

				req.Interfaces[i].MAC = mac

				if DBIsMACFree(mac) {
					break
				}
			}
		}

		if len(iface.IP) > 0 {
			ip := net.ParseIP(iface.IP)
			if ip == nil {
				return fmt.Errorf("Invalid 'IP' for interface"), 400
			}
		}
	}

	return nil, 200
}

// POST /machines
func HandleMachineCreate(w http.ResponseWriter, r *http.Request) {
	var req shared.MachineDef

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ErrorResponse(w, r, err, 400)
		return
	}

	err, status := validateMachine(&req)
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

// GET /machines
func HandleMachineList(w http.ResponseWriter, r *http.Request) {
	machines, err := DBMachineList()
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, machines)
}

// GET /machines/<id>
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

// POST /machines/<id>
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

	err, status := validateMachine(&req)
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

// DELET /machines/<id>
func HandleMachineDelete(w http.ResponseWriter, r *http.Request) {
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

	err = os.RemoveAll(MachinePath(id))
	if err != nil {
		log.Printf("Could not delete machine '%s' files: %s\n", id, err)
	}

	for i, _ := range machine.Interfaces {
		err := system.DeleteInterface(MachineIface(id, i))
		if err != nil {
			log.Printf("Could not delete machine '%s' interfaces: %s\n", id, err)
		}
	}

	err = DBMachineDelete(id)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, nil)
}

// GET /machines/<id>/start
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

// GET /machines/<id>/stop
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

// GET /machines/<id>/status
func HandleMachineStatus(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	if !DBMachineExists(id) {
		ErrorResponse(w, r, fmt.Errorf("Machine not found"), 404)
		return
	}

	def, err := MachineKvmStatus(id)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, def)
}
