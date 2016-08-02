package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/net"
	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/utils"
)

const (
	Version = "0.0.1"
)

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	ErrorResponse(errors.NotFound).Send(w, r)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	type stats struct {
		CPUUsage  float32
		RAMUsage  uint64
		RAMTotal  uint64
		FreeSpace uint64
	}

	type info struct {
		Name          string
		Version       string
		Configuration shared.APIConfigStruct
		Stats         stats
	}

	cpu, err := utils.GetCpuUsagePercentage()
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	ru, rt, err := utils.GetRamUsage()
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	fs, err := utils.GetFreeSpace(shared.APIConfig.MachinePath)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	i := info{"wir api", Version, shared.APIConfig, stats{cpu, ru, rt, fs}}
	SuccessResponse(i).Send(w, r)
}

func Start() error {
	err := os.MkdirAll(filepath.Dir(shared.APIConfig.DatabaseFile), 0775)
	if err != nil {
		return err
	}

	err = os.MkdirAll(shared.APIConfig.ImagePath, 0775)
	if err != nil {
		return err
	}

	err = os.MkdirAll(shared.APIConfig.MigrationPath, 0775)
	if err != nil {
		return err
	}

	err = os.MkdirAll(shared.APIConfig.MachinePath, 0775)
	if err != nil {
		return err
	}

	err = DBOpen(shared.APIConfig.DatabaseFile)
	if err != nil {
		return err
	}

	err = net.Init(shared.APIConfig.Ebtables, shared.APIConfig.BridgeIface)
	if err != nil {
		return err
	}

	r := mux.NewRouter()
	r.HandleFunc("/", handleIndex).Methods("GET")

	r.HandleFunc("/images", handleImageCreate).Methods("POST")
	r.HandleFunc("/images", handleImageList).Methods("GET")
	r.HandleFunc("/images/{name}", handleImageGet).Methods("GET")
	r.HandleFunc("/images/{name}", handleImageDelete).Methods("DELETE")

	r.HandleFunc("/machines", handleMachineCreate).Methods("POST")
	r.HandleFunc("/machines", handleMachineList).Methods("GET")
	r.HandleFunc("/machines/{name}", handleMachineUpdate).Methods("POST")
	r.HandleFunc("/machines/{name}", handleMachineLinuxSysprep).Methods("SYSPREP")
	r.HandleFunc("/machines/{name}", handleMachineGet).Methods("GET")
	r.HandleFunc("/machines/{name}", handleMachineStart).Methods("START")
	r.HandleFunc("/machines/{name}", handleMachineStats).Methods("STATS")
	r.HandleFunc("/machines/{name}", handleMachineStop).Methods("STOP")
	r.HandleFunc("/machines/{name}", handleMachineMigrate).Methods("MIGRATE")
	r.HandleFunc("/machines/{name}", handleMachineDelete).Methods("DELETE")

	r.NotFoundHandler = http.HandlerFunc(handleNotFound)
	http.Handle("/", handlers.CORS()(r))

	return http.ListenAndServe(fmt.Sprintf("%s:%d", shared.APIConfig.Address, shared.APIConfig.Port), nil)
}
