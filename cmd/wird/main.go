package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/quadrifoglio/wir/api"
)

type config struct {
	Address      string
	DatabaseFile string
}

func readConfig(file string) (config, error) {
	var c config

	f, err := os.Open(file)
	if err != nil {
		return c, err
	}

	err = json.NewDecoder(f).Decode(&c)
	if err != nil {
		return c, err
	}

	return c, nil
}

func main() {
	log.Println("wird", api.Version)

	var confFile = flag.String("config", "etc/wird.json", "Configuration file to use")
	flag.Parse()

	config, err := readConfig(*confFile)
	if err != nil {
		log.Fatalf("Can not read configuration file: %s\n", err)
	}

	err = api.Start(config.DatabaseFile, config.Address)
	if err != nil {
		log.Fatalf("Can not start API: %s\n", err)
	}
}
