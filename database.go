package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"sync"
)

var (
	Database      *sql.DB
	DatabaseMutex = &sync.Mutex{}
)

func DatabaseOpen() error {
	DatabaseMutex.Lock()
	defer DatabaseMutex.Unlock()

	db, err := sql.Open("sqlite3", "wird.db")
	if err != nil {
		return err
	}

	sql := `
	CREATE TABLE IF NOT EXISTS vm (
		vmid INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		vmbackend CHAR(50) NOT NULL,
		vmcores INTEGER NOT NULL,
		vmmem INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS vm_attr (
		attrvm INTEGER NOT NULL REFERENCES vm(vmid),
		attrkey CHAR(50) NOT NULL,
		attrval CHAR(50) NOT NULL,
		PRIMARY KEY(attrvm, attrkey)
	);
	CREATE TABLE IF NOT EXISTS vm_drive (
		driveid INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		drivevm INTEGER NOT NULL REFERENCES vm(vmid),
		drivetype CHAR(10) NOT NULL,
		drivefile CHAR(1024) NOT NULL
	);`

	_, err = db.Exec(sql)
	if err != nil {
		return err
	}

	Database = db
	return nil
}

func DatabaseInsertVm(vm *Vm) error {
	DatabaseMutex.Lock()
	defer DatabaseMutex.Unlock()

	_, err := Database.Exec("INSERT INTO vm (vmbackend, vmcores, vmmem) VALUES (?, ?, ?)", vm.Params.Backend, vm.Params.Cores, vm.Params.Memory)
	if err != nil {
		return err
	}

	for _, d := range vm.Params.Drives {
		_, err := Database.Exec("INSERT INTO vm_drive (drivevm, drivetype, drivefile) VALUES (?, ?, ?)", vm.ID, d.Type, d.File)
		if err != nil {
			return err
		}
	}

	return nil
}

func DatabaseListVms() ([]Vm, error) {
	DatabaseMutex.Lock()
	defer DatabaseMutex.Unlock()

	res, err := Database.Query("SELECT * FROM vm")
	if err != nil {
		return nil, err
	}

	defer res.Close()

	var vms = make([]Vm, 0)

	for res.Next() {
		var vm Vm
		vm.Attrs = make(map[string]string)
		vm.Params.Drives = make([]VmDrive, 0)

		err = res.Scan(&vm.ID, &vm.Params.Backend, &vm.Params.Cores, &vm.Params.Memory)
		if err != nil {
			return nil, err
		}

		res1, err := Database.Query("SELECT attrkey, attrval FROM vm_attr WHERE attrvm = ?", vm.ID)
		if err != nil {
			return nil, err
		}

		defer res1.Close()

		for res1.Next() {
			var key string
			var val string

			err = res1.Scan(&key, &val)
			if err != nil {
				return nil, err
			}

			vm.Attrs[key] = val
		}

		res2, err := Database.Query("SELECT drivetype, drivefile FROM vm_drive WHERE drivevm = ?", vm.ID)
		if err != nil {
			return nil, err
		}

		defer res2.Close()

		for res2.Next() {
			var d VmDrive

			err = res2.Scan(&d.Type, &d.File)
			if err != nil {
				return nil, err
			}

			vm.Params.Drives = append(vm.Params.Drives, d)
		}

		vms = append(vms, vm)
	}

	return vms, nil
}

func DatabaseGetVmByID(id int) (Vm, error) {
	DatabaseMutex.Lock()
	defer DatabaseMutex.Unlock()

	var vm Vm
	vm.Attrs = make(map[string]string, 0)
	vm.Params.Drives = make([]VmDrive, 0)

	res, err := Database.Query("SELECT * FROM vm WHERE vmid = ?", id)
	if err != nil {
		return vm, err
	}

	defer res.Close()

	if res.Next() {
		err = res.Scan(&vm.ID, &vm.Params.Backend, &vm.Params.Cores, &vm.Params.Memory)
		if err != nil {
			return vm, err
		}

		res1, err := Database.Query("SELECT attrkey, attrval FROM vm_attr WHERE attrvm = ?", vm.ID)
		if err != nil {
			return vm, err
		}

		defer res1.Close()

		for res1.Next() {
			var key string
			var val string

			err = res1.Scan(&key, &val)
			if err != nil {
				return vm, err
			}

			vm.Attrs[key] = val
		}

		res2, err := Database.Query("SELECT drivetype, drivefile FROM vm_drive WHERE drivevm = ?", vm.ID)
		if err != nil {
			return vm, err
		}

		defer res2.Close()

		for res2.Next() {
			var d VmDrive

			err = res2.Scan(&d.Type, &d.File)
			if err != nil {
				return vm, err
			}

			vm.Params.Drives = append(vm.Params.Drives, d)
		}
	}

	return vm, nil
}

func DatabaseSetVmAttr(vm *Vm, key, value string) error {
	DatabaseMutex.Lock()
	defer DatabaseMutex.Unlock()

	_, err := Database.Exec("INSERT OR REPLACE INTO vm_attr (attrvm, attrkey, attrval) VALUES (?, ?, ?)", vm.ID, key, value)
	if err != nil {
		return err
	}

	return nil
}
