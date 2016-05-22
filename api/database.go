package api

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/quadrifoglio/wir/errors"
	"github.com/quadrifoglio/wir/image"
)

var (
	Database *bolt.DB

	ImagesBucket   = []byte("images")
	MachinesBucket = []byte("images")
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
