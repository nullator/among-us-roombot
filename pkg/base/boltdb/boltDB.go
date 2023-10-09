package boltdb

import (
	"sync"

	"github.com/boltdb/bolt"
)

type base struct {
	db    *bolt.DB
	mutex sync.Mutex
}

func NewBase(db *bolt.DB) *base {
	return &base{db: db}
}

func (db *base) Save(key string, value string, bucket string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	err := db.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(key), []byte(value))
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *base) Get(key string, bucket string) (string, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	var value string
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b != nil {
			data := b.Get([]byte(key))
			value = string(data)
		} else {
			value = ""
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, err
}
