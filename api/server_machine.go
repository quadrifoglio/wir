package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/gorilla/mux"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/utils"
)

func handleMachineCreate(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	var m shared.MachineInfo

	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	if len(m.Name) == 0 {
		m.Name = utils.UniqueID(shared.APIConfig.NodeID)
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

	var mm Machine

	switch img.Info().Type {
	case shared.TypeQemu:
		mm = new(QemuMachine)
		err = mm.Create(img, m)
		break
	case shared.TypeLXC:
		mm = new(LxcMachine)
		err = mm.Create(img, m)
		break
	}

	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = SetupMachineNetwork(mm, m.Network)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = DBStoreMachine(mm)
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

	var nm shared.MachineInfo

	err = json.NewDecoder(r.Body).Decode(&nm)
	if err != nil {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	err = m.Update(nm)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = UpdateMachineNetwork(m, nm.Network)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = DBStoreMachine(m)
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

	img, err := DBGetImage(m.Info().Image)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	if img.Info().Type == shared.TypeQemu && img.Info().MainPartition == 0 {
		ErrorResponse(fmt.Errorf("image does not have a specified main partition. can not sysprep.")).Send(w, r)
		return
	}

	var sp client.LinuxSysprep

	err = json.NewDecoder(r.Body).Decode(&sp)
	if err != nil {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	err = m.Sysprep("linux", sp.Hostname, sp.RootPasswd)
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
		prevState := ms[i].State()

		if ms[i].State() != prevState {
			err = DBStoreMachine(ms[i])
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

	err = DBStoreMachine(m)
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

	if m.State() != shared.StateDown {
		ErrorResponse(errors.InvalidMachineState).Send(w, r)
		return
	}

	err = m.Start()
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = DBStoreMachine(m)
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

	stats, err := m.Stats()
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

	err = m.Stop()
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = DBStoreMachine(m)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(nil).Send(w, r)
}

func handleMachineMigrate(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	type Request struct {
		Target shared.Remote
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

	i, err := DBGetImage(m.Info().Image)
	if err != nil {
		ErrorResponse(errors.ImageNotFound).Send(w, r)
		return
	}

	if (req.Live && m.State() != shared.StateUp) || (!req.Live && m.State() != shared.StateDown) {
		ErrorResponse(errors.InvalidMachineState).Send(w, r)
		return
	}

	switch i.Info().Type {
	case shared.TypeQemu:
		if req.Live {
			err = LiveMigrateQemu(m, i, shared.Remote{shared.APIConfig.Address, "root", shared.APIConfig.Port}, req.Target)
		} else {
			err = MigrateQemu(m, i, shared.Remote{shared.APIConfig.Address, "root", shared.APIConfig.Port}, req.Target)
		}
		break
	case shared.TypeLXC:
		if req.Live {
			err = LiveMigrateLxc(m, i, shared.Remote{shared.APIConfig.Address, "root", shared.APIConfig.Port}, req.Target)
		} else {
			err = MigrateLxc(m, i, shared.Remote{shared.APIConfig.Address, "root", shared.APIConfig.Port}, req.Target)
		}
		break
	default:
		err = errors.InvalidImageType
		return
	}

	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = m.Delete()
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = DBDeleteMachine(name)
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

	err = m.Delete()
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = DBDeleteMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(nil).Send(w, r)
}
