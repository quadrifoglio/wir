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
	"github.com/quadrifoglio/wir/utils"
)

const (
	Version = "0.0.1"
)

type Config struct {
	Address     string
	Port        int
	BridgeIface string

	EnableKVM bool
	Ebtables  string `json:"EbtablesCommand"`
	QemuImg   string `json:"QemuImgCommand"`
	Qemu      string `json:"QemuCommand"`
	Vzctl     string `json:"VzctlCommand"`

	DatabaseFile string
	ImagePath    string
	MachinePath  string
}

var (
	Conf Config
)

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	ErrorResponse(errors.NotFound).Send(w, r)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	type stats struct {
		CPUUsage  int
		RAMUsage  uint64
		RAMTotal  uint64
		FreeSpace uint64
	}

	type info struct {
		Name          string
		Version       string
		Configuration Config
		Stats         stats
	}

	cpu, err := utils.GetCpuUsage()
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	ru, rt, err := utils.GetRamUsage()
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	fs, err := utils.GetFreeSpace(Conf.MachinePath)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	i := info{"wir api", Version, Conf, stats{cpu, ru, rt, fs}}
	SuccessResponse(i).Send(w, r)
}

func Start(conf Config) error {
	Conf = conf

	err := os.MkdirAll(filepath.Dir(Conf.DatabaseFile), 0777)
	if err != nil {
		return err
	}

	err := os.MkdirAll(Conf.ImagePath, 0777)
	if err != nil {
		return err
	}

	err := os.MkdirAll(Conf.MachinePath, 0777)
	if err != nil {
		return err
	}

	err = DBOpen(Conf.DatabaseFile)
	if err != nil {
		return err
	}

	err = net.Init(Conf.Ebtables, Conf.BridgeIface)
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
	r.HandleFunc("/machines/{name}", handleMachineGet).Methods("GET")
	r.HandleFunc("/machines/{name}", handleMachineStart).Methods("START")
	r.HandleFunc("/machines/{name}", handleMachineStop).Methods("STOP")
	r.HandleFunc("/machines/{name}", handleMachineMigrate).Methods("MIGRATE")
	r.HandleFunc("/machines/{name}", handleMachineDelete).Methods("DELETE")

	r.NotFoundHandler = http.HandlerFunc(handleNotFound)
	http.Handle("/", handlers.CORS()(r))

	return http.ListenAndServe(fmt.Sprintf("%s:%d", Conf.Address, Conf.Port), nil)
}
