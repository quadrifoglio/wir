package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var (
	Database *sql.DB
)

func DatabaseOpen() error {
	db, err := sql.Open("sqlite3", "wird.db")
	if err != nil {
		return err
	}

	sql := `
	CREATE TABLE IF NOT EXISTS vm (
		vmid INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		vmbackend CHAR(50) NOT NULL,
		vmstate INTEGER NOT NULL
	)`

	_, err = db.Exec(sql)
	if err != nil {
		return err
	}

	Database = db
	return nil
}

func DatabaseInsertVm(vm *Vm) error {
	stmt, err := Database.Prepare("INSERT INTO vm (vmbackend, vmstate) VALUES (?, ?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(vm.Params.Backend, StateDown)
	if err != nil {
		return err
	}

	return nil
}

func DatabaseListVms() ([]Vm, error) {
	res, err := Database.Query("SELECT vmid, vmbackend, vmstate FROM vm")
	if err != nil {
		return nil, err
	}

	var vms = make([]Vm, 0)

	for res.Next() {
		var vm Vm

		err = res.Scan(&vm.ID, &vm.Params.Backend, &vm.State)
		if err != nil {
			return nil, err
		}

		vms = append(vms, vm)
	}

	return vms, nil
}

func DatabaseGetVmByID(id int) (Vm, error) {
	var vm Vm

	stmt, err := Database.Prepare("SELECT vmid, vmbackend, vmstate FROM vm WHERE vmid = ?")
	if err != nil {
		return vm, err
	}

	res, err := stmt.Query(id)
	if err != nil {
		return vm, err
	}

	if res.Next() {
		err = res.Scan(&vm.ID, &vm.Params.Backend, &vm.State)
		if err != nil {
			return vm, err
		}
	}

	return vm, nil
}
