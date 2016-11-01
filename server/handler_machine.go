package server

import (
	"encoding/json"
	"fmt"
	"io"
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
	} else if req.Disk == 0 {
		return fmt.Errorf("Specify an 'Image' and/or a 'Disk' size"), 400
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
		} else {
			netw, err := DBNetworkGet(iface.Network)
			if err != nil {
				return fmt.Errorf("Failed to get '%s' network: %s", iface.Network, err), 500
			}

			if netw.DHCP.Enabled { // If internal DHCP is used, we should associate an IP to the VM
				ip, err := NetworkFreeLease(netw)
				if err != nil {
					return fmt.Errorf("Can't get free lease in network '%s': %s\n", netw.ID, err), 500
				}

				req.Interfaces[i].IP = ip.String()
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
		req.ID = utils.RandID()
		if !DBMachineExists(req.ID) {
			break
		}
	}

	// Check if the MAC addresses are available
	for _, iface := range req.Interfaces {
		if !DBIsMACFree(iface.MAC) {
			ErrorResponse(w, r, fmt.Errorf("MAC address is already in use"), 400)
			return
		}
	}

	err = MachineKvmCreate(&req)
	if err != nil {
		ErrorResponse(w, r, err, 500)
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

// DELETE /machines/<id>
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
		err := system.DeleteInterface(MachineNicName(id, i))
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

// GET /machines/<id>/kvm
func HandleMachineGetKvmOpts(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	if !DBMachineExists(id) {
		ErrorResponse(w, r, fmt.Errorf("Machine not found"), 404)
		return
	}

	opts, err := DBMachineGetKvmOpts(id)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, opts)
}

// POST /machines/<id>/kvm
func HandleMachineSetKvmOpts(w http.ResponseWriter, r *http.Request) {
	var req shared.KvmOptsDef

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

	req.PID = -1 // The PID should not be modified by the request

	if req.VNC.Enabled {
		if net.ParseIP(req.VNC.Address) == nil {
			ErrorResponse(w, r, fmt.Errorf("Invalid 'VNC.Address' IP address"), 400)
			return
		}
		if req.VNC.Port == 0 {
			ErrorResponse(w, r, fmt.Errorf("Invalid 'VNC.Port' number"), 400)
			return
		}
	}

	err = DBMachineSetKvmOpts(id, req)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	opts, err := DBMachineGetKvmOpts(id)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, opts)
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

// GET /machines/<id>/disk/data
func HandleMachineDiskData(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	if !DBMachineExists(id) {
		ErrorResponse(w, r, fmt.Errorf("Machine not found"), 404)
		return
	}

	f, err := os.Open(MachineDisk(id))
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	defer f.Close()

	// TODO: Log

	_, err = io.Copy(w, f)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}
}
