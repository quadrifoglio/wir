package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/quadrifoglio/wir/shared"
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
	if len(req.URL) == 0 {
		ErrorResponse(w, r, fmt.Errorf("Missing 'URL'"), 400)
		return
	}

	err = DBImageCreate(req)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}
}

func HandleImageList(w http.ResponseWriter, r *http.Request) {
	images, err := DBImageList()
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	err = json.NewEncoder(w).Encode(images)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}
}

func HandleImageGet(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	image, err := DBImageGet(id)
	if err != nil {
		ErrorResponse(w, r, err, 404)
		return
	}

	err = json.NewEncoder(w).Encode(image)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}
}

func HandleImageUpdate(w http.ResponseWriter, r *http.Request) {
	var req shared.ImageDef

	v := mux.Vars(r)
	id := v["id"]

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
	if len(req.URL) == 0 {
		ErrorResponse(w, r, fmt.Errorf("Missing 'URL'"), 400)
		return
	}

	err = DBImageUpdate(req)
	if err != nil {
		ErrorResponse(w, r, err, 404)
		return
	}

	err = json.NewEncoder(w).Encode(req)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}
}

func HandleImageDelete(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	err := DBImageDelete(id)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}
}
