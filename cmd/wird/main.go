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
		Listen   string // Listen address of the HTTP server
		Database string // Path of the database file
	}
}

func main() {
	log.Println("starting wird")

	var c config
	if _, err := toml.DecodeFile("wird.toml", &c); err != nil {
		log.Fatal(err)
	}

	err := server.InitDatabase(c.Server.Database)
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

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(c.Server.Listen, nil))
}
