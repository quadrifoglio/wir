package utils

import (
	"fmt"
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

// FileSize returns the size of
// the specified file in bytes
func FileSize(path string) (uint64, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return uint64(stat.Size()), nil
}

// ReplaceFileContents replaces the content of the specified file
// byte the data provided in 'data'
func ReplaceFileContents(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return fmt.Errorf("rewrite %s: %s", path, err)
	}

	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("rewrite %s: %s", path, err)
	}

	return nil
}
