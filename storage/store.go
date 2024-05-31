package storage

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

type KVStore struct {
	db *bolt.DB
}

var (
	ErrNotFound = errors.New("skv: key not found")
	ErrBadValue = errors.New("skv: bad value")
	bucketName  = []byte("byod")
	Store       *KVStore
)

func init() {
	var err error
	os.Remove("byod.db")
	Store, err = Open()
	if err != nil {
		log.Println("Unable to open byod.db")
	}
}

func Open() (*KVStore, error) {
	opts := &bolt.Options{
		Timeout: 50 * time.Millisecond,
	}
	if db, err := bolt.Open("byod.db", 0640, opts); err != nil {
		return nil, err
	} else {
		err := db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists(bucketName)
			return err
		})
		if err != nil {
			return nil, err
		} else {
			return &KVStore{db: db}, nil
		}
	}
}

func (kvs *KVStore) Put(key string, value interface{}) error {
	if value == nil {
		return ErrBadValue
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(value); err != nil {
		return err
	}
	return kvs.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketName).Put([]byte(key), buf.Bytes())
	})
}

func (kvs *KVStore) Get(key string, value interface{}) error {
	return kvs.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketName).Cursor()
		if k, v := c.Seek([]byte(key)); k == nil || string(k) != key {
			return ErrNotFound
		} else if value == nil {
			return nil
		} else {
			d := gob.NewDecoder(bytes.NewReader(v))
			return d.Decode(value)
		}
	})
}

func (kvs *KVStore) Delete(key string) error {
	return kvs.db.Update(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketName).Cursor()
		if k, _ := c.Seek([]byte(key)); k == nil || string(k) != key {
			return ErrNotFound
		} else {
			return c.Delete()
		}
	})
}

func (kvs *KVStore) Close() error {
	return kvs.db.Close()
}
