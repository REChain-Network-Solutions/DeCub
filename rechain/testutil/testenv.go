package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/rechain/rechain/internal/storage"
	"github.com/rechain/rechain/pkg/config"
)

// TestEnvironment manages the test environment for integration tests
type TestEnvironment struct {
	T       *testing.T
	TempDir string
	Config  *config.Config
	Store   storage.Store
}

// NewTestEnvironment creates a new test environment
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "rechain-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create a test config
	cfg := config.DefaultConfig()
	cfg.Node.DataDir = tempDir
	cfg.Storage.Path = filepath.Join(tempDir, "data")

	// Create a BadgerDB store
	db, err := storage.NewBadgerStore(cfg.Storage.Path, cfg.Storage.CacheSize, cfg.Storage.Sync)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to create BadgerDB store: %v", err)
	}

	return &TestEnvironment{
		T:       t,
		TempDir: tempDir,
		Config:  cfg,
		Store:   db,
	}
}

// Close cleans up the test environment
func (env *TestEnvironment) Close() {
	env.T.Helper()

	if env.Store != nil {
		if err := env.Store.Close(); err != nil {
			env.T.Logf("error closing store: %v", err)
		}
	}

	if env.TempDir != "" {
		if err := os.RemoveAll(env.TempDir); err != nil {
			env.T.Logf("error removing temp dir: %v", err)
		}
	}
}

// WithMerkleStore creates a MerkleStore for testing
func (env *TestEnvironment) WithMerkleStore() *storage.MerkleStore {
	env.T.Helper()

	ms, err := storage.NewMerkleStore(env.Store)
	if err != nil {
		env.T.Fatalf("failed to create MerkleStore: %v", err)
	}

	return ms
}

// MustSet sets a key-value pair in the store, failing the test on error
func (env *TestEnvironment) MustSet(ctx context.Context, key, value []byte) {
	env.T.Helper()

	if err := env.Store.Set(ctx, key, value); err != nil {
		env.T.Fatalf("failed to set key %q: %v", key, err)
	}
}

// MustGet gets a value from the store, failing the test on error
func (env *TestEnvironment) MustGet(ctx context.Context, key []byte) []byte {
	env.T.Helper()

	value, err := env.Store.Get(ctx, key)
	if err != nil {
		env.T.Fatalf("failed to get key %q: %v", key, err)
	}

	return value
}

// MustNotExist verifies that a key does not exist in the store
func (env *TestEnvironment) MustNotExist(ctx context.Context, key []byte) {
	env.T.Helper()

	has, err := env.Store.Has(ctx, key)
	if err != nil {
		env.T.Fatalf("failed to check key %q: %v", key, err)
	}

	if has {
		env.T.Fatalf("key %q exists but should not", key)
	}
}

// MustCommit commits the current state and returns the root hash
func (env *TestEnvironment) MustCommit(ms *storage.MerkleStore) []byte {
	env.T.Helper()

	root, err := ms.Commit()
	if err != nil {
		env.T.Fatalf("failed to commit: %v", err)
	}

	return root
}

// MustVerifyProof verifies a Merkle proof
func (env *TestEnvironment) MustVerifyProof(ms *storage.MerkleStore, key, value, root []byte, proof [][]byte) bool {
	env.T.Helper()

	return ms.VerifyProof(key, value, proof)
}

// MustLoadState loads a previously committed state
func (env *TestEnvironment) MustLoadState(ms *storage.MerkleStore, height uint64) []byte {
	env.T.Helper()

	root, err := ms.LoadState(height)
	if err != nil {
		env.T.Fatalf("failed to load state at height %d: %v", height, err)
	}

	return root
}

// MustCreateCRDT creates a new CRDT instance for testing
func (env *TestEnvironment) MustCreateCRDT(t crdt.CRDTType, nodeID string) crdt.CRDT {
	env.T.Helper()

	c, err := crdt.New(t, nodeID)
	if err != nil {
		env.T.Fatalf("failed to create CRDT %q: %v", t, err)
	}

	return c
}
