package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/shared"
)

func handleImageCreate(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	var img shared.Image

	err := json.NewDecoder(r.Body).Decode(&img)
	if err != nil {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	if len(img.Source) <= 0 || len(img.Name) <= 0 {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	err = CreateImage(&img)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = DBStoreImage(img)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(img).Send(w, r)
}

func handleImageList(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	imgs, err := DBListImages()
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(imgs).Send(w, r)
}

func handleImageGet(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	vars := mux.Vars(r)
	name := vars["name"]

	img, err := DBGetImage(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(img).Send(w, r)
}

func handleImageDelete(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	vars := mux.Vars(r)
	name := vars["name"]

	err := DBDeleteImage(name)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(nil).Send(w, r)
}
