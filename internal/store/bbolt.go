package store

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

var bucketName = []byte("secrets")

// BoltStore implements Store using bbolt.
type BoltStore struct {
	db *bolt.DB
}

// Open opens (or creates) a BoltDB at the given path with 0600 permissions.
func Open(path string) (*BoltStore, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	// Ensure bucket exists.
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create bucket: %w", err)
	}
	return &BoltStore{db: db}, nil
}

func (s *BoltStore) Get(key string) ([]byte, error) {
	var val []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		v := b.Get([]byte(key))
		if v == nil {
			return ErrNotFound
		}
		val = make([]byte, len(v))
		copy(val, v)
		return nil
	})
	return val, err
}

func (s *BoltStore) Set(key string, value []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Put([]byte(key), value)
	})
}

func (s *BoltStore) Delete(key string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b.Get([]byte(key)) == nil {
			return ErrNotFound
		}
		return b.Delete([]byte(key))
	})
}

func (s *BoltStore) List() ([]string, error) {
	var keys []string
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.ForEach(func(k, _ []byte) error {
			keys = append(keys, string(k))
			return nil
		})
	})
	return keys, err
}

func (s *BoltStore) Close() error {
	return s.db.Close()
}
