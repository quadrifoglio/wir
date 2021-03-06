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
		name VARCHAR(12) NOT NULL UNIQUE PRIMARY KEY,
		cidr VARCHAR(255) NOT NULL,
		gw VARCHAR(255) NOT NULL,

		dhcp_enabled BOOLEAN NOT NULL,
		dhcp_start VARCHAR(255),
		dhcp_num INTEGER,
		dhcp_router VARCHAR(255)
	);

	CREATE TABLE IF NOT EXISTS volume (
		id CHAR(8) NOT NULL UNIQUE PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(255) NOT NULL,
		size BIGINT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS machine (
		id CHAR(8) NOT NULL UNIQUE PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		img CHAR(8) NOT NULL REFERENCES image(id),
		cores INTEGER NOT NULL,
		mem BIGINT NOT NULL,
		disk BIGINT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS iface (
		machine CHAR(8) NOT NULL REFERENCES machine(id),
		net VARCHAR(255) NOT NULL,
		mac VARCHAR(255) NOT NULL,
		ip VARCHAR(255)
	);

	CREATE TABLE IF NOT EXISTS attach (
		machine CHAR(8) NOT NULL REFERENCES machine(id),
		volume CHAR(8) NOT NULL REFERENCES volume(id)
	);

	CREATE TABLE IF NOT EXISTS kvm_opt (
		machine CHAR(8) NOT NULL UNIQUE REFERENCES machine(id),
		pid INTEGER,
		cd VARCHAR(255),
		vnc_enabled BOOLEAN NOT NULL,
		vnc_addr VARCHAR(255),
		vnc_port INTEGER,
		vnc_ws_port INTEGER,
		vnc_passwd VARCHAR(255)
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
func DBImageCreate(def shared.ImageDef) error {
	_, err := DB.Exec("INSERT INTO image VALUES (?, ?, ?, ?)", def.ID, def.Name, def.Type, def.Source)
	if err != nil {
		return err
	}

	return nil
}

// DBImageFetch fetches a corresponding data structure
// from the database
func DBImageFetch(rows *sql.Rows) (shared.ImageDef, error) {
	var def shared.ImageDef

	err := rows.Scan(
		&def.ID,
		&def.Name,
		&def.Type,
		&def.Source,
	)

	return def, err
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
		def, err := DBImageFetch(rows)
		if err != nil {
			return nil, err
		}

		images = append(images, def)
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
		def, err := DBImageFetch(rows)
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
func DBImageUpdate(def shared.ImageDef) error {
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

// DBNetworkExists checks if the specified network name
// exists in the database
func DBNetworkExists(name string) bool {
	rows, err := DB.Query("SELECT name FROM network WHERE name = ? LIMIT 1", name)
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
func DBNetworkCreate(def shared.NetworkDef) error {
	_, err := DB.Exec(
		"INSERT INTO network VALUES (?, ?, ?, ?, ?, ?, ?)",
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

// DBNetworkFetch fetches a corresponding data structure
// from the database
func DBNetworkFetch(rows *sql.Rows) (shared.NetworkDef, error) {
	var def shared.NetworkDef

	err := rows.Scan(
		&def.Name,
		&def.CIDR,
		&def.GatewayIface,
		&def.DHCP.Enabled,
		&def.DHCP.StartIP,
		&def.DHCP.NumIP,
		&def.DHCP.Router,
	)

	return def, err
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
		def, err := DBNetworkFetch(rows)
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
func DBNetworkGet(name string) (shared.NetworkDef, error) {
	var def shared.NetworkDef

	rows, err := DB.Query("SELECT * FROM network WHERE name = ?", name)
	if err != nil {
		return def, err
	}

	defer rows.Close()

	if rows.Next() {
		def, err = DBNetworkFetch(rows)
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
func DBNetworkUpdate(def shared.NetworkDef) error {
	sqls := `
		UPDATE network SET
			cidr = ?, gw = ?,
			dhcp_enabled = ?, dhcp_start = ?,
			dhcp_num = ?, dhcp_router = ?
		WHERE name = ?
	`

	_, err := DB.Exec(sqls,
		def.CIDR,
		def.GatewayIface,
		def.DHCP.Enabled,
		def.DHCP.StartIP,
		def.DHCP.NumIP,
		def.DHCP.Router,
		def.Name,
	)

	if err != nil {
		return err
	}

	return nil
}

// DBNetworkDelete deletes the specified network
// from the database
func DBNetworkDelete(name string) error {
	_, err := DB.Exec("DELETE FROM network WHERE name = ?", name)
	if err != nil {
		return err
	}

	return nil
}

// VOLUMES

// DBVolumeExists checks if the specified volume ID
// exists in the database
func DBVolumeExists(id string) bool {
	rows, err := DB.Query("SELECT id FROM volume WHERE id = ? LIMIT 1", id)
	if err != nil {
		log.Println("Volume exists check:", err)
		return false
	}

	defer rows.Close()

	if rows.Next() {
		return true
	}

	return false
}

// DBVolumeCreate creates a new volume in the database
// using the specified definition
func DBVolumeCreate(def shared.VolumeDef) error {
	_, err := DB.Exec(
		"INSERT INTO volume VALUES (?, ?, ?, ?)",
		def.ID,
		def.Name,
		def.Type,
		def.Size,
	)

	if err != nil {
		return err
	}

	return nil
}

// DBVolumeFetch fetches a corresponding data structure
// from the database
func DBVolumeFetch(rows *sql.Rows) (shared.VolumeDef, error) {
	var def shared.VolumeDef

	err := rows.Scan(
		&def.ID,
		&def.Name,
		&def.Type,
		&def.Size,
	)

	return def, err
}

// DBVolumeList returns all the volumes
// stored in the database
func DBVolumeList() ([]shared.VolumeDef, error) {
	rows, err := DB.Query("SELECT * FROM volume")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	volumes := make([]shared.VolumeDef, 0)
	for rows.Next() {
		def, err := DBVolumeFetch(rows)
		if err != nil {
			return nil, err
		}

		volumes = append(volumes, def)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return volumes, nil
}

// DBVolumeGet returns the requested volume
// from the database
func DBVolumeGet(id string) (shared.VolumeDef, error) {
	var def shared.VolumeDef

	rows, err := DB.Query("SELECT * FROM volume WHERE id = ?", id)
	if err != nil {
		return def, err
	}

	defer rows.Close()

	if rows.Next() {
		def, err = DBVolumeFetch(rows)
		if err != nil {
			return def, err
		}

		return def, nil
	}

	if err := rows.Err(); err != nil {
		return def, err
	}

	return def, fmt.Errorf("Volume not found")
}

// DBVolumeUpdate replaces all the values of the specified volume
// with the new ones
func DBVolumeUpdate(def shared.VolumeDef) error {
	_, err := DB.Exec("UPDATE volume SET name = ?, type = ?, size = ? WHERE id = ?",
		def.Name,
		def.Type,
		def.Size,
		def.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

// DBVolumeDelete deletes the specified volume
// from the database
func DBVolumeDelete(id string) error {
	_, err := DB.Exec("DELETE FROM volume WHERE id = ?", id)
	if err != nil {
		return err
	}

	return nil
}

// MACHINES

// DBMachineExists checks if the specified machine ID
// exists in the database
func DBMachineExists(id string) bool {
	rows, err := DB.Query("SELECT id FROM machine WHERE id = ? LIMIT 1", id)
	if err != nil {
		log.Println("Machine exists check:", err)
		return false
	}

	defer rows.Close()

	if rows.Next() {
		return true
	}

	return false
}

// DBMachineSetVolumes flushes the volumes associated with the machine,
// and updates them
func DBMachineSetVolumes(def shared.MachineDef) error {
	_, err := DB.Exec("DELETE FROM attach WHERE machine = ?", def.ID)
	if err != nil {
		return err
	}

	for _, v := range def.Volumes {
		_, err := DB.Exec("INSERT INTO attach VALUES (?, ?)", def.ID, v)
		if err != nil {
			return err
		}
	}

	return nil
}

// DBMachineSetInterfaces flushes the volumes associated with the machine,
// and updates them
func DBMachineSetInterfaces(def shared.MachineDef) error {
	_, err := DB.Exec("DELETE FROM iface WHERE machine = ?", def.ID)
	if err != nil {
		return err
	}

	for _, i := range def.Interfaces {
		_, err := DB.Exec("INSERT INTO iface VALUES (?, ?, ?, ?)", def.ID, i.Network, i.MAC, i.IP)
		if err != nil {
			return err
		}
	}

	return nil
}

// DBMachineGetVolumes returns the list of volume IDs
// associated with the machine
func DBMachineGetVolumes(id string) ([]string, error) {
	rows, err := DB.Query("SELECT volume FROM attach WHERE machine = ?", id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	vols := make([]string, 0)
	for rows.Next() {
		var v string

		err := rows.Scan(&v)
		if err != nil {
			return nil, err
		}

		vols = append(vols, v)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return vols, nil
}

// DBMachineSetKvmOpts saves the KVM-specific options of
// the machine into the database
func DBMachineSetKvmOpts(id string, def shared.KvmOptsDef) error {
	if !DBMachineExists(id) {
		return fmt.Errorf("Machine not found")
	}

	// The PID should not me modified
	if def.PID == -1 {
		opts, err := DBMachineGetKvmOpts(id)
		if err != nil {
			return err
		}

		def.PID = opts.PID
	}

	_, err := DB.Exec(
		"INSERT OR REPLACE INTO kvm_opt VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		id,
		def.PID,
		def.CDRom,
		def.VNC.Enabled,
		def.VNC.Address,
		def.VNC.Port,
		def.VNC.WebsocketPort,
		def.VNC.Password,
	)

	if err != nil {
		return err
	}

	return nil
}

// DBMachineGetKvmOpts retreives the KVM-specific options of
// the machine as a data structure
func DBMachineGetKvmOpts(id string) (shared.KvmOptsDef, error) {
	var def shared.KvmOptsDef

	rows, err := DB.Query("SELECT * FROM kvm_opt WHERE machine = ? LIMIT 1", id)
	if err != nil {
		return def, err
	}

	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&id, &def.PID, &def.CDRom, &def.VNC.Enabled, &def.VNC.Address, &def.VNC.Port, &def.VNC.WebsocketPort, &def.VNC.Password)
		if err != nil {
			return def, err
		}

		return def, nil
	}

	return def, fmt.Errorf("KVM options not found")
}

// DBMachineGetInterfaces returns the details of the interfaces
// associated with the machine
func DBMachineGetInterfaces(id string) ([]shared.InterfaceDef, error) {
	rows, err := DB.Query("SELECT * FROM iface WHERE machine = ?", id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	ifaces := make([]shared.InterfaceDef, 0)
	for rows.Next() {
		var id string
		var i shared.InterfaceDef

		err := rows.Scan(&id, &i.Network, &i.MAC, &i.IP)
		if err != nil {
			return nil, err
		}

		ifaces = append(ifaces, i)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ifaces, nil
}

// DBMachineCreate creates a new machine in the database
// using the specified definition
func DBMachineCreate(def shared.MachineDef) error {
	_, err := DB.Exec(
		"INSERT INTO machine VALUES (?, ?, ?, ?, ?, ?)",
		def.ID,
		def.Name,
		def.Image,
		def.Cores,
		def.Memory,
		def.Disk,
	)

	if err != nil {
		return err
	}

	if err := DBMachineSetVolumes(def); err != nil {
		return err
	}
	if err := DBMachineSetInterfaces(def); err != nil {
		return err
	}

	var opts shared.KvmOptsDef

	err = DBMachineSetKvmOpts(def.ID, opts)
	if err != nil {
		return err
	}

	return nil
}

// DBMachineFetch fetches a corresponding data structure
// from the database
func DBMachineFetch(rows *sql.Rows) (shared.MachineDef, error) {
	var def shared.MachineDef

	err := rows.Scan(
		&def.ID,
		&def.Name,
		&def.Image,
		&def.Cores,
		&def.Memory,
		&def.Disk,
	)

	if err != nil {
		return def, err
	}

	def.Volumes, err = DBMachineGetVolumes(def.ID)
	if err != nil {
		return def, err
	}

	def.Interfaces, err = DBMachineGetInterfaces(def.ID)
	if err != nil {
		return def, err
	}

	return def, nil
}

// DBMachineList returns all the machines
// stored in the database
func DBMachineList() ([]shared.MachineDef, error) {
	rows, err := DB.Query("SELECT * FROM machine")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	machines := make([]shared.MachineDef, 0)
	for rows.Next() {
		def, err := DBMachineFetch(rows)
		if err != nil {
			return nil, err
		}

		machines = append(machines, def)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return machines, nil
}

// DBMachineListOnNetwork finds the machines associated with
// the specified network
func DBMachineListOnNetwork(netw string) ([]shared.MachineDef, error) {
	sqls := `
		SELECT machine.* FROM machine
		INNER JOIN iface ON machine.id = iface.machine
		WHERE iface.net = ?
	`

	rows, err := DB.Query(sqls, netw)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	machines := make([]shared.MachineDef, 0)
	for rows.Next() {
		def, err := DBMachineFetch(rows)
		if err != nil {
			return nil, err
		}

		machines = append(machines, def)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return machines, nil
}

// DBMachineGet returns the requested machine
// from the database
func DBMachineGet(id string) (shared.MachineDef, error) {
	var def shared.MachineDef

	rows, err := DB.Query("SELECT * FROM machine WHERE id = ?", id)
	if err != nil {
		return def, err
	}

	defer rows.Close()

	if rows.Next() {
		def, err = DBMachineFetch(rows)
		if err != nil {
			return def, err
		}

		return def, nil
	}

	if err := rows.Err(); err != nil {
		return def, err
	}

	return def, fmt.Errorf("Machine not found")
}

// DBMachineGetByMAC finds the machine associated with
// the specified MAC address
func DBMachineGetByMAC(mac string) (shared.MachineDef, error) {
	var def shared.MachineDef

	sqls := `
		SELECT * FROM machine WHERE id = (
			SELECT machine FROM iface WHERE mac = ? LIMIT 1
		) LIMIT 1
	`

	rows, err := DB.Query(sqls, mac)
	if err != nil {
		return def, err
	}

	defer rows.Close()

	if rows.Next() {
		def, err := DBMachineFetch(rows)
		if err != nil {
			return def, err
		}

		return def, nil
	}

	if err := rows.Err(); err != nil {
		return def, err
	}

	return def, fmt.Errorf("Machine not found")
}

// DBMachineUpdate replaces all the values of the specified machine
// with the new ones
func DBMachineUpdate(def shared.MachineDef) error {
	sqls := `
		UPDATE machine SET
		name = ?, cores = ?, mem = ?
		WHERE id = ?
	`
	_, err := DB.Exec(sqls,
		def.Name,
		def.Cores,
		def.Memory,
		def.ID,
	)

	if err != nil {
		return err
	}

	if err := DBMachineSetVolumes(def); err != nil {
		return err
	}
	if err := DBMachineSetInterfaces(def); err != nil {
		return err
	}

	return nil
}

// DBMachineDelete deletes the specified machine
// from the database
func DBMachineDelete(id string) error {
	_, err := DB.Exec("DELETE FROM machine WHERE id = ?", id)
	if err != nil {
		return err
	}

	_, err = DB.Exec("DELETE FROM attach WHERE machine = ?", id)
	if err != nil {
		return err
	}

	_, err = DB.Exec("DELETE FROM iface WHERE machine = ?", id)
	if err != nil {
		return err
	}

	_, err = DB.Exec("DELETE FROM kvm_opt WHERE machine = ?", id)
	if err != nil {
		return err
	}

	return nil
}

// MISC

// DBIsMACFree checks if the specified MAC address
// is in use on an interface
func DBIsMACFree(mac string) bool {
	rows, err := DB.Query("SELECT mac FROM iface WHERE mac = ?", mac)
	if err != nil {
		log.Println("Free MAC check: ", err)
		return false
	}

	defer rows.Close()

	if rows.Next() {
		return false
	}

	return true
}
