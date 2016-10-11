package server

import (
	"fmt"
	"net/http"
)

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world !")
}
