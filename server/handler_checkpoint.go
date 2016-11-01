package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/quadrifoglio/wir/shared"
)

// validateCheckpoint validates the requested image definition
// and returns the coresponding http status code
func validateCheckpoint(req shared.CheckpointDef) (error, int) {
	if len(req.Name) == 0 {
		return fmt.Errorf("Missing 'Name'"), 400
	}

	return nil, 200
}

// POST /machines/<id>/checkpoints
func HandleCheckpointCreate(w http.ResponseWriter, r *http.Request) {
	var req shared.CheckpointDef

	v := mux.Vars(r)
	machine := v["id"]

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ErrorResponse(w, r, err, 400)
		return
	}

	err, status := validateCheckpoint(req)
	if err != nil {
		ErrorResponse(w, r, err, status)
		return
	}

	req.Timestamp = time.Now().Unix()

	err = MachineKvmCreateCheckpoint(machine, req.Name)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, req)
}

// GET /machines/<id>/checkpoints
func HandleCheckpointList(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	machine := v["id"]

	chks, err := MachineKvmListCheckpoints(machine)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, chks)
}

// GET /machines/<id>/checkpoints/<name>/restore
func HandleCheckpointRestore(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	machine := v["id"]
	name := v["name"]

	err := MachineKvmRestoreCheckpoint(machine, name)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, nil)
}

// DELETE /machines/<id>/checkpoints/<name>
func HandleCheckpointDelete(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	machine := v["id"]
	name := v["name"]

	err := MachineKvmDeleteCheckpoint(machine, name)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, nil)
}
