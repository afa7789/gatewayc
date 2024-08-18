// boltdb_wrapper.go
package boltdb

import (
	"log"

	"go.etcd.io/bbolt"
)

// BoltDBWrapper is a struct that wraps the BoltDB instance
type BoltDBWrapper struct {
	db *bbolt.DB
}

// NewBoltDB initializes and opens a new BoltDB instance
func NewBoltDB(path string) *BoltDBWrapper {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		log.Fatalf("Failed to open BoltDB: %v", err)
	}

	return &BoltDBWrapper{db: db}
}

// Close closes the BoltDB database
func (b *BoltDBWrapper) Close() {
	if err := b.db.Close(); err != nil {
		log.Fatalf("Failed to close BoltDB: %v", err)
	}
}

// WriteToDB writes a key-value pair to the BoltDB instance
func (b *BoltDBWrapper) WriteToDB(bucketName, key, value string) error {
	err := b.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			var err error
			bucket, err = tx.CreateBucket([]byte(bucketName))
			if err != nil {
				return err
			}
		}

		return bucket.Put([]byte(key), []byte(value))
	})

	if err != nil {
		return err
	}

	return nil
}

// ReadFromDB reads the value for a given key from the BoltDB instance
func (b *BoltDBWrapper) ReadFromDB(bucketName, key string) (error, string) {
	var value string
	err := b.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return nil
		}

		v := bucket.Get([]byte(key))
		if v == nil {
			return nil
		}

		value = string(v)
		return nil
	})

	if err != nil {
		return err, ""
	}

	return nil, value
}
