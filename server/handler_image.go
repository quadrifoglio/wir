package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/system"
	"github.com/quadrifoglio/wir/utils"
)

func validateImage(req shared.ImageDef) (error, int) {
	if len(req.Name) == 0 {
		return fmt.Errorf("Missing 'Name'"), 400
	}
	if req.Type != shared.BackendKVM && req.Type != shared.BackendLXC {
		return fmt.Errorf("Missing or unsupported 'Type'"), 400
	}
	if len(req.Source) == 0 {
		return fmt.Errorf("Missing 'Source'"), 400
	}

	return nil, 200
}

func HandleImageCreate(w http.ResponseWriter, r *http.Request) {
	var req shared.ImageDef

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		ErrorResponse(w, r, err, 400)
		return
	}

	err, status := validateImage(req)
	if err != nil {
		ErrorResponse(w, r, err, status)
		return
	}

	for {
		req.ID = utils.RandID(GlobalNodeID)
		if !DBImageExists(req.ID) {
			break
		}
	}

	dst := ImageFile(req.ID)

	err = system.FetchURL(req.Source, dst)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	req.Source = dst

	err = DBImageCreate(req)
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

	err, status := validateImage(req)
	if err != nil {
		ErrorResponse(w, r, err, status)
		return
	}

	err = DBImageUpdate(req)
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

	err := os.Remove(ImageFile(id))
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	err = DBImageDelete(id)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	SuccessResponse(w, r, nil)
}

func HandleImageData(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id := v["id"]

	if !DBImageExists(id) {
		ErrorResponse(w, r, fmt.Errorf("Image not found"), 404)
		return
	}

	f, err := os.Open(ImageFile(id))
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}

	defer f.Close()

	// TODO: Log

	_, err = io.Copy(w, f)
	if err != nil {
		ErrorResponse(w, r, err, 500)
		return
	}
}
