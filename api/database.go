package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/shared"
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

func DBStoreImage(i Image) error {
	err := Database.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(ImagesBucket)
		if err != nil {
			return err
		}

		data, err := json.Marshal(i)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(i.Info().Name), data)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func DBListImages() ([]Image, error) {
	var imgs []Image = make([]Image, 0)

	Database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(ImagesBucket)
		if b == nil {
			return nil
		}

		b.ForEach(func(_, v []byte) error {
			if strings.Contains(string(v), "\"Type\":\"qemu\"") {
				i := new(QemuImage)

				err := json.Unmarshal(v, i)
				if err != nil {
					return err
				}

				imgs = append(imgs, i)
			} else if strings.Contains(string(v), "\"Type\":\"lxc\"") {
				i := new(LxcImage)

				err := json.Unmarshal(v, i)
				if err != nil {
					return err
				}

				imgs = append(imgs, i)
			}

			return nil
		})

		return nil
	})

	return imgs, nil
}

func DBGetImage(name string) (Image, error) {
	var img Image

	err := Database.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(ImagesBucket)
		if bucket == nil {
			return errors.NotFound
		}

		data := bucket.Get([]byte(name))
		if data == nil {
			return errors.NotFound
		}

		if strings.Contains(string(data), "\"Type\":\"qemu\"") {
			i := new(QemuImage)

			err := json.Unmarshal(data, i)
			if err != nil {
				return err
			}

			img = i
		} else if strings.Contains(string(data), "\"Type\":\"lxc\"") {
			i := new(LxcImage)

			err := json.Unmarshal(data, i)
			if err != nil {
				return err
			}

			img = i
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

		data, err := json.Marshal(m)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(m.Info().Name), data)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func DBListMachines() (Machines, error) {
	var ms []Machine = make([]Machine, 0)

	Database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MachinesBucket)
		if b == nil {
			return nil
		}

		b.ForEach(func(_, v []byte) error {
			if strings.Contains(string(v), "\"Type\":\"qemu\"") {
				m := new(QemuMachine)

				err := json.Unmarshal(v, m)
				if err != nil {
					return err
				}

				ms = append(ms, m)
			} else if strings.Contains(string(v), "\"Type\":\"lxc\"") {
				m := new(LxcMachine)

				err := json.Unmarshal(v, m)
				if err != nil {
					return err
				}

				ms = append(ms, m)
			}

			return nil
		})

		return nil
	})

	return Machines(ms), nil
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

		if strings.Contains(string(data), "\"Type\":\"qemu\"") {
			mm := new(QemuMachine)

			err := json.Unmarshal(data, mm)
			if err != nil {
				return err
			}

			m = mm
		} else if strings.Contains(string(data), "\"Type\":\"lxc\"") {
			mm := new(LxcMachine)

			err := json.Unmarshal(data, mm)
			if err != nil {
				return err
			}

			m = mm
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
