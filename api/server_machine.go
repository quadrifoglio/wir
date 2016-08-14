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

	switch img.Type {
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

	if req.Live {
		err = LiveMigrateMachine(m, i, shared.Remote{shared.APIConfig.Address, "root", shared.APIConfig.Port}, req.Target)
	} else {
		err = MigrateMachine(m, i, shared.Remote{shared.APIConfig.Address, "root", shared.APIConfig.Port}, req.Target)
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

	if img.Type == shared.TypeQemu && img.MainPartition == 0 {
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

func handleMachineStart(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	vars := mux.Vars(r)
	name := vars["name"]

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
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

func handleMachineListVolumes(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	vars := mux.Vars(r)
	name := vars["name"]

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	vols, err := m.ListVolumes()
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(vols).Send(w, r)
}

func handleMachineCreateVolume(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	vars := mux.Vars(r)
	name := vars["name"]

	var vol shared.Volume

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&vol)
	if err != nil {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	err = m.CreateVolume(vol)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(vol).Send(w, r)
}

func handleMachineDeleteVolume(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	vars := mux.Vars(r)
	name := vars["name"]
	vol := vars["volume"]

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = m.DeleteVolume(vol)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(nil).Send(w, r)
}

func handleMachineListCheckpoints(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	vars := mux.Vars(r)
	name := vars["name"]

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	chks, err := m.ListCheckpoints()
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(chks).Send(w, r)
}

func handleMachineCreateCheckpoint(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	vars := mux.Vars(r)
	name := vars["name"]
	chk := vars["checkpoint"]

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = m.CreateCheckpoint(chk)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(chk).Send(w, r)
}

func handleMachineRestoreCheckpoint(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	vars := mux.Vars(r)
	name := vars["name"]
	chk := vars["checkpoint"]

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = m.RestoreCheckpoint(chk)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(nil).Send(w, r)
}

func handleMachineDeleteCheckpoint(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	vars := mux.Vars(r)
	name := vars["name"]
	chk := vars["checkpoint"]

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = m.DeleteCheckpoint(chk)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(nil).Send(w, r)
}

func handleMachineListBackups(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	vars := mux.Vars(r)
	name := vars["name"]

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	bks, err := m.ListBackups()
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(bks).Send(w, r)
}

func handleMachineCreateBackup(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	vars := mux.Vars(r)
	name := vars["name"]

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	bk, err := m.CreateBackup()
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(bk).Send(w, r)
}

func handleMachineRestoreBackup(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	vars := mux.Vars(r)
	name := vars["name"]
	bk := vars["backup"]

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = m.RestoreBackup(bk)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(nil).Send(w, r)
}

func handleMachineDeleteBackup(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	vars := mux.Vars(r)
	name := vars["name"]
	bk := vars["backup"]

	m, err := DBGetMachine(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = m.DeleteBackup(bk)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(nil).Send(w, r)
}
