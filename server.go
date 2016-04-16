package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func HandleImageCreate(w http.ResponseWriter, r *http.Request) {
	var img Image

	err := json.NewDecoder(r.Body).Decode(&img)
	if err != nil {
		SendError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := os.Stat(img.Path); os.IsNotExist(err) {
		SendError(w, r, http.StatusBadRequest, "Invalid image path. File not found")
		return
	}

	img.ID = 0
	err = DatabaseInsertImage(&img)
	if err != nil {
		SendError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	err = json.NewEncoder(w).Encode(img)
	if err != nil {
		SendError(w, r, http.StatusInternalServerError, "Can not encode img to json: "+err.Error())
		return
	}
}

func HandleImageList(w http.ResponseWriter, r *http.Request) {
	imgs, err := DatabaseListImages()
	if err != nil {
		SendError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	err = json.NewEncoder(w).Encode(imgs)
	if err != nil {
		SendError(w, r, http.StatusInternalServerError, "Can not encode image list to json: "+err.Error())
		return
	}
}

func HandleVmCreate(w http.ResponseWriter, r *http.Request) {
	var params VmParams

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		SendError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if len(params.Backend) == 0 || params.Cores == 0 || params.Memory == 0 || params.ImageID == 0 {
		SendError(w, r, http.StatusBadRequest, "Required fields: backend, cores, memory, image_id")
		return
	}

	if params.Backend != "qemu" {
		SendError(w, r, http.StatusBadRequest, "Invalid backend. Supported: qemu")
		return
	}

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