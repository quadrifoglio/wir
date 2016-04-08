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

	r := mux.NewRouter()

	r.HandleFunc("/vm/list", HandleVmList).Methods("GET")
	r.HandleFunc("/vm/{id}", HandleVmGet).Methods("GET")
	r.HandleFunc("/vm", HandleVmCreate).Methods("POST")

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
