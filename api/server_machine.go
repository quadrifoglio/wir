package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/machine"
)

type MachinePost struct {
	Image  string
	Cores  int
	Memory int
}

func handleMachineCreate(w http.ResponseWriter, r *http.Request) {
	var m MachinePost

	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	img, err := DBGetImage(m.Image)
	if err != nil {
		ErrorResponse(errors.ImageNotFound).Send(w, r)
		return
	}

	var mm machine.Machine

	switch img.Type {
	case image.TypeQemu:
		mm, err = machine.QemuCreate(&img, m.Cores, m.Memory)
		break
	default:
		err = errors.InvalidImageType
		break
	}

	err = DBStoreMachine(&mm)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(mm).Send(w, r)
}

func handleMachineGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	m, err := DBGetMachine(id)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(m).Send(w, r)
}

func handleMachineDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := DBDeleteMachine(id)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(nil).Send(w, r)
}
