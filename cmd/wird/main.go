package main

import (
	"flag"
	"log"

	"github.com/quadrifoglio/wir/api"
	"github.com/quadrifoglio/wir/config"
)

func main() {
	log.Println("wird", api.Version)

	var confFile = flag.String("config", "etc/wird.json", "Configuration file to use")
	flag.Parse()

	err := config.ReadAPIConfig(*confFile)
	if err != nil {
		log.Fatalf("Can not read configuration file: %s\n", err)
	}

	err = api.Start()
	if err != nil {
		log.Fatalf("Can not start API: %s\n", err)
	}
}
