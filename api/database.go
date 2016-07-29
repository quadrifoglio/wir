package api

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/quadrifoglio/wir/shared"
	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/image"
	"github.com/quadrifoglio/wir/machine"
)

var (
	Database *bolt.DB

	ImagesBucket   = []byte("images")
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

func DBStoreImage(i *image.Image) error {
	err := Database.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(ImagesBucket)
		if err != nil {
			return err
		}

		data, err := json.Marshal(i)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(i.Name), data)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func DBListImages() ([]image.Image, error) {
	var imgs []image.Image = make([]image.Image, 0)

	Database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(ImagesBucket)
		if b == nil {
			return nil
		}

		b.ForEach(func(_, v []byte) error {
			var i image.Image

			err := json.Unmarshal(v, &i)
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

func DBGetImage(name string) (image.Image, error) {
	var img image.Image

	err := Database.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(ImagesBucket)
		if bucket == nil {
			return errors.NotFound
		}

		data := bucket.Get([]byte(name))
		if data == nil {
			return errors.NotFound
		}

		err := json.Unmarshal(data, &img)
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

// TODO: Opzimize / find a better way
func DBFreeMachineIndex() (uint64, error) {
	var i uint64

	err := Database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MachinesBucket)
		if b == nil {
			return nil
		}

		b.ForEach(func(_, v []byte) error {
			var m machine.Machine

			err := json.Unmarshal(v, &m)
			if err != nil {
				return err
			}

			if m.Index > i {
				i = m.Index
			}

			return nil
		})

		return nil
	})

	if err != nil {
		return 0, err
	}

	return i + 1, nil
}

func DBStoreMachine(m *machine.Machine) error {
	err := Database.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(MachinesBucket)
		if err != nil {
			return err
		}

		if m.Index == 0 {
			n, err := bucket.NextSequence()
			if err != nil {
				return err
			}

			m.Index = n
		}

		data, err := json.Marshal(m)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(m.Name), data)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func DBListMachines() (machine.Machines, error) {
	var ms []machine.Machine = make([]machine.Machine, 0)

	Database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MachinesBucket)
		if b == nil {
			return nil
		}

		b.ForEach(func(_, v []byte) error {
			var m machine.Machine

			err := json.Unmarshal(v, &m)
			if err != nil {
				return err
			}

			ms = append(ms, m)
			return nil
		})

		return nil
	})

	return machine.Machines(ms), nil
}

func DBGetMachine(name string) (machine.Machine, error) {
	var m machine.Machine

	err := Database.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(MachinesBucket)

		if bucket == nil {
			return fmt.Errorf("Missing database bucket: %s", MachinesBucket)
		}

		data := bucket.Get([]byte(name))
		if data == nil {
			return errors.NotFound
		}

		err := json.Unmarshal(data, &m)
		if err != nil {
			return err
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
