package server

import (
	"fmt"
	"net/http"
)

func HandleImageCreate(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Image create")
}

func HandleImageList(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Image list")
}

func HandleImageGet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Image get")
}

func HandleImageUpdate(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Image update")
}

func HandleImageDelete(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Image delete")
}
