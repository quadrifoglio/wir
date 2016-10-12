package utils

import (
	"log"
	"os"
	"path/filepath"
)

// FileExists returns wether the specified
// file is existing
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

// FileExtension returns the extension of the file
// or empty string if there isn't
func FileExtension(path string) string {
	return filepath.Ext(path)
}
