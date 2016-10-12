package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/system"
)

func HandleImageCreate(w http.ResponseWriter, r *http.Request) {
	var req shared.ImageDef

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ErrorResponse(w, r, err, 400)
		return
	}

	if len(req.Name) == 0 {
		ErrorResponse(w, r, fmt.Errorf("Missing 'Name'"), 400)
		return
	}
	if len(req.Source) == 0 {
		ErrorResponse(w, r, fmt.Errorf("Missing 'Source'"), 400)
		return
	}

	dst := ImageFile(req.Name)

	err = system.FetchURL(req.Source, dst)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	req.Source = dst

	err = DBImageCreate(&req)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, req)
}

func HandleImageList(w http.ResponseWriter, r *http.Request) {
	images, err := DBImageList()
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, images)
}

func HandleImageGet(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	if !DBImageExists(id) {
		ErrorResponse(w, r, fmt.Errorf("Image not found"), 404)
		return
	}

	image, err := DBImageGet(id)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, image)
}

func HandleImageUpdate(w http.ResponseWriter, r *http.Request) {
	var req shared.ImageDef

	v := mux.Vars(r)
	id := v["id"]

	if !DBImageExists(id) {
		ErrorResponse(w, r, fmt.Errorf("Image not found"), 404)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ErrorResponse(w, r, err, 400)
		return
	}

	req.ID = id

	if len(req.Name) == 0 {
		ErrorResponse(w, r, fmt.Errorf("Missing 'Name'"), 400)
		return
	}
	if len(req.Source) == 0 {
		ErrorResponse(w, r, fmt.Errorf("Missing 'Source'"), 400)
		return
	}

	err = DBImageUpdate(&req)
	if err != nil {
		ErrorResponse(w, r, err, 404)
		return
	}

	SuccessResponse(w, r, req)
}

func HandleImageDelete(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	if !DBImageExists(id) {
		ErrorResponse(w, r, fmt.Errorf("Image not found"), 404)
		return
	}

	err := DBImageDelete(id)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, nil)
}
