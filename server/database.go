package server

import (
	"database/sql"
	"fmt"

	"github.com/quadrifoglio/wir/shared"

	_ "github.com/mattn/go-sqlite3"
)

const (
	req = `
	CREATE TABLE IF NOT EXISTS image (
		id CHAR(8) NOT NULL UNIQUE PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		url VARCHAR(255) NOT NULL
	);
	`
)

var (
	DB *sql.DB
)

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

func CloseDatabase() error {
	return DB.Close()
}

func DBImageCreate(def shared.ImageDef) error {
	_, err := DB.Exec("INSERT INTO image VALUES (?, ?, ?)", def.ID, def.Name, def.URL)
	if err != nil {
		return err
	}

	return nil
}

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
		var url string

		err := rows.Scan(&id, &name, &url)
		if err != nil {
			return nil, err
		}

		images = append(images, shared.ImageDef{id, name, url})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return images, nil
}

func DBImageGet(id string) (shared.ImageDef, error) {
	var def shared.ImageDef

	rows, err := DB.Query("SELECT * FROM image WHERE id = ?", id)
	if err != nil {
		return def, err
	}

	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&def.ID, &def.Name, &def.URL)
		if err != nil {
			return def, err
		}
	}

	if err := rows.Err(); err != nil {
		return def, err
	}

	return def, fmt.Errorf("Image not found")
}

func DBImageUpdate(def shared.ImageDef) error {
	_, err := DB.Exec("UPDATE image set name = ?, url = ? WHERE id = ?", def.ID, def.Name, def.URL)
	if err != nil {
		return err
	}

	return nil
}

func DBImageDelete(id string) error {
	_, err := DB.Exec("DELETE FROM image WHERE id = ?", id)
	if err != nil {
		return err
	}

	return nil
}
