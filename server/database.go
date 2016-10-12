package server

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/utils"

	_ "github.com/mattn/go-sqlite3"
)

const (
	req = `
	CREATE TABLE IF NOT EXISTS image (
		id CHAR(8) NOT NULL UNIQUE PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		src VARCHAR(255) NOT NULL
	);
	`
)

var (
	DB *sql.DB
)

// InitDatabase opens the specified SQLite database
// and creates the tables if they don't exist
func InitDatabase(path string) error {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return err
	}

	_, err = db.Exec(req)
	if err != nil {
		return err
	}

	DB = db
	return nil
}

// CloseDatabase closes the database
func CloseDatabase() error {
	return DB.Close()
}

// DBImageExists checks if the specified image ID
// exists in the database
func DBImageExists(id string) bool {
	rows, err := DB.Query("SELECT id FROM image WHERE id = ? LIMIT 1", id)
	if err != nil {
		log.Println("Image exists check:", err)
		return false
	}

	defer rows.Close()

	if rows.Next() {
		return true
	}

	return false
}

// DBImageCreate creates a new image in the database
// using the specified definition
func DBImageCreate(def *shared.ImageDef) error {
	for {
		def.ID = utils.RandID(GlobalNodeID)
		if !DBImageExists(def.ID) {
			break
		}
	}

	_, err := DB.Exec("INSERT INTO image VALUES (?, ?, ?)", def.ID, def.Name, def.Source)
	if err != nil {
		return err
	}

	return nil
}

// DBImageList returns all the images
// stored in the database
func DBImageList() ([]shared.ImageDef, error) {
	rows, err := DB.Query("SELECT * FROM image")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	images := make([]shared.ImageDef, 0)
	for rows.Next() {
		var id string
		var name string
		var src string

		err := rows.Scan(&id, &name, &src)
		if err != nil {
			return nil, err
		}

		images = append(images, shared.ImageDef{id, name, src})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return images, nil
}

// DBImageGet returns the requested image
// from the database
func DBImageGet(id string) (shared.ImageDef, error) {
	var def shared.ImageDef

	rows, err := DB.Query("SELECT * FROM image WHERE id = ?", id)
	if err != nil {
		return def, err
	}

	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&def.ID, &def.Name, &def.Source)
		if err != nil {
			return def, err
		}

		return def, nil
	}

	if err := rows.Err(); err != nil {
		return def, err
	}

	return def, fmt.Errorf("Image not found")
}

// DBImageUpdate replaces all the values of the specified image
// with the new ones
func DBImageUpdate(def *shared.ImageDef) error {
	_, err := DB.Exec("UPDATE image SET name = ?, src = ? WHERE id = ?", def.Name, def.Source, def.ID)
	if err != nil {
		return err
	}

	return nil
}

// DBImageDelete deletes the specified image
// from the database
func DBImageDelete(id string) error {
	_, err := DB.Exec("DELETE FROM image WHERE id = ?", id)
	if err != nil {
		return err
	}

	return nil
}
