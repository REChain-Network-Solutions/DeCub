package storage

import "context"

// Store defines the interface for the storage layer
type Store interface {
	// Get retrieves a value by key
	Get(ctx context.Context, key []byte) ([]byte, error)
	
	// Set sets a value for a key
	Set(ctx context.Context, key, value []byte) error
	
	// Delete removes a key
	Delete(ctx context.Context, key []byte) error
	
	// Has checks if a key exists
	Has(ctx context.Context, key []byte) (bool, error)
	
	// Iterate iterates over all keys with the given prefix
	Iterate(ctx context.Context, prefix []byte, fn func(key, value []byte) error) error
	
	// Close closes the store and releases resources
	Close() error
}
