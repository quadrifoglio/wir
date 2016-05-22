package api

import (
	"encoding/json"
	"net/http"

	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/image"
)

type ImagePost struct {
	Name   string
	Type   int
	Source string
}

func handleImageCreate(w http.ResponseWriter, r *http.Request) {
	var i ImagePost

	err := json.NewDecoder(r.Body).Decode(&i)
	if err != nil {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	if len(i.Source) <= 0 {
		ErrorResponse(errors.BadRequest).Send(w, r)
		return
	}

	switch i.Type {
	case image.TypeQemu:
		err = image.QemuFetch(i.Name, i.Source, Conf.ImagePath)
		break
	default:
		err = errors.InvalidImageType
		break
	}

	if err != nil {
		ErrorResponse(err).Send(w, r)
		return
	}

	SuccessResponse(nil).Send(w, r)
}
