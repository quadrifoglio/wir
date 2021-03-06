package main

import (
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/mux"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/quadrifoglio/wir/server"
)

var (
	CConfig = kingpin.Flag("config", "Configuration file to use").Default("/etc/wird.toml").String()
)

type config struct {
	Server struct {
		Node     byte
		Listen   string // Listen address of the HTTP server
		Database string // Path of the database file
	}

	Storage struct {
		Images   string // Folder in which images are stored
		Volumes  string // Folder in which volumes are stored
		Machines string // Folder in which machines are stored
	}
}

func main() {
	kingpin.Parse()

	var c config
	if _, err := toml.DecodeFile(*CConfig, &c); err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting wird - Node #%d\n", c.Server.Node)

	err := server.Init(c.Server.Node, c.Server.Database, c.Storage.Images, c.Storage.Volumes, c.Storage.Machines)
	if err != nil {
		log.Fatal(err)
	}

	defer server.CloseDatabase()

	r := mux.NewRouter()

	r.HandleFunc("/", server.HandleIndex).Methods("GET")

	r.HandleFunc("/images", server.HandleImageCreate).Methods("POST")
	r.HandleFunc("/images", server.HandleImageList).Methods("GET")
	r.HandleFunc("/images/{id}", server.HandleImageGet).Methods("GET")
	r.HandleFunc("/images/{id}", server.HandleImageUpdate).Methods("POST")
	r.HandleFunc("/images/{id}", server.HandleImageDelete).Methods("DELETE")
	r.HandleFunc("/images/{id}/data", server.HandleImageData).Methods("GET")

	r.HandleFunc("/networks", server.HandleNetworkCreate).Methods("POST")
	r.HandleFunc("/networks", server.HandleNetworkList).Methods("GET")
	r.HandleFunc("/networks/{name}", server.HandleNetworkGet).Methods("GET")
	r.HandleFunc("/networks/{name}", server.HandleNetworkUpdate).Methods("POST")
	r.HandleFunc("/networks/{name}", server.HandleNetworkDelete).Methods("DELETE")

	r.HandleFunc("/volumes", server.HandleVolumeCreate).Methods("POST")
	r.HandleFunc("/volumes", server.HandleVolumeList).Methods("GET")
	r.HandleFunc("/volumes/{id}", server.HandleVolumeGet).Methods("GET")
	r.HandleFunc("/volumes/{id}", server.HandleVolumeUpdate).Methods("POST")
	r.HandleFunc("/volumes/{id}", server.HandleVolumeDelete).Methods("DELETE")

	r.HandleFunc("/machines", server.HandleMachineCreate).Methods("POST")
	r.HandleFunc("/machines", server.HandleMachineList).Methods("GET")
	r.HandleFunc("/machines/fetch", server.HandleMachineFetch).Methods("POST")
	r.HandleFunc("/machines/{id}", server.HandleMachineGet).Methods("GET")
	r.HandleFunc("/machines/{id}", server.HandleMachineUpdate).Methods("POST")
	r.HandleFunc("/machines/{id}", server.HandleMachineDelete).Methods("DELETE")
	r.HandleFunc("/machines/{id}/kvm", server.HandleMachineGetKvmOpts).Methods("GET")
	r.HandleFunc("/machines/{id}/kvm", server.HandleMachineSetKvmOpts).Methods("POST")
	r.HandleFunc("/machines/{id}/start", server.HandleMachineStart).Methods("GET")
	r.HandleFunc("/machines/{id}/stop", server.HandleMachineStop).Methods("GET")
	r.HandleFunc("/machines/{id}/status", server.HandleMachineStatus).Methods("GET")
	r.HandleFunc("/machines/{id}/disk/data", server.HandleMachineDiskData).Methods("GET")

	r.HandleFunc("/machines/{id}/checkpoints", server.HandleCheckpointCreate).Methods("POST")
	r.HandleFunc("/machines/{id}/checkpoints", server.HandleCheckpointList).Methods("GET")
	r.HandleFunc("/machines/{id}/checkpoints/{name}", server.HandleCheckpointDelete).Methods("DELETE")
	r.HandleFunc("/machines/{id}/checkpoints/{name}/restore", server.HandleCheckpointRestore).Methods("GET")

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(c.Server.Listen, nil))
}
