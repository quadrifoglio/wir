package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

const (
	Version    = "0.0.1"
	ConfigFile = "/etc/wir.json"
)

var (
	Config WirConfig
)

type WirConfig struct {
	ImagesDir    string `json:"images_location"`
	DrivesDir    string `json:"drives_location"`
	DatabaseFile string `json:"database_location"`
}

func main() {
	log.Println("Starting wird version", Version)

	f, err := os.Open(ConfigFile)
	if err != nil {
		log.Fatal("Can not open config: ", err)
	}

	err = json.NewDecoder(f).Decode(&Config)
	if err != nil {
		f.Close()
		log.Fatal("Can not decode config file: ", err)
	}

	f.Close()

	if len(Config.ImagesDir) == 0 || len(Config.ImagesDir) == 0 || len(Config.DatabaseFile) == 0 {
		log.Fatal("Invalid configuration file")
	}

	err = DatabaseOpen(Config.DatabaseFile)
	if err != nil {
		log.Fatal("Can not initialize database: ", err)
	}

	defer Database.Close()

	r := mux.NewRouter()

	r.HandleFunc("/image/create", HandleImageCreate).Methods("POST")
	r.HandleFunc("/image/list", HandleImageList).Methods("GET")

	r.HandleFunc("/vm/create", HandleVmCreate).Methods("POST")
	r.HandleFunc("/vm/list", HandleVmList).Methods("GET")

	r.HandleFunc("/vm/{id}", HandleVmGet).Methods("GET")
	r.HandleFunc("/vm/{id}/start", HandleVmStart).Methods("GET")
	r.HandleFunc("/vm/{id}/stop", HandleVmStop).Methods("GET")

	http.Handle("/", r)

	log.Println("Listening on 0.0.0.0:8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
