# Storage Package

This package provides a unified storage interface with Merkle tree-based state verification.

## Overview

The storage package offers:
- A generic `Store` interface for key-value storage
- A `MerkleStore` implementation that maintains a Merkle tree for state verification
- Integration with various storage backends (BadgerDB, etc.)

## Store Interface

The `Store` interface defines the basic operations for a key-value store:

```go
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
    
    // Close closes the store
    Close() error
}
```

## MerkleStore

The `MerkleStore` wraps any `Store` implementation and maintains a Merkle tree for efficient state verification.

### Features
- **State Verification**: Generate and verify Merkle proofs for key-value pairs
- **State Snapshots**: Take snapshots of the state at specific points in time
- **Efficient Updates**: Only update the parts of the Merkle tree that change
- **Concurrency Safe**: Supports concurrent reads and writes

### Usage

```go
import (
    "context"
    "github.com/rechain/rechain/internal/storage"
)

// Create a base store (e.g., BadgerDB)
baseStore, err := storage.NewBadgerStore("./data", 0, true)
if err != nil {
    // handle error
}

defer baseStore.Close()

// Create a MerkleStore
merkleStore, err := storage.NewMerkleStore(baseStore)
if err != nil {
    // handle error
}

// Set a value
ctx := context.Background()
err = merkleStore.Set(ctx, []byte("key"), []byte("value"))

// Get a value
value, err := merkleStore.Get(ctx, []byte("key"))

// Generate a proof for a key
proof, err := merkleStore.GetProof([]byte("key"))

// Verify a proof
isValid := merkleStore.VerifyProof(
    []byte("key"),
    []byte("value"),
    proof,
)

// Commit the current state
rootHash, err := merkleStore.Commit()

// Load a previous state
prevRoot, err := merkleStore.LoadState(0) // 0 = first commit
```

## Storage Backends

### BadgerDB

BadgerDB is a fast key-value store that's perfect for embedded use cases.

```go
import "github.com/rechain/rechain/internal/storage"

// Open a BadgerDB store
store, err := storage.NewBadgerStore(
    "./data",  // data directory
    0,         // cache size (0 = default)
    true,      // sync writes
)
if err != nil {
    // handle error
}
defer store.Close()
```

## State Management

The `MerkleStore` maintains a history of state roots. Each call to `Commit()` creates a new state version.

```go
// Initial state
merkleStore.Set(ctx, []byte("a"), []byte("1"))
root1, _ := merkleStore.Commit()

// Update state
merkleStore.Set(ctx, []byte("b"), []byte("2"))
root2, _ := merkleStore.Commit()

// Load previous state
prevRoot, _ := merkleStore.LoadState(0)
value, _ := merkleStore.Get(ctx, []byte("a")) // "1"
value, _ = merkleStore.Get(ctx, []byte("b"))  // "" (not set in this version)

// Go back to latest state
merkleStore.LoadState(1)
value, _ = merkleStore.Get(ctx, []byte("b"))  // "2"
```

## Performance Considerations

- **Batch Operations**: For better performance, batch multiple operations together
- **Cache Size**: Adjust the cache size based on your working set
- **Sync Writes**: Disable sync writes (`sync=false`) for better performance (at the risk of data loss)

## Benchmarks

Run benchmarks with:

```bash
cd internal/storage
go test -bench=.
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
