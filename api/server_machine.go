package api

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/inter"
	"github.com/quadrifoglio/wir/machine"
	"github.com/quadrifoglio/wir/net"
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
		m.Name = utils.UniqueID(Conf.NodeID)
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
		mm, err = machine.QemuCreate(Conf.QemuImg, Conf.MachinePath, m.Name, &img, m.Cores, m.Memory)
		break
	case image.TypeVz:
		i, err := DBFreeMachineIndex()
		if err == nil {
			mm, err = machine.VzCreate(Conf.Vzctl, Conf.MachinePath, m.Name, i, &img, m.Cores, m.Memory)
		}
		break
	case image.TypeLXC:
		mm, err = machine.LxcCreate(Conf.MachinePath, m.Name, &img, m.Cores, m.Memory)
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

	if len(mm.Network.Mode) > 0 {
		if len(mm.Network.MAC) == 0 {
			mm.Network.MAC, err = net.GenerateMAC(Conf.NodeID)
			if err != nil {
				ErrorResponse(err).Send(w, r)
				return
			}
		}

		err = net.GrantBasic(Conf.Ebtables, mm.Network.MAC)
		if err != nil {
			ErrorResponse(err).Send(w, r)
			return
		}

		if len(m.Network.Mode) > 0 && len(mm.Network.IP) > 0 {
			err := net.GrantTraffic(Conf.Ebtables, mm.Network.MAC, mm.Network.IP)
			if err != nil {
				ErrorResponse(err).Send(w, r)
				return
			}
		}
	}

	err = DBStoreMachine(&mm)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(mm).Send(w, r)
}

func handleMachineUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	var nm client.MachineRequest

	err = json.NewDecoder(r.Body).Decode(&nm)
	if err != nil {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	if nm.Cores != 0 {
		m.Cores = nm.Cores
	}

	if nm.Memory != 0 {
		m.Memory = nm.Memory
	}

	if len(nm.Network.Mode) > 0 && nm.Network.Mode != m.Network.Mode {
		m.Network.Mode = nm.Network.Mode
	}

	if len(nm.Network.MAC) > 0 && nm.Network.MAC != m.Network.MAC {
		net.DenyTraffic(Conf.Ebtables, m.Network.MAC, m.Network.IP) // Not handling errors: can fail if no ip was previously registered

		m.Network.MAC = nm.Network.MAC
		err = net.GrantTraffic(Conf.Ebtables, m.Network.MAC, m.Network.IP)
		if err != nil {
			ErrorResponse(err).Send(w, r)
			return
		}
	}

	if len(nm.Network.IP) > 0 && nm.Network.IP != m.Network.IP {
		net.DenyTraffic(Conf.Ebtables, m.Network.MAC, m.Network.IP)

		m.Network.IP = nm.Network.IP

		err = net.GrantTraffic(Conf.Ebtables, m.Network.MAC, m.Network.IP)
		if err != nil {
			ErrorResponse(err).Send(w, r)
			return
		}
	}

	err = DBStoreMachine(&m)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(nil).Send(w, r)
}

func handleMachineList(w http.ResponseWriter, r *http.Request) {
	ms, err := DBListMachines()
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	sort.Sort(ms)

	for i, _ := range ms {
		prevState := ms[i].State

		switch ms[i].Type {
		case image.TypeQemu:
			machine.QemuCheck(&ms[i])
			break
		case image.TypeVz:
			machine.VzCheck(Conf.Vzctl, &ms[i])
			break
		case image.TypeLXC:
			machine.LxcCheck(Conf.MachinePath, &ms[i])
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
		machine.VzCheck(Conf.Vzctl, &m)
		break
	case image.TypeLXC:
		machine.LxcCheck(Conf.MachinePath, &m)
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
		machine.VzCheck(Conf.Vzctl, &m)
		break
	case image.TypeLXC:
		machine.LxcCheck(Conf.MachinePath, &m)
		break
	}

	if m.State != machine.StateDown {
		ErrorResponse(errors.InvalidMachineState).Send(w, r)
		return
	}

	switch m.Type {
	case image.TypeQemu:
		err = machine.QemuStart(Conf.Qemu, Conf.EnableKVM, &m, Conf.MachinePath)
		break
	case image.TypeVz:
		err = machine.VzStart(Conf.Vzctl, &m)
		break
	case image.TypeLXC:
		err = machine.LxcStart(Conf.MachinePath, &m)
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

func handleMachineStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	var stats machine.Stats

	switch m.Type {
	case image.TypeQemu:
		stats, err = machine.QemuStats(&m)
		break
	case image.TypeVz:
		//stats, err = machine.VzStats(Conf.MachinePath, &m)
		break
	case image.TypeLXC:
		stats, err = machine.LxcStats(Conf.MachinePath, &m)
		break
	}

	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(stats).Send(w, r)
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
		machine.VzCheck(Conf.Vzctl, &m)
		break
	case image.TypeLXC:
		machine.LxcCheck(Conf.MachinePath, &m)
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
		err = machine.VzStop(Conf.Vzctl, &m)
		if err != nil {
			ErrorResponse(err).Send(w, r)
			return
		}
		break
	case image.TypeLXC:
		err = machine.LxcStop(Conf.MachinePath, &m)
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

func handleMachineMigrate(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		Target client.Remote
	}

	var req Request

	vars := mux.Vars(r)
	name := vars["name"]

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	i, err := DBGetImage(m.Image)
	if err != nil {
		ErrorResponse(errors.ImageNotFound).Send(w, r)
		return
	}

	switch m.Type {
	case image.TypeQemu:
		machine.QemuCheck(&m)
		break
	case image.TypeVz:
		machine.VzCheck(Conf.Vzctl, &m)
		break
	case image.TypeLXC:
		machine.LxcCheck(Conf.MachinePath, &m)
		break
	}

	if m.State != machine.StateDown {
		ErrorResponse(errors.InvalidMachineState).Send(w, r)
		return
	}

	switch m.Type {
	case image.TypeQemu:
		err = inter.MigrateQemu(Conf.MachinePath, m, i, client.Remote{Conf.Address, "root", Conf.Port}, req.Target)
		break
	case image.TypeLXC:
		err = inter.MigrateLxc(Conf.MachinePath, m, i, client.Remote{Conf.Address, "root", Conf.Port}, req.Target)
		break
	default:
		ErrorResponse(errors.InvalidImageType)
		return
	}

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
		machine.VzCheck(Conf.Vzctl, &m)
		break
	case image.TypeLXC:
		machine.LxcCheck(Conf.MachinePath, &m)
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
		err = machine.VzDelete(Conf.Vzctl, &m)
		break
	case image.TypeLXC:
		err = machine.LxcDelete(Conf.MachinePath, &m)
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
