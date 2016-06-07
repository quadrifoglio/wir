package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/machine"
	"github.com/quadrifoglio/wir/utils"
)

func handleMachineCreate(w http.ResponseWriter, r *http.Request) {
	var m client.MachineRequest

	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	if len(m.Name) == 0 {
		m.Name = utils.UniqueID()
	}

	if !DBMachineNameFree(m.Name) {
		ErrorResponse(errors.NameUsed).Send(w, r)
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
		mm, err = machine.QemuCreate(Conf.MachinePath, m.Name, &img, m.Cores, m.Memory)
		break
	case image.TypeVz:
		mm, err = machine.VzCreate(Conf.MachinePath, m.Name, &img, m.Cores, m.Memory)
		break
	default:
		err = errors.InvalidImageType
		break
	}

	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	mm.Network = m.Network

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
		case image.TypeVz:
			machine.VzCheck(&ms[i])
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
	name := vars["name"]

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	switch m.Type {
	case image.TypeQemu:
		machine.QemuCheck(&m)
		break
	case image.TypeVz:
		machine.VzCheck(&m)
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
	name := vars["name"]

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	switch m.Type {
	case image.TypeQemu:
		machine.QemuCheck(&m)
		break
	case image.TypeVz:
		machine.VzCheck(&m)
		break
	}

	if m.State != machine.StateDown {
		ErrorResponse(errors.InvalidMachineState).Send(w, r)
		return
	}

	switch m.Type {
	case image.TypeQemu:
		err = machine.QemuStart(&m, Conf.MachinePath)
		break
	case image.TypeVz:
		err = machine.VzStart(&m)
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
	name := vars["name"]

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	switch m.Type {
	case image.TypeQemu:
		machine.QemuCheck(&m)
		break
	case image.TypeVz:
		machine.VzCheck(&m)
		break
	default:
		ErrorResponse(errors.InvalidImageType)
		return
	}

	if m.State != machine.StateUp {
		ErrorResponse(errors.InvalidMachineState)
		return
	}

	switch m.Type {
	case image.TypeQemu:
		err = machine.QemuStop(&m)
		if err != nil {
			ErrorResponse(err).Send(w, r)
			return
		}
		break
	case image.TypeVz:
		err = machine.VzStop(&m)
		if err != nil {
			ErrorResponse(err).Send(w, r)
			return
		}
		break
	default:
		ErrorResponse(errors.InvalidImageType)
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
	name := vars["name"]

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	switch m.Type {
	case image.TypeQemu:
		machine.QemuCheck(&m)
		break
	case image.TypeVz:
		machine.VzCheck(&m)
		break
	default:
		ErrorResponse(errors.InvalidImageType)
		return
	}

	if m.State != machine.StateDown {
		ErrorResponse(errors.InvalidMachineState).Send(w, r)
		return
	}

	switch m.Type {
	case image.TypeQemu:
		err = machine.QemuDelete(&m)
		break
	case image.TypeVz:
		err = machine.VzDelete(&m)
		break
	}

	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	DBDeleteMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(nil).Send(w, r)
}
