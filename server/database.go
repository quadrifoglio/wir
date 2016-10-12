package server

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/quadrifoglio/wir/shared"

	_ "github.com/mattn/go-sqlite3"
)

const (
	req = `
	CREATE TABLE IF NOT EXISTS image (
		id CHAR(8) NOT NULL UNIQUE PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(255) NOT NULL,
		src VARCHAR(255) NOT NULL
	);

	CREATE TABLE IF NOT EXISTS network (
		id CHAR(8) NOT NULL UNIQUE PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		cidr VARCHAR(255) NOT NULL,
		gw VARCHAR(255) NOT NULL,

		dhcp_enabled BOOLEAN NOT NULL,
		dhcp_start VARCHAR(255) NOT NULL,
		dhcp_num INTEGER NOT NULL,
		dhcp_router VARCHAR(255) NOT NULL
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

// IMAGES

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
	_, err := DB.Exec("INSERT INTO image VALUES (?, ?, ?, ?)", def.ID, def.Name, def.Type, def.Source)
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
		var t string
		var src string

		err := rows.Scan(&id, &name, &t, &src)
		if err != nil {
			return nil, err
		}

		images = append(images, shared.ImageDef{id, name, t, src})
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
		err := rows.Scan(&def.ID, &def.Name, &def.Type, &def.Source)
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
	_, err := DB.Exec("UPDATE image SET name = ?, type = ?, src = ? WHERE id = ?", def.Name, def.Type, def.Source, def.ID)
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

// NETWORKS

// DBNetworkExists checks if the specified network ID
// exists in the database
func DBNetworkExists(id string) bool {
	rows, err := DB.Query("SELECT id FROM network WHERE id = ? LIMIT 1", id)
	if err != nil {
		log.Println("Network exists check:", err)
		return false
	}

	defer rows.Close()

	if rows.Next() {
		return true
	}

	return false
}

// DBNetworkCreate creates a new network in the database
// using the specified definition
func DBNetworkCreate(def *shared.NetworkDef) error {
	_, err := DB.Exec(
		"INSERT INTO network VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		def.ID,
		def.Name,
		def.CIDR,
		def.GatewayIface,
		def.DHCP.Enabled,
		def.DHCP.StartIP,
		def.DHCP.NumIP,
		def.DHCP.Router,
	)

	if err != nil {
		return err
	}

	return nil
}

// DBNetworkList returns all the networks
// stored in the database
func DBNetworkList() ([]shared.NetworkDef, error) {
	rows, err := DB.Query("SELECT * FROM network")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	networks := make([]shared.NetworkDef, 0)
	for rows.Next() {
		var def shared.NetworkDef

		err := rows.Scan(
			&def.ID,
			&def.Name,
			&def.CIDR,
			&def.GatewayIface,
			&def.DHCP.Enabled,
			&def.DHCP.StartIP,
			&def.DHCP.NumIP,
			&def.DHCP.Router,
		)

		if err != nil {
			return nil, err
		}

		networks = append(networks, def)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return networks, nil
}

// DBNetworkGet returns the requested network
// from the database
func DBNetworkGet(id string) (shared.NetworkDef, error) {
	var def shared.NetworkDef

	rows, err := DB.Query("SELECT * FROM network WHERE id = ?", id)
	if err != nil {
		return def, err
	}

	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(
			&def.ID,
			&def.Name,
			&def.CIDR,
			&def.GatewayIface,
			&def.DHCP.Enabled,
			&def.DHCP.StartIP,
			&def.DHCP.NumIP,
			&def.DHCP.Router,
		)

		if err != nil {
			return def, err
		}

		return def, nil
	}

	if err := rows.Err(); err != nil {
		return def, err
	}

	return def, fmt.Errorf("Network not found")
}

// DBNetworkUpdate replaces all the values of the specified network
// with the new ones
func DBNetworkUpdate(def *shared.NetworkDef) error {
	sqls := `
		UPDATE network SET
			name = ?, cidr = ?, gw = ?,
			dhcp_enabled = ?, dhcp_start = ?,
			dhcp_num = ?, dhcp_router = ?
		WHERE id = ?
	`

	_, err := DB.Exec(sqls,
		def.Name,
		def.CIDR,
		def.GatewayIface,
		def.DHCP.Enabled,
		def.DHCP.StartIP,
		def.DHCP.NumIP,
		def.DHCP.Router,
		def.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

// DBNetworkDelete deletes the specified network
// from the database
func DBNetworkDelete(id string) error {
	_, err := DB.Exec("DELETE FROM network WHERE id = ?", id)
	if err != nil {
		return err
	}

	return nil
}
