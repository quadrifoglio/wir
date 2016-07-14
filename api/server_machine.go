package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/gorilla/mux"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/config"
	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/inter"
	"github.com/quadrifoglio/wir/machine"
	"github.com/quadrifoglio/wir/net"
	"github.com/quadrifoglio/wir/utils"
)

func handleMachineCreate(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	var m client.MachineRequest

	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	if len(m.Name) == 0 {
		m.Name = utils.UniqueID(config.API.NodeID)
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
		mm, err = machine.QemuCreate(m.Name, &img, m.Cores, m.Memory)
		break
	case image.TypeVz:
		i, err := DBFreeMachineIndex()
		if err == nil {
			mm, err = machine.VzCreate(m.Name, i, &img, m.Cores, m.Memory)
		}
		break
	case image.TypeLXC:
		mm, err = machine.LxcCreate(m.Name, &img, m.Cores, m.Memory)
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
			mm.Network.MAC, err = net.GenerateMAC(config.API.NodeID)
			if err != nil {
				ErrorResponse(err).Send(w, r)
				return
			}
		}

		err = net.GrantBasic(config.API.Ebtables, mm.Network.MAC)
		if err != nil {
			ErrorResponse(err).Send(w, r)
			return
		}

		if len(m.Network.Mode) > 0 && len(mm.Network.IP) > 0 {
			err := net.GrantTraffic(config.API.Ebtables, mm.Network.MAC, mm.Network.IP)
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
	PrepareResponse(w, r)

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
		net.DenyTraffic(config.API.Ebtables, m.Network.MAC, m.Network.IP) // Not handling errors: can fail if no ip was previously registered

		m.Network.MAC = nm.Network.MAC
		err = net.GrantTraffic(config.API.Ebtables, m.Network.MAC, m.Network.IP)
		if err != nil {
			ErrorResponse(err).Send(w, r)
			return
		}
	}

	if len(nm.Network.IP) > 0 && nm.Network.IP != m.Network.IP {
		net.DenyTraffic(config.API.Ebtables, m.Network.MAC, m.Network.IP)

		m.Network.IP = nm.Network.IP

		err = net.GrantTraffic(config.API.Ebtables, m.Network.MAC, m.Network.IP)
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

func handleMachineLinuxSysprep(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	vars := mux.Vars(r)
	name := vars["name"]

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	img, err := DBGetImage(m.Image)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	if img.Type == image.TypeQemu && img.MainPartition == 0 {
		ErrorResponse(fmt.Errorf("image does not have a specified main partition. can not sysprep.")).Send(w, r)
		return
	}

	var sp client.LinuxSysprep

	err = json.NewDecoder(r.Body).Decode(&sp)
	if err != nil {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	switch m.Type {
	case image.TypeQemu:
		err = machine.QemuLinuxSysprep(&m, img.MainPartition, sp.Hostname, sp.RootPasswd)
		break
	case image.TypeVz:
		//err = machine.VzLinuxSysprep()
		break
	case image.TypeLXC:
		err = machine.LxcLinuxSysprep(&m, sp.Hostname, sp.RootPasswd)
		break
	}

	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(nil).Send(w, r)
}

func handleMachineList(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

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
			machine.VzCheck(&ms[i])
			break
		case image.TypeLXC:
			machine.LxcCheck(&ms[i])
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
	PrepareResponse(w, r)

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
	case image.TypeLXC:
		machine.LxcCheck(&m)
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
	PrepareResponse(w, r)

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
	case image.TypeLXC:
		machine.LxcCheck(&m)
		break
	}

	if m.State != machine.StateDown {
		ErrorResponse(errors.InvalidMachineState).Send(w, r)
		return
	}

	switch m.Type {
	case image.TypeQemu:
		err = machine.QemuStart(&m, config.API.MachinePath)
		break
	case image.TypeVz:
		err = machine.VzStart(&m)
		break
	case image.TypeLXC:
		err = machine.LxcStart(&m)
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
	PrepareResponse(w, r)

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
		//stats, err = machine.VzStats(&m)
		break
	case image.TypeLXC:
		stats, err = machine.LxcStats(&m)
		break
	}

	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(stats).Send(w, r)
}

func handleMachineStop(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

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
	case image.TypeLXC:
		machine.LxcCheck(&m)
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
	case image.TypeLXC:
		err = machine.LxcStop(&m)
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
	PrepareResponse(w, r)

	type Request struct {
		Target client.Remote
		Live   bool
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
		machine.VzCheck(&m)
		break
	case image.TypeLXC:
		machine.LxcCheck(&m)
		break
	}

	if !req.Live && m.State != machine.StateDown {
		ErrorResponse(errors.InvalidMachineState).Send(w, r)
		return
	}

	switch m.Type {
	case image.TypeQemu:
		err = inter.MigrateQemu(m, i, client.Remote{config.API.Address, "root", config.API.Port}, req.Target)
		break
	case image.TypeLXC:
		if req.Live {
			err = inter.LiveMigrateLxc(m, client.Remote{config.API.Address, "root", config.API.Port}, req.Target)
		} else {
			err = inter.MigrateLxc(m, i, client.Remote{config.API.Address, "root", config.API.Port}, req.Target)
		}
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
	PrepareResponse(w, r)

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
	case image.TypeLXC:
		machine.LxcCheck(&m)
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
	case image.TypeLXC:
		err = machine.LxcDelete(&m)
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
