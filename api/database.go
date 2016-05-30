package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
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
	db, err := bolt.Open(Conf.DatabaseFile, 0600, nil)
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
			return fmt.Errorf("Missing database bucket: %s", ImagesBucket)
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
		b := tx.Bucket(MachinesBucket)
		if b == nil {
			return nil
		}

		return b.ForEach(func(_, v []byte) error {
			var m machine.Machine

			err := json.Unmarshal(v, &m)
			if err != nil {
				return fmt.Errorf("JSON: %s", err)
			}

			if m.Name == name {
				return fmt.Errorf("Found")
			}

			return nil
		})
	})

	if err != nil {
		if err.Error() != "Found" {
			log.Println("Database: ", err)
		}

		return false
	}

	return true
}

func DBStoreMachine(m *machine.Machine) error {
	err := Database.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(MachinesBucket)
		if err != nil {
			return err
		}

		data, err := json.Marshal(m)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(m.ID), data)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func DBListMachines() ([]machine.Machine, error) {
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

	return ms, nil
}

func DBGetMachine(idf string) (machine.Machine, error) {
	var m machine.Machine

	err := Database.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(MachinesBucket)

		if bucket == nil {
			return fmt.Errorf("Missing database bucket: %s", MachinesBucket)
		}

		_, err := hex.DecodeString(idf)
		if err == nil {
			data := bucket.Get([]byte(idf))
			if data == nil {
				return errors.NotFound
			}

			err := json.Unmarshal(data, &m)
			if err != nil {
				return err
			}
		} else {
			bucket.ForEach(func(_, v []byte) error {
				var mm machine.Machine

				err := json.Unmarshal(v, &mm)
				if err != nil {
					return fmt.Errorf("JSON: %s", err)
				}

				if mm.Name == idf {
					m = mm
				}

				return nil
			})
		}

		return nil
	})

	return m, err
}

func DBDeleteMachine(id string) error {
	err := Database.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(MachinesBucket)

		if bucket == nil {
			return fmt.Errorf("Missing database bucket: %s", MachinesBucket)
		}

		err := bucket.Delete([]byte(id))
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
