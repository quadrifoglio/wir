package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/quadrifoglio/wir/client"
	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/image"
)

func handleImageCreate(w http.ResponseWriter, r *http.Request) {
	PrepareResponse(w, r)

	var i client.ImageRequest

	err := json.NewDecoder(r.Body).Decode(&i)
	if err != nil {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	if len(i.Source) <= 0 {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	img, err := image.Create(i.Type, i.Name, i.Source, i.Arch, i.Distro, i.Release, i.MainPartition)
	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	err = DBStoreImage(&img)
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
