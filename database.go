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

func DatabaseOpen(file string) error {
	DatabaseMutex.Lock()
	defer DatabaseMutex.Unlock()

	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return err
	}

	sql := `
	CREATE TABLE IF NOT EXISTS image (
		imgid INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		imgname CHAR(50) NOT NULL UNIQUE,
		imgtype CHAR(10) NOT NULL,
		imgstate CHAR(10) NOT NULL,
		imgpath CHAR(1024) NOT NULL
	);
	CREATE TABLE IF NOT EXISTS vm (
		vmid INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
		vmbackend CHAR(10) NOT NULL,
		vmcores INTEGER NOT NULL,
		vmmem INTEGER NOT NULL,
		vmimg INTEGER NOT NULL REFERENCES images(imgid),
		vmnetbron CHAR(10) NOT NULL
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

func DatabaseFreeImageId() (int, error) {
	var id int

	res, err := Database.Query("SELECT seq FROM sqlite_sequence WHERE name='image'")
	if err != nil {
		return 0, err
	}

	defer res.Close()

	if res.Next() {
		err = res.Scan(&id)
		if err != nil {
			return 0, err
		}

		id++
	}

	if id == 0 {
		id = 1
	}

	return id, nil
}

func DatabaseInsertImage(img *Image) error {
	DatabaseMutex.Lock()
	defer DatabaseMutex.Unlock()

	res, err := Database.Exec("INSERT INTO image (imgtype, imgname, imgstate, imgpath) VALUES (?, ?, ?, ?)", img.Type, img.Name, img.State, img.Path)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	img.ID = int(id)
	return nil
}

func DatabaseUpdateImage(img *Image) error {
	DatabaseMutex.Lock()
	defer DatabaseMutex.Unlock()

	_, err := Database.Exec("INSERT OR REPLACE INTO image VALUES (?, ?, ?, ?, ?)", img.ID, img.Name, img.Type, img.State, img.Path)
	if err != nil {
		return err
	}

	return nil
}

func DatabaseListImages() ([]Image, error) {
	DatabaseMutex.Lock()
	defer DatabaseMutex.Unlock()

	res, err := Database.Query("SELECT * FROM image")
	if err != nil {
		return nil, err
	}

	defer res.Close()

	var imgs = make([]Image, 0)

	for res.Next() {
		var img Image

		err = res.Scan(&img.ID, &img.Name, &img.Type, &img.State, &img.Path)
		if err != nil {
			return imgs, err
		}

		imgs = append(imgs, img)
	}

	return imgs, nil
}

func DatabaseGetImage(id int) (Image, error) {
	DatabaseMutex.Lock()
	defer DatabaseMutex.Unlock()

	var img Image

	res, err := Database.Query("SELECT * FROM image WHERE imgid = ? LIMIT 1", id)
	if err != nil {
		return img, err
	}

	defer res.Close()

	if res.Next() {
		err = res.Scan(&img.ID, &img.Name, &img.Type, &img.State, &img.Path)
		if err != nil {
			return img, err
		}
	} else {
		return img, ErrImageNotFound
	}

	return img, nil
}

func DatabaseFreeVmId() (int, error) {
	var id int

	res, err := Database.Query("SELECT seq FROM sqlite_sequence WHERE name='vm'")
	if err != nil {
		return 0, err
	}

	defer res.Close()

	if res.Next() {
		err = res.Scan(&id)
		if err != nil {
			return 0, err
		}

		id++
	}

	if id == 0 {
		id = 1
	}

	return id, nil
}

func DatabaseInsertVm(vm *Vm) error {
	DatabaseMutex.Lock()
	defer DatabaseMutex.Unlock()

	res, err := Database.Exec("INSERT INTO vm (vmbackend, vmcores, vmmem, vmimg, vmnetbron) VALUES (?, ?, ?, ?, ?)", vm.Params.Backend, vm.Params.Cores, vm.Params.Memory, vm.Params.ImageID, vm.Params.NetBridgeOn)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	vm.ID = int(id)
	vm.State = VmStateDown

	for _, d := range vm.Drives {
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
		vm.Drives = make([]VmDrive, 0)

		err = res.Scan(&vm.ID, &vm.Params.Backend, &vm.Params.Cores, &vm.Params.Memory, &vm.Params.ImageID, &vm.Params.NetBridgeOn)
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

			vm.Drives = append(vm.Drives, d)
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
	vm.Drives = make([]VmDrive, 0)

	res, err := Database.Query("SELECT * FROM vm WHERE vmid = ? LIMIT 1", id)
	if err != nil {
		return vm, err
	}

	defer res.Close()

	if res.Next() {
		err = res.Scan(&vm.ID, &vm.Params.Backend, &vm.Params.Cores, &vm.Params.Memory, &vm.Params.ImageID, &vm.Params.NetBridgeOn)
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

			vm.Drives = append(vm.Drives, d)
		}
	} else {
		return vm, ErrVmNotFound
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
