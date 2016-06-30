package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting web server...")

	var addr = flag.String("address", "127.0.0.1:3000", "Address to listen on")
	flag.Parse()

	static := http.FileServer(http.Dir("webroot"))
	http.Handle("/", static)

	log.Fatal(http.ListenAndServe(*addr, nil))
}
