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

	Images struct {
		Endpoint string
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

	r.HandleFunc(fmt.Sprintf("%s", c.Images.Endpoint), server.HandleImageCreate).Methods("POST")
	r.HandleFunc(fmt.Sprintf("%s", c.Images.Endpoint), server.HandleImageList).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s/{id}", c.Images.Endpoint), server.HandleImageGet).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s/{id}", c.Images.Endpoint), server.HandleImageUpdate).Methods("POST")
	r.HandleFunc(fmt.Sprintf("%s/{id}", c.Images.Endpoint), server.HandleImageDelete).Methods("DELETE")

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(c.HTTP.Listen, nil))
}
