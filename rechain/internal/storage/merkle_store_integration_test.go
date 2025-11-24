package storage_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/rechain/rechain/internal/storage"
	"github.com/rechain/rechain/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMerkleStore_Integration(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Close()

	// Create a new MerkleStore
	ms := env.WithMerkleStore()

	// Test data
	ctx := context.Background()
	key1 := []byte("test-key-1")
	value1 := []byte("test-value-1")
	key2 := []byte("test-key-2")
	value2 := []byte("test-value-2")

	t.Run("Set and Get", func(t *testing.T) {
		// Set a key-value pair
		err := ms.Set(ctx, key1, value1)
		require.NoError(t, err)

		// Get the value back
		gotValue, err := ms.Get(ctx, key1)
		require.NoError(t, err)
		assert.Equal(t, value1, gotValue)

		// Verify the value exists in the underlying store
		gotValue, err = env.Store.Get(ctx, key1)
		require.NoError(t, err)
		assert.Equal(t, value1, gotValue)
	})

	t.Run("Merkle Proof", func(t *testing.T) {
		// Add more data
		err := ms.Set(ctx, key2, value2)
		require.NoError(t, err)

		// Get a proof for key1
		proof, err := ms.GetProof(key1)
		require.NoError(t, err)
		require.NotNil(t, proof)

		// Verify the proof
		root := ms.RootHash()
		isValid := storage.VerifyProof(root, key1, value1, proof)
		assert.True(t, isValid, "Merkle proof verification failed")

		// Verify with wrong value should fail
		isValid = storage.VerifyProof(root, key1, []byte("wrong-value"), proof)
		assert.False(t, isValid, "Merkle proof verification should fail with wrong value")
	})

	t.Run("Commit and Load State", func(t *testing.T) {
		// Commit the current state
		root1, err := ms.Commit()
		require.NoError(t, err)
		require.NotNil(t, root1)

		// Modify the data
		newValue1 := []byte("new-test-value-1")
		err = ms.Set(ctx, key1, newValue1)
		require.NoError(t, err)

		// Commit the new state
		root2, err := ms.Commit()
		require.NoError(t, err)
		require.NotNil(t, root2)

		// Verify the roots are different
		assert.NotEqual(t, root1, root2, "Root hashes should be different after modification")

		// Load the first state
		loadedRoot1, err := ms.LoadState(0)
		require.NoError(t, err)
		assert.Equal(t, root1, loadedRoot1, "Loaded root hash does not match expected")

		// Load the second state
		loadedRoot2, err := ms.LoadState(1)
		require.NoError(t, err)
		assert.Equal(t, root2, loadedRoot2, "Loaded root hash does not match expected")

		// Verify the value in the first state
		gotValue, err := ms.Get(ctx, key1)
		require.NoError(t, err)
		assert.Equal(t, value1, gotValue, "Value in first state does not match expected")

		// Load the second state and verify the updated value
		_, err = ms.LoadState(1)
		require.NoError(t, err)

		gotValue, err = ms.Get(ctx, key1)
		require.NoError(t, err)
		assert.Equal(t, newValue1, gotValue, "Value in second state does not match expected")
	})

	t.Run("Concurrent Access", func(t *testing.T) {
		// This test verifies that the MerkleStore can handle concurrent access
		// by multiple goroutines.

		const numGoroutines = 10
		const numOperations = 100

		// Channel to collect errors from goroutines
		errCh := make(chan error, numGoroutines*numOperations)

		// Start multiple goroutines that perform operations concurrently
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				for j := 0; j < numOperations; j++ {
					key := []byte(fmt.Sprintf("concurrent-%d-%d", id, j))
					value := []byte(fmt.Sprintf("value-%d-%d", id, j))

					// Set a key-value pair
					if err := ms.Set(ctx, key, value); err != nil {
						errCh <- fmt.Errorf("failed to set %q: %w", key, err)
						return
					}

					// Get the value back
					gotValue, err := ms.Get(ctx, key)
					if err != nil {
						errCh <- fmt.Errorf("failed to get %q: %w", key, err)
						return
					}

					if string(gotValue) != string(value) {
						errCh <- fmt.Errorf("value mismatch for %q: got %q, want %q", 
							key, gotValue, value)
						return
					}
				}
				errCh <- nil
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			if err := <-errCh; err != nil {
				t.Fatal(err)
			}
		}

		// Verify all key-value pairs
		for i := 0; i < numGoroutines; i++ {
			for j := 0; j < numOperations; j++ {
				key := []byte(fmt.Sprintf("concurrent-%d-%d", i, j))
				expectedValue := []byte(fmt.Sprintf("value-%d-%d", i, j))

				gotValue, err := ms.Get(ctx, key)
				require.NoError(t, err)
				assert.Equal(t, expectedValue, gotValue)
			}
		}
	})
}
