package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/quadrifoglio/wir/errors"
)

type Response struct {
	Status  int
	Message string
	Content interface{} `json:",omitempty"`
}

func SuccessResponse(c interface{}) Response {
	return Response{200, "Success", c}
}

func ErrorResponse(err error) Response {
	var re Response
	re.Message = err.Error()
	re.Content = nil

	switch err {
	case errors.NotFound:
		re.Status = 404
		break
	case errors.BadRequest:
		re.Status = 400
		break
	case errors.InvalidImageType:
		re.Status = 400
		break
	case errors.InvalidURL:
		re.Status = 400
		break
	case errors.UnsupportedProto:
		re.Status = 400
		break
	case errors.ImageNotFound:
		re.Status = 404
		break
	default:
		re.Status = 500
		break
	}

	return re
}

func (re Response) Send(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(re)
	if err != nil {
		re.Status = 500
		re.Message = err.Error()

		w.WriteHeader(re.Status)
	}

	if re.Status != 200 {
		log.Printf("%s %s from %s - %s\n", r.Method, r.URL.String(), r.RemoteAddr, re.Message)
	}
}
