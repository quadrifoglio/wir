package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/quadrifoglio/wir/errors"
)

const (
	Version = "0.0.1"
)

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	ErrorResponse(errors.NotFound).Send(w, r)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	type info struct {
		Name    string
		Version string
	}

	i := info{"wir api", Version}
	SuccessResponse(i).Send(w, r)
}

func Start(dbf string, addr string) error {
	r := mux.NewRouter()

	r.HandleFunc("/", handleIndex)
	r.HandleFunc("/image", handleImageCreate).Methods("POST")

	r.NotFoundHandler = http.HandlerFunc(handleNotFound)
	http.Handle("/", r)

	return http.ListenAndServe(addr, nil)
}
