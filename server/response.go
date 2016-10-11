package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// SuccessResponse responds to the HTTP request with the requested data
// encoded in JSON
func SuccessResponse(w http.ResponseWriter, r *http.Request, v interface{}) {
	req := fmt.Sprintf("%s %s from %s", r.Method, r.URL, r.RemoteAddr)

	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "{\"Error\": \"Failed to encode response\"}")

		log.Printf("%s - Failed to encore response: %s\n", req, err)
		return
	}

	log.Printf("%s - 200 OK\n", req)
}

// ErrorResponse responds to the HTTP request with the error and
// logs it to the system
func ErrorResponse(w http.ResponseWriter, r *http.Request, err error, status int) {
	req := fmt.Sprintf("%s %s from %s", r.Method, r.URL, r.RemoteAddr)

	w.WriteHeader(status)
	fmt.Fprintf(w, "{\"Error\": \"%s\"}", err)

	log.Printf("%s - %s\n", req, err)
}
