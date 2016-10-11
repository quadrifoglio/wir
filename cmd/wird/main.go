package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/mux"

	"github.com/quadrifoglio/wir/server"
)

type config struct {
	HTTP struct {
		Listen string
	}
}

func main() {
	log.Println("starting wird")

	var c config
	if _, err := toml.DecodeFile("wird.toml", &c); err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/", server.HandleIndex).Methods("GET")

	r.HandleFunc("/images", server.HandleImageCreate).Methods("POST")
	r.HandleFunc("/images", server.HandleImageList).Methods("GET")
	r.HandleFunc("/images/{id}", server.HandleImageGet).Methods("GET")
	r.HandleFunc("/images/{id}", server.HandleImageUpdate).Methods("POST")
	r.HandleFunc("/images/{id}", server.HandleImageDelete).Methods("DELETE")

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(c.HTTP.Listen, nil))
}
