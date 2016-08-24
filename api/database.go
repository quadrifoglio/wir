package api

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/shared"
)

var (
	Database *bolt.DB

	ImagesBucket   = []byte("images")
	NetworksBucket = []byte("networks")
	MachinesBucket = []byte("machines")
)

func DBOpen(file string) error {
	db, err := bolt.Open(shared.APIConfig.DatabaseFile, 0600, nil)
	if err != nil {
		return err
	}

	Database = db
	return nil
}

func DBStoreImage(i shared.Image) error {
	err := Database.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(ImagesBucket)
		if err != nil {
			return err
		}

		buf := bytes.NewBuffer(nil)
		enc := gob.NewEncoder(buf)

		err = enc.Encode(i)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(i.Name), buf.Bytes())
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func DBListImages() ([]shared.Image, error) {
	imgs := make([]shared.Image, 0)

	Database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(ImagesBucket)
		if b == nil {
			return nil
		}

		b.ForEach(func(_, v []byte) error {
			var i shared.Image

			buf := bytes.NewBuffer(v)
			dec := gob.NewDecoder(buf)

			err := dec.Decode(&i)
			if err != nil {
				return err
			}

			imgs = append(imgs, i)
			return nil
		})

		return nil
	})

	return imgs, nil
}

func DBGetImage(name string) (shared.Image, error) {
	var img shared.Image

	err := Database.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(ImagesBucket)
		if bucket == nil {
			return errors.NotFound
		}

		data := bucket.Get([]byte(name))
		if data == nil {
			return errors.NotFound
		}

		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)

		err := dec.Decode(&img)
		if err != nil {
			return err
		}

		return nil
	})

	return img, err
}

func DBDeleteImage(name string) error {
	err := Database.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(ImagesBucket)
		if err != nil {
			return err
		}

		err = bucket.Delete([]byte(name))
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func DBStoreNetwork(i shared.Network) error {
	err := Database.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(NetworksBucket)
		if err != nil {
			return err
		}

		buf := bytes.NewBuffer(nil)
		enc := gob.NewEncoder(buf)

		err = enc.Encode(i)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(i.Name), buf.Bytes())
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func DBListNetworks() ([]shared.Network, error) {
	netws := make([]shared.Network, 0)

	Database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(NetworksBucket)
		if b == nil {
			return nil
		}

		b.ForEach(func(_, v []byte) error {
			var i shared.Network

			buf := bytes.NewBuffer(v)
			dec := gob.NewDecoder(buf)

			err := dec.Decode(&i)
			if err != nil {
				return err
			}

			netws = append(netws, i)
			return nil
		})

		return nil
	})

	return netws, nil
}

func DBGetNetwork(name string) (shared.Network, error) {
	var netw shared.Network

	err := Database.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(NetworksBucket)
		if bucket == nil {
			return errors.NotFound
		}

		data := bucket.Get([]byte(name))
		if data == nil {
			return errors.NotFound
		}

		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)

		err := dec.Decode(&netw)
		if err != nil {
			return err
		}

		return nil
	})

	return netw, err
}

func DBDeleteNetwork(name string) error {
	err := Database.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(NetworksBucket)
		if err != nil {
			return err
		}

		err = bucket.Delete([]byte(name))
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func DBMachineNameFree(name string) bool {
	err := Database.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(MachinesBucket)

		if bucket == nil {
			return fmt.Errorf("Missing database bucket: %s", MachinesBucket)
		}

		data := bucket.Get([]byte(name))
		if data == nil {
			return errors.NotFound
		}

		return nil
	})

	if err != nil {
		return true
	}

	return false
}

func DBStoreMachine(m Machine) error {
	err := Database.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(MachinesBucket)
		if err != nil {
			return err
		}

		if m.Info().Index == 0 {
			n, err := bucket.NextSequence()
			if err != nil {
				return err
			}

			m.Info().Index = n
		}

		buf := bytes.NewBuffer(nil)
		enc := gob.NewEncoder(buf)

		err = enc.Encode(m)
		if err != nil {
			return err
		}

		data := []byte(fmt.Sprintf("%s:", m.Type()))
		data = append(data, buf.Bytes()...)

		err = bucket.Put([]byte(m.Info().Name), data)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func DBMachineType(data []byte) string {
	var t string

	i := strings.Index(string(data), ":")
	if i > 0 {
		t = string(data[:i])
	}

	return t
}

func DBListMachines() (Machines, error) {
	var ms []Machine = make([]Machine, 0)

	Database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MachinesBucket)
		if b == nil {
			return nil
		}

		b.ForEach(func(_, v []byte) error {
			t := DBMachineType(v)
			if len(t) == 0 {
				return fmt.Errorf("database: no machine type stored")
			}

			if t == "qemu" {
				m := new(QemuMachine)
				dec := gob.NewDecoder(bytes.NewBuffer(v[len(t)+1:]))

				err := dec.Decode(m)
				if err != nil {
					return err
				}

				ms = append(ms, m)
			} else if t == "lxc" {
				m := new(LxcMachine)
				dec := gob.NewDecoder(bytes.NewBuffer(v[len(t)+1:]))

				err := dec.Decode(m)
				if err != nil {
					return err
				}

				ms = append(ms, m)
			} else {
				return fmt.Errorf("database: invalid machine type: %s", t)
			}

			return nil
		})

		return nil
	})

	return Machines(ms), nil
}

func DBGetMachineByMAC(mac string) (Machine, error) {
	ms, err := DBListMachines()
	if err != nil {
		return nil, err
	}

	for _, m := range ms {
		for _, i := range m.ListInterfaces() {
			if i.MAC == mac {
				return m, nil
			}
		}
	}

	return nil, errors.NotFound
}

func DBGetMachine(name string) (Machine, error) {
	var m Machine

	err := Database.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(MachinesBucket)

		if bucket == nil {
			return fmt.Errorf("Missing database bucket: %s", MachinesBucket)
		}

		data := bucket.Get([]byte(name))
		if data == nil {
			return errors.NotFound
		}

		t := DBMachineType(data)
		if len(t) == 0 {
			return fmt.Errorf("database: no machine type stored")
		}

		if t == "qemu" {
			mm := new(QemuMachine)
			dec := gob.NewDecoder(bytes.NewBuffer(data[len(t)+1:]))

			err := dec.Decode(mm)
			if err != nil {
				return err
			}

			m = mm
		} else if t == "lxc" {
			mm := new(LxcMachine)
			dec := gob.NewDecoder(bytes.NewBuffer(data[len(t)+1:]))

			err := dec.Decode(mm)
			if err != nil {
				return err
			}

			m = mm
		} else {
			return fmt.Errorf("database: invalid machine type: %s", t)
		}

		return nil
	})

	return m, err
}

func DBDeleteMachine(name string) error {
	err := Database.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(MachinesBucket)

		if bucket == nil {
			return fmt.Errorf("Missing database bucket: %s", MachinesBucket)
		}

		err := bucket.Delete([]byte(name))
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
