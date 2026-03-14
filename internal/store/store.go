package store

import "errors"

// ErrNotFound is returned when a key does not exist.
var ErrNotFound = errors.New("key not found")

// Entry holds an encrypted value and its last-write timestamp.
type Entry struct {
	Key       string
	Value     []byte
	UpdatedAt int64
}

// Store is a simple key-value store with byte slice values.
type Store interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
	Delete(key string) error
	List() ([]string, error)
	Close() error

	// SetRaw writes an encrypted value with an explicit timestamp (for sync pull).
	SetRaw(key string, value []byte, ts int64) error
	// GetUpdatedAt returns the unix timestamp for when a key was last written.
	GetUpdatedAt(key string) (int64, error)
	// ListEntries returns all keys with their encrypted values and timestamps.
	ListEntries() ([]Entry, error)
}
