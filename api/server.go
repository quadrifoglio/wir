package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/quadrifoglio/wir/errors"
)

const (
	Version = "0.0.1"
)

type Config struct {
	Address      string
	DatabaseFile string

	ImagePath   string
	MachinePath string
}

var (
	Conf Config
)

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	ErrorResponse(errors.NotFound).Send(w, r)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	type info struct {
		Name    string
		Version string
	}

	i := info{"wir api", Version}
	SuccessResponse(i).Send(w, r)
}

func Start(conf Config) error {
	Conf = conf

	err := DBOpen(Conf.DatabaseFile)
	if err != nil {
		return err
	}

	r := mux.NewRouter()
	r.HandleFunc("/", handleIndex).Methods("GET")

	r.HandleFunc("/image", handleImageCreate).Methods("POST")
	r.HandleFunc("/image/{name}", handleImageGet).Methods("GET")
	r.HandleFunc("/image/{name}", handleImageDelete).Methods("DELETE")

	r.HandleFunc("/machine", handleMachineCreate).Methods("POST")
	r.HandleFunc("/machine/{id}", handleMachineGet).Methods("GET")
	r.HandleFunc("/machine/{id}", handleMachineDelete).Methods("DELETE")

	r.NotFoundHandler = http.HandlerFunc(handleNotFound)
	http.Handle("/", r)

	return http.ListenAndServe(Conf.Address, nil)
}
