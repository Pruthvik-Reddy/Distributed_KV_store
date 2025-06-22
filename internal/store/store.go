package store

import (
	"fmt"
	"os"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// Store is the interface for our key-value store.
// It will be the "state machine" for our Raft implementation.
type Store interface {
	Get(key string) (string, error)
	Put(key, value string) error
	Delete(key string) error
	Close() error
}

// levelDBStore is a concrete implementation of the Store interface using LevelDB.
type levelDBStore struct {
	db *leveldb.DB
	// A mutex is used to protect concurrent access to the database.
	// While our first task is single-threaded, this prepares us for the future
	// when multiple goroutines (e.g., from gRPC requests) will interact with the store.
	mu sync.RWMutex
}

// NewLevelDBStore creates and returns a new instance of a LevelDB-backed store.
// dbPath is the directory where the database files will be stored.
func NewLevelDBStore(dbPath string) (Store, error) {
	// Ensure the directory exists.
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db path %s: %w", dbPath, err)
	}

	// Open the LevelDB database. The options are minimal for now.
	db, err := leveldb.OpenFile(dbPath, &opt.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to open leveldb: %w", err)
	}

	return &levelDBStore{
		db: db,
	}, nil
}

// Get retrieves the value for a given key.
// It returns an error if the key is not found or if there's a db issue.
func (s *levelDBStore) Get(key string) (string, error) {
	s.mu.RLock() // Acquire a read lock
	defer s.mu.RUnlock()

	value, err := s.db.Get([]byte(key), nil)
	if err != nil {
		// Differentiate between "not found" and other errors.
		if err == leveldb.ErrNotFound {
			return "", fmt.Errorf("key '%s' not found", key)
		}
		return "", fmt.Errorf("failed to get key '%s': %w", key, err)
	}

	return string(value), nil
}

// Put sets the value for a given key.
func (s *levelDBStore) Put(key, value string) error {
	s.mu.Lock() // Acquire a write lock
	defer s.mu.Unlock()

	err := s.db.Put([]byte(key), []byte(value), nil)
	if err != nil {
		return fmt.Errorf("failed to put key '%s': %w", key, err)
	}
	return nil
}

// Delete removes a key-value pair from the store.
func (s *levelDBStore) Delete(key string) error {
	s.mu.Lock() // Acquire a write lock
	defer s.mu.Unlock()

	err := s.db.Delete([]byte(key), nil)
	if err != nil {
		return fmt.Errorf("failed to delete key '%s': %w", key, err)
	}
	return nil
}

// Close closes the database connection.
func (s *levelDBStore) Close() error {
	return s.db.Close()
}