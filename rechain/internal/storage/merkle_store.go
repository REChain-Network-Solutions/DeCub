package storage

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rechain/rechain/pkg/merkle"
)

// MerkleStore is a storage implementation that maintains a Merkle tree for state verification
type MerkleStore struct {
	base   Store       // Underlying key-value store
	tree   *merkle.Tree // Merkle tree for state verification
	mu     sync.RWMutex
	height uint64 // Current height of the state
}

// NewMerkleStore creates a new Merkle-backed store
func NewMerkleStore(base Store) (*MerkleStore, error) {
	// Initialize with an empty tree
	tree, err := merkle.NewTree(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Merkle tree: %w", err)
	}

	ms := &MerkleStore{
		base:   base,
		tree:   tree,
		height: 0,
	}

	// Rebuild the Merkle tree from the underlying store
	if err := ms.rebuildTree(); err != nil {
		return nil, fmt.Errorf("failed to rebuild Merkle tree: %w", err)
	}

	return ms, nil
}

// rebuildTree rebuilds the Merkle tree from the underlying store
func (ms *MerkleStore) rebuildTree() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Get all key-value pairs from the base store
	var data = make(map[string][]byte)
	err := ms.base.Iterate(context.Background(), nil, func(key, value []byte) error {
		// Skip internal keys
		if isInternalKey(key) {
			return nil
		}

		// Add to our data map
		data[string(key)] = value
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to iterate over base store: %w", err)
	}

	// Create a new Merkle tree with the data
	tree, err := merkle.NewTree(data)
	if err != nil {
		return fmt.Errorf("failed to create Merkle tree: %w", err)
	}

	ms.tree = tree
	return nil
}

// Get retrieves a value by key
func (ms *MerkleStore) Get(ctx context.Context, key []byte) ([]byte, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return ms.base.Get(ctx, key)
}

// Set sets a value for a key and updates the Merkle tree
func (ms *MerkleStore) Set(ctx context.Context, key, value []byte) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Update the base store
	if err := ms.base.Set(ctx, key, value); err != nil {
		return fmt.Errorf("failed to set key in base store: %w", err)
	}

	// Update the Merkle tree
	ms.tree.Update(string(key), value)

	return nil
}

// Delete removes a key and updates the Merkle tree
func (ms *MerkleStore) Delete(ctx context.Context, key []byte) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Delete from the base store
	if err := ms.base.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete key from base store: %w", err)
	}

	// Update the Merkle tree (set to empty value to mark as deleted)
	ms.tree.Update(string(key), nil)

	return nil
}

// Has checks if a key exists
func (ms *MerkleStore) Has(ctx context.Context, key []byte) (bool, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return ms.base.Has(ctx, key)
}

// Iterate iterates over all keys with the given prefix
func (ms *MerkleStore) Iterate(ctx context.Context, prefix []byte, fn func(key, value []byte) error) error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return ms.base.Iterate(ctx, prefix, fn)
}

// Close closes the store
func (ms *MerkleStore) Close() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return ms.base.Close()
}

// RootHash returns the current Merkle root hash
func (ms *MerkleStore) RootHash() []byte {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return []byte(ms.tree.RootHash())
}

// GetProof returns a Merkle proof for the given key
func (ms *MerkleStore) GetProof(key []byte) ([][]byte, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return ms.tree.GetProof(key)
}

// VerifyProof verifies a Merkle proof
func (ms *MerkleStore) VerifyProof(key, value []byte, proof [][]byte) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return merkle.VerifyProof([]byte(ms.tree.RootHash()), key, value, proof)
}

// Commit commits the current state and returns the new root hash
func (ms *MerkleStore) Commit() ([]byte, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Get the current root hash
	rootHash := ms.tree.RootHash()

	// Store the root hash in the base store
	rootKey := ms.rootKey(ms.height)
	if err := ms.base.Set(context.Background(), rootKey, []byte(rootHash)); err != nil {
		return nil, fmt.Errorf("failed to store root hash: %w", err)
	}

	// Increment the height
	ms.height++

	return []byte(rootHash), nil
}

// LoadState loads a previously committed state
func (ms *MerkleStore) LoadState(height uint64) ([]byte, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	rootKey := ms.rootKey(height)
	rootHash, err := ms.base.Get(context.Background(), rootKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load root hash for height %d: %w", height, err)
	}

	return rootHash, nil
}

// rootKey returns the storage key for a root hash at the given height
func (ms *MerkleStore) rootKey(height uint64) []byte {
	return []byte(fmt.Sprintf("_root/%d", height))
}

// isInternalKey checks if a key is used internally by the MerkleStore
func isInternalKey(key []byte) bool {
	return len(key) >= 6 && string(key[:6]) == "_root/"
}
