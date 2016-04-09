package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func HandleVmList(w http.ResponseWriter, r *http.Request) {
	vms, err := VmGetAll()
	if err != nil {
		SendError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	err = json.NewEncoder(w).Encode(vms)
	if err != nil {
		SendError(w, r, http.StatusInternalServerError, "Can not encode vm list to json: "+err.Error())
		return
	}
}

func HandleVmGet(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)

	id, err := strconv.Atoi(v["id"])
	if err != nil {
		SendError(w, r, http.StatusBadRequest, "Bad request: invalid id, should be an integer")
		return
	}

	vm, err := VmGet(id)
	if err != nil {
		SendError(w, r, http.StatusNotFound, err.Error())
		return
	}

	err = json.NewEncoder(w).Encode(vm)
	if err != nil {
		SendError(w, r, http.StatusInternalServerError, "Can not encode vm to json: "+err.Error())
		return
	}
}

func HandleVmCreate(w http.ResponseWriter, r *http.Request) {
	var params VmParams

	params.Backend = BackendQemu
	params.Cores = 1
	params.Memory = 512

	vm, err := VmCreate(&params)
	if err != nil {
		SendError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	err = json.NewEncoder(w).Encode(vm)
	if err != nil {
		SendError(w, r, http.StatusInternalServerError, "Can not encode vm to json: "+err.Error())
		return
	}
}

func HandleVmStart(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)

	id, err := strconv.Atoi(v["id"])
	if err != nil {
		SendError(w, r, http.StatusBadRequest, "Bad request: invalid id, should be an integer")
		return
	}

	vm, err := VmGet(id)
	if err != nil {
		SendError(w, r, http.StatusNotFound, err.Error())
		return
	}

	err = vm.Start()
	if err != nil {
		SendError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	err = json.NewEncoder(w).Encode(vm)
	if err != nil {
		SendError(w, r, http.StatusInternalServerError, "Can not encode vm to json: "+err.Error())
		return
	}
}

func HandleVmStop(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)

	id, err := strconv.Atoi(v["id"])
	if err != nil {
		SendError(w, r, http.StatusBadRequest, "Bad request: invalid id, should be an integer")
		return
	}

	vm, err := VmGet(id)
	if err != nil {
		SendError(w, r, http.StatusNotFound, err.Error())
		return
	}

	err = vm.Stop()
	if err != nil {
		SendError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	err = json.NewEncoder(w).Encode(vm)
	if err != nil {
		SendError(w, r, http.StatusInternalServerError, "Can not encode vm to json: "+err.Error())
		return
	}
}

func SendError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	log.Println("Error", r.RemoteAddr, r.Method, r.RequestURI, ":", msg)

	msg = strings.Replace(msg, "\"", "\\\"", -1)

	w.WriteHeader(status)
	fmt.Fprintf(w, "{\"success\": false, \"message\": \"%s\"}", msg)
}
