package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/quadrifoglio/wir/shared"
)

// valnameateNetwork checks if the requested definiton
// is valname, and returns the coresponding http status code
func valnameateNetwork(req shared.NetworkDef) (error, int) {
	if len(req.Name) == 0 {
		return fmt.Errorf("Missing 'Name'"), 400
	}
	if len(req.Name) > 12 {
		return fmt.Errorf("'Name' must be less than or equal to 12 characters"), 400
	}
	if len(req.CIDR) == 0 {
		return fmt.Errorf("Missing 'CIDR'"), 400
	}
	if req.DHCP.Enabled {
		if len(req.DHCP.StartIP) > 0 {
			if ip := net.ParseIP(req.DHCP.StartIP); ip == nil {
				return fmt.Errorf("Invalname 'DHCP.StartIP'"), 400
			}
		} else {
			return fmt.Errorf("Missing 'DHCP.StartIP'"), 400
		}

		if req.DHCP.NumIP == 0 {
			return fmt.Errorf("'DHCP.NumIP' can't be 0"), 400
		}

		if len(req.DHCP.Router) > 0 {
			if ip := net.ParseIP(req.DHCP.Router); ip == nil {
				return fmt.Errorf("Invalname 'DHCP.Router'"), 400
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

	err, status := valnameateNetwork(req)
	if err != nil {
		ErrorResponse(w, r, err, status)
		return
	}

	if DBNetworkExists(req.Name) {
		ErrorResponse(w, r, fmt.Errorf("Network already exists"), 400)
		return
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

// GET /networks/<name>
func HandleNetworkGet(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	name := v["name"]

	if !DBNetworkExists(name) {
		ErrorResponse(w, r, fmt.Errorf("Network not found"), 404)
		return
	}

	network, err := DBNetworkGet(name)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, network)
}

// POST /networks/<name>
func HandleNetworkUpdate(w http.ResponseWriter, r *http.Request) {
	var req shared.NetworkDef

	v := mux.Vars(r)
	name := v["name"]

	if !DBNetworkExists(name) {
		ErrorResponse(w, r, fmt.Errorf("Network not found"), 404)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ErrorResponse(w, r, err, 400)
		return
	}

	err, status := valnameateNetwork(req)
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

// DELETE /networks/<name>
func HandleNetworkDelete(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	name := v["name"]

	if !DBNetworkExists(name) {
		ErrorResponse(w, r, fmt.Errorf("Network not found"), 404)
		return
	}

	err := DeleteNetwork(name)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	err = DBNetworkDelete(name)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, nil)
}
