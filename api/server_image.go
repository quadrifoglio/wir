package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/utils"
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

	if !image.TypeExists(i.Type) {
		ErrorResponse(errors.InvalidImageType).Send(w, r)
		return
	}

	url, err := url.Parse(i.Source)
	if err != nil {
		ErrorResponse(errors.InvalidURL).Send(w, r)
		return
	}

	_, err = utils.FetchResource(url)
	if err != nil {
		ErrorResponse(err).Send(w, r)
	}
}
