package merkle

import (
	"crypto/rand"
	"testing"
)

func TestNewTree(t *testing.T) {
	// Test with empty data
	_, err := NewTree(map[string][]byte{})
	if err == nil {
		t.Error("expected error for empty data, got nil")
	}

	// Test with single entry
	tree, err := NewTree(map[string][]byte{"key1": []byte("value1")})
	if err != nil {
		t.Fatalf("failed to create tree: %v", err)
	}

	if tree == nil {
		t.Error("expected tree to be created, got nil")
	}

	if tree.Root == nil {
		t.Error("expected root node, got nil")
	}
}

func TestGet(t *testing.T) {
	data := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
	}

	tree, err := NewTree(data)
	if err != nil {
		t.Fatalf("failed to create tree: %v", err)
	}

	// Test existing key
	val, exists := tree.Get([]byte("key1"))
	if !exists {
		t.Error("expected key1 to exist")
	}
	if string(val) != "value1" {
		t.Errorf("expected value1, got %s", val)
	}

	// Test non-existent key
	_, exists = tree.Get([]byte("nonexistent"))
	if exists {
		t.Error("expected key to not exist")
	}
}

func TestProof(t *testing.T) {
	data := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
	}

	tree, err := NewTree(data)
	if err != nil {
		t.Fatalf("failed to create tree: %v", err)
	}

	// Test proof for existing key
	proof, err := tree.GetProof([]byte("key1"))
	if err != nil {
		t.Fatalf("failed to get proof: %v", err)
	}

	if len(proof) == 0 {
		t.Error("expected non-empty proof")
	}

	// Verify the proof
	key := []byte("key1")
	value := []byte("value1")
	if !VerifyProof(tree.Root.Hash, key, value, proof) {
		t.Error("proof verification failed")
	}

	// Test proof for non-existent key
	_, err = tree.GetProof([]byte("nonexistent"))
	if err == nil {
		t.Error("expected error for non-existent key")
	}
}

func TestVerifyProof(t *testing.T) {
	data := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
	}

	tree, err := NewTree(data)
	if err != nil {
		t.Fatalf("failed to create tree: %v", err)
	}

	// Get proof for key1
	proof, err := tree.GetProof([]byte("key1"))
	if err != nil {
		t.Fatalf("failed to get proof: %v", err)
	}

	// Test with correct key-value pair
	if !VerifyProof(tree.Root.Hash, []byte("key1"), []byte("value1"), proof) {
		t.Error("proof verification failed for correct key-value pair")
	}

	// Test with incorrect value
	if VerifyProof(tree.Root.Hash, []byte("key1"), []byte("wrongvalue"), proof) {
		t.Error("proof verification should fail for incorrect value")
	}

	// Test with incorrect key
	if VerifyProof(tree.Root.Hash, []byte("wrongkey"), []byte("value1"), proof) {
		t.Error("proof verification should fail for incorrect key")
	}
}

func TestLargeTree(t *testing.T) {
	// Test with a larger dataset
	data := make(map[string][]byte)
	for i := 0; i < 1000; i++ {
		key := make([]byte, 10)
		value := make([]byte, 100)
		rand.Read(key)
		rand.Read(value)
		data[string(key)] = value
	}

	tree, err := NewTree(data)
	if err != nil {
		t.Fatalf("failed to create large tree: %v", err)
	}

	// Test a few random keys
	for k, v := range data {
		// Test Get
		val, exists := tree.Get([]byte(k))
		if !exists || string(val) != string(v) {
			t.Errorf("failed to get value for key %s", k)
		}

		// Test proof
		proof, err := tree.GetProof([]byte(k))
		if err != nil {
			t.Fatalf("failed to get proof for key %s: %v", k, err)
		}

		if !VerifyProof(tree.Root.Hash, []byte(k), v, proof) {
			t.Errorf("proof verification failed for key %s", k)
		}

		// Only test a few keys to keep tests fast
		if len(data) > 10 {
			break
		}
	}
}
