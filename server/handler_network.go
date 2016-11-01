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

// validateNetwork checks if the requested definiton
// is valid, and returns the coresponding http status code
func validateNetwork(req shared.NetworkDef) (error, int) {
	if len(req.Name) == 0 {
		return fmt.Errorf("Missing 'Name'"), 400
	}
	if !DBNetworkNameFree(req.Name) {
		return fmt.Errorf("'Name' already used"), 400
	}
	if len(req.CIDR) == 0 {
		return fmt.Errorf("Missing 'CIDR'"), 400
	}
	if req.DHCP.Enabled {
		if len(req.DHCP.StartIP) > 0 {
			if ip := net.ParseIP(req.DHCP.StartIP); ip == nil {
				return fmt.Errorf("Invalid 'DHCP.StartIP'"), 400
			}
		} else {
			return fmt.Errorf("Missing 'DHCP.StartIP'"), 400
		}

		if req.DHCP.NumIP == 0 {
			return fmt.Errorf("'DHCP.NumIP' can't be 0"), 400
		}

		if len(req.DHCP.Router) > 0 {
			if ip := net.ParseIP(req.DHCP.Router); ip == nil {
				return fmt.Errorf("Invalid 'DHCP.Router'"), 400
			}
		}
	}

	return nil, 200
}

// POST /networks
func HandleNetworkCreate(w http.ResponseWriter, r *http.Request) {
	var req shared.NetworkDef

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ErrorResponse(w, r, err, 400)
		return
	}

	err, status := validateNetwork(req)
	if err != nil {
		ErrorResponse(w, r, err, status)
		return
	}

	for {
		req.ID = utils.RandID()
		if !DBNetworkExists(req.ID) {
			break
		}
	}

	err = CreateNetwork(req)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	err = DBNetworkCreate(req)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, req)
}

// GET /networks
func HandleNetworkList(w http.ResponseWriter, r *http.Request) {
	networks, err := DBNetworkList()
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, networks)
}

// GET /networks/<id>
func HandleNetworkGet(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	if !DBNetworkExists(id) {
		ErrorResponse(w, r, fmt.Errorf("Network not found"), 404)
		return
	}

	network, err := DBNetworkGet(id)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, network)
}

// POST /networks/<id>
func HandleNetworkUpdate(w http.ResponseWriter, r *http.Request) {
	var req shared.NetworkDef

	v := mux.Vars(r)
	id := v["id"]

	if !DBNetworkExists(id) {
		ErrorResponse(w, r, fmt.Errorf("Network not found"), 404)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ErrorResponse(w, r, err, 400)
		return
	}

	req.ID = id

	err, status := validateNetwork(req)
	if err != nil {
		ErrorResponse(w, r, err, status)
		return
	}

	err = DBNetworkUpdate(req)
	if err != nil {
		ErrorResponse(w, r, err, 404)
		return
	}

	SuccessResponse(w, r, req)
}

// DELETE /networks/<id>
func HandleNetworkDelete(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	if !DBNetworkExists(id) {
		ErrorResponse(w, r, fmt.Errorf("Network not found"), 404)
		return
	}

	err := DeleteNetwork(id)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	err = DBNetworkDelete(id)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, nil)
}
