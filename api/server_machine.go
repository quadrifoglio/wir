package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/machine"
)

func handleMachineCreate(w http.ResponseWriter, r *http.Request) {
	var m client.MachineRequest

	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	if len(m.Name) == 0 {
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
		mm, err = machine.QemuCreate(Conf.MachinePath, &img, m.Cores, m.Memory)
		break
	default:
		err = errors.InvalidImageType
		break
	}

	mm.NetBridgeOn = m.NetBridgeOn

	err = DBStoreMachine(&mm)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(mm).Send(w, r)
}

func handleMachineList(w http.ResponseWriter, r *http.Request) {
	ms, err := DBListMachines()
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	for i, _ := range ms {
		prevState := ms[i].State

		switch ms[i].Type {
		case image.TypeQemu:
			machine.QemuCheck(&ms[i])
			break
		}

		if ms[i].State != prevState {
			err = DBStoreMachine(&ms[i])
			if err != nil {
				ErrorResponse(err).Send(w, r)
				return
			}
		}
	}

	SuccessResponse(ms).Send(w, r)
}

func handleMachineGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	m, err := DBGetMachine(id)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	switch m.Type {
	case image.TypeQemu:
		machine.QemuCheck(&m)
		break
	}

	err = DBStoreMachine(&m)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(m).Send(w, r)
}

func handleMachineStart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	m, err := DBGetMachine(id)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	if m.State != machine.StateDown {
		ErrorResponse(errors.InvalidMachineState).Send(w, r)
		return
	}

	switch m.Type {
	case image.TypeQemu:
		err = machine.QemuStart(&m, Conf.MachinePath)
		break
	default:
		ErrorResponse(errors.InvalidImageType)
		return
	}

	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = DBStoreMachine(&m)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(nil).Send(w, r)
}

func handleMachineStop(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	m, err := DBGetMachine(id)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	if m.State != machine.StateUp {
		ErrorResponse(errors.InvalidMachineState).Send(w, r)
		return
	}

	switch m.Type {
	case image.TypeQemu:
		err = machine.QemuStop(&m)
		break
	default:
		ErrorResponse(errors.InvalidImageType)
		return
	}

	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = DBStoreMachine(&m)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(nil).Send(w, r)
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
