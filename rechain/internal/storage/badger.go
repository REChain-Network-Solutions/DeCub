package storage

import (
	"context"
	"fmt"

	"github.com/dgraph-io/badger/v3"
)

// BadgerStore implements the Store interface using BadgerDB
type BadgerStore struct {
	db *badger.DB
}

// NewBadgerStore creates a new BadgerDB-backed store
func NewBadgerStore(path string) (*BadgerStore, error) {
	opts := badger.DefaultOptions(path)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db: %w", err)
	}

	return &BadgerStore{db: db}, nil
}

// Get retrieves a value by key
func (s *BadgerStore) Get(_ context.Context, key []byte) ([]byte, error) {
	var valCopy []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			valCopy = append([]byte{}, val...)
			return nil
		})
	})

	if err == badger.ErrKeyNotFound {
		return nil, nil
	}

	return valCopy, err
}

// Set sets a value for a key
func (s *BadgerStore) Set(_ context.Context, key, value []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

// Delete removes a key
func (s *BadgerStore) Delete(_ context.Context, key []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// Has checks if a key exists
func (s *BadgerStore) Has(_ context.Context, key []byte) (bool, error) {
	err := s.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		return err
	})

	if err == badger.ErrKeyNotFound {
		return false, nil
	}

	return err == nil, err
}

// Iterate iterates over all keys with the given prefix
func (s *BadgerStore) Iterate(_ context.Context, prefix []byte, fn func(key, value []byte) error) error {
	return s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			err := item.Value(func(val []byte) error {
				key := item.KeyCopy(nil)
				valCopy := append([]byte{}, val...)
				return fn(key, valCopy)
			})

			if err != nil {
				return err
			}
		}

		return nil
	})
}

// Close closes the store and releases resources
func (s *BadgerStore) Close() error {
	return s.db.Close()
}
