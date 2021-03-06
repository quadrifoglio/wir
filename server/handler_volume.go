package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/utils"
)

// validateVolume validates the requested volume definition
// and returns the coresponding http satus code
func validateVolume(req shared.VolumeDef) (error, int) {
	if len(req.Name) == 0 {
		return fmt.Errorf("Missing 'Name'"), 400
	}
	if req.Size == 0 {
		return fmt.Errorf("'Size' can't be 0"), 400
	}

	if len(req.Type) > 0 {
		if req.Type != shared.BackendKVM && req.Type != shared.BackendLXC {
			return fmt.Errorf("Invalid 'Type' (must be kvm, lxc)"), 400
		}
	} else {
		return fmt.Errorf("Missing 'Type'"), 400
	}

	return nil, 200
}

// POST /volumes
func HandleVolumeCreate(w http.ResponseWriter, r *http.Request) {
	var req shared.VolumeDef

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ErrorResponse(w, r, err, 400)
		return
	}

	err, status := validateVolume(req)
	if err != nil {
		ErrorResponse(w, r, err, status)
		return
	}

	for {
		req.ID = utils.RandID()
		if !DBVolumeExists(req.ID) {
			break
		}
	}

	err = CreateVolume(req)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	err = DBVolumeCreate(req)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, req)
}

// GET /volumes
func HandleVolumeList(w http.ResponseWriter, r *http.Request) {
	volumes, err := DBVolumeList()
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, volumes)
}

// GET /volumes/<id>
func HandleVolumeGet(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	if !DBVolumeExists(id) {
		ErrorResponse(w, r, fmt.Errorf("Volume not found"), 404)
		return
	}

	volume, err := DBVolumeGet(id)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, volume)
}

// POST /volumes/<id>
func HandleVolumeUpdate(w http.ResponseWriter, r *http.Request) {
	var req shared.VolumeDef

	v := mux.Vars(r)
	id := v["id"]

	if !DBVolumeExists(id) {
		ErrorResponse(w, r, fmt.Errorf("Volume not found"), 404)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ErrorResponse(w, r, err, 400)
		return
	}

	req.ID = id

	err, status := validateVolume(req)
	if err != nil {
		ErrorResponse(w, r, err, status)
		return
	}

	err = DBVolumeUpdate(req)
	if err != nil {
		ErrorResponse(w, r, err, 404)
		return
	}

	SuccessResponse(w, r, req)
}

// DELETE /volumes/<id>
func HandleVolumeDelete(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	if !DBVolumeExists(id) {
		ErrorResponse(w, r, fmt.Errorf("Volume not found"), 404)
		return
	}

	err := DeleteVolume(id)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	err = DBVolumeDelete(id)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, nil)
}
