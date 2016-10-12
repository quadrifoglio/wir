package main

import (
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/mux"

	"github.com/quadrifoglio/wir/server"
)

type config struct {
	Server struct {
		Node     byte
		Listen   string // Listen address of the HTTP server
		Database string // Path of the database file
	}

	Storage struct {
		Images   string // Folder in which images are stored
		Machines string // Folder in which machines are stored
	}
}

func main() {
	log.Println("Starting wird")

	var c config
	if _, err := toml.DecodeFile("wird.toml", &c); err != nil {
		log.Fatal(err)
	}

	err := server.Init(c.Server.Node, c.Server.Database, c.Storage.Images, c.Storage.Machines)
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
	r.HandleFunc("/images/{id}", server.HandleImageData).Methods("DATA")

	r.HandleFunc("/networks", server.HandleNetworkCreate).Methods("POST")
	r.HandleFunc("/networks", server.HandleNetworkList).Methods("GET")
	r.HandleFunc("/networks/{id}", server.HandleNetworkGet).Methods("GET")
	r.HandleFunc("/networks/{id}", server.HandleNetworkUpdate).Methods("POST")
	r.HandleFunc("/networks/{id}", server.HandleNetworkDelete).Methods("DELETE")

	r.HandleFunc("/volumes", server.HandleVolumeCreate).Methods("POST")
	r.HandleFunc("/volumes", server.HandleVolumeList).Methods("GET")
	r.HandleFunc("/volumes/{id}", server.HandleVolumeGet).Methods("GET")
	r.HandleFunc("/volumes/{id}", server.HandleVolumeUpdate).Methods("POST")
	r.HandleFunc("/volumes/{id}", server.HandleVolumeDelete).Methods("DELETE")

	r.HandleFunc("/machines", server.HandleMachineCreate).Methods("POST")
	r.HandleFunc("/machines", server.HandleMachineList).Methods("GET")
	r.HandleFunc("/machines/{id}", server.HandleMachineGet).Methods("GET")
	r.HandleFunc("/machines/{id}", server.HandleMachineUpdate).Methods("POST")
	r.HandleFunc("/machines/{id}", server.HandleMachineDelete).Methods("DELETE")

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(c.Server.Listen, nil))
}
