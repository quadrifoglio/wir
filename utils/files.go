package utils

import (
	"log"
	"os"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Println("file exists check: %s", err)
		}

		return false
	}

	return true
}
