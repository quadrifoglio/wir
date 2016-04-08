package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	Version = "0.0.1"
)

func main() {
	log.Println("Starting wird version", Version)

	err := DatabaseOpen()
	if err != nil {
		log.Fatal("Can not initialize database: ", err)
	}

	defer Database.Close()

	r := mux.NewRouter()

	r.HandleFunc("/vm/create", HandleVmCreate).Methods("POST")
	r.HandleFunc("/vm/list", HandleVmList).Methods("GET")
	r.HandleFunc("/vm/{id}", HandleVmGet).Methods("GET")

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
