package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
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
		vmstate INTEGER NOT NULL,
		vmcores INTEGER NOT NULL,
		vmmem INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS vm_attr (
		attrvm INTEGER NOT NULL REFERENCES vm(vmid),
		attrkey CHAR(50) NOT NULL,
		attrval CHAR(50) NOT NULL,
		PRIMARY KEY(attrvm, attrkey)
	);`

	_, err = db.Exec(sql)
	if err != nil {
		return err
	}

	Database = db
	return nil
}

func DatabaseInsertVm(vm *Vm) error {
	stmt, err := Database.Prepare("INSERT INTO vm (vmbackend, vmstate, vmcores, vmmem) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(vm.Params.Backend, StateDown, vm.Params.Cores, vm.Params.Memory)
	if err != nil {
		return err
	}

	return nil
}

func DatabaseListVms() ([]Vm, error) {
	res, err := Database.Query("SELECT * FROM vm")
	if err != nil {
		return nil, err
	}

	var vms = make([]Vm, 0)

	for res.Next() {
		var vm Vm
		vm.Attrs = make(map[string]string)

		err = res.Scan(&vm.ID, &vm.Params.Backend, &vm.State, &vm.Params.Cores, &vm.Params.Memory)
		if err != nil {
			return nil, err
		}

		res1, err := Database.Query("SELECT attrkey, attrval FROM vm_attr WHERE attrvm = ?", vm.ID)
		if err != nil {
			return nil, err
		}

		for res1.Next() {
			var key string
			var val string

			err = res1.Scan(&key, &val)
			if err != nil {
				return nil, err
			}

			vm.Attrs[key] = val
		}

		vms = append(vms, vm)
	}

	return vms, nil
}

func DatabaseGetVmByID(id int) (Vm, error) {
	var vm Vm

	stmt, err := Database.Prepare("SELECT * FROM vm WHERE vmid = ?")
	if err != nil {
		return vm, err
	}

	res, err := stmt.Query(id)
	if err != nil {
		return vm, err
	}

	if res.Next() {
		err = res.Scan(&vm.ID, &vm.Params.Backend, &vm.State, &vm.Params.Cores, &vm.Params.Memory)
		if err != nil {
			return vm, err
		}

		res1, err := Database.Query("SELECT attrkey, attrval FROM vm_attr WHERE attrvm = ?", vm.ID)
		if err != nil {
			return vm, err
		}

		for res1.Next() {
			var key string
			var val string

			err = res1.Scan(&key, &val)
			if err != nil {
				return vm, err
			}

			vm.Attrs[key] = val
		}
	}

	return vm, nil
}

func DatabaseSetVmAttr(vm *Vm, key, value string) error {
	_, err := Database.Exec("INSERT INTO vm_attr (attrvm, attrkey, attrval) VALUES (?, ?, ?)", vm.ID, key, value)
	if err != nil {
		log.Println("mdr", key, value)
		return err
	}

	return nil
}
