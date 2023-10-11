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

func (db *base) SaveBytes(key string, value []byte, bucket string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	err := db.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(key), value)
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

func (db *base) GetBytes(key string, bucket string) ([]byte, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	var data []byte
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b != nil {
			data = b.Get([]byte(key))
			return nil
		} else {
			data = nil
			return nil
		}
	})
	if err != nil {
		return nil, err
	}
	return data, err
}

func (db *base) Delete(key string, bucket string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	err := db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b != nil {
			return b.Delete([]byte(key))
		} else {
			return nil
		}
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *base) GetAll(bucket string) ([][]byte, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	var data [][]byte
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b != nil {
			b.ForEach(func(k, v []byte) error {
				data = append(data, v)
				return nil
			})
			return nil
		} else {
			data = nil
			return nil
		}
	})
	if err != nil {
		return nil, err
	}
	return data, err
}
