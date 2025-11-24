package tests

import (
	"crypto/sha256"
	"testing"
	"../src/snapshot"
	"../src/catalog"
)

func TestComputeMerkleRoot(t *testing.T) {
	hashes := []string{
		"hash1",
		"hash2",
		"hash3",
		"hash4",
	}

	root := snapshot.ComputeMerkleRoot(hashes)
	if root == "" {
		t.Error("Merkle root should not be empty")
	}
	t.Logf("Merkle root: %s", root)
}

func TestORSet(t *testing.T) {
	set := catalog.NewORSet()

	set.Add("key1", "value1")
	set.Add("key2", "value2")

	result := set.Query()
	if len(result) != 2 {
		t.Errorf("Expected 2 elements, got %d", len(result))
	}

	set.Remove("key1")
	result = set.Query()
	if len(result) != 1 {
		t.Errorf("Expected 1 element after remove, got %d", len(result))
	}

	val, exists := set.Get("key2")
	if !exists || val != "value2" {
		t.Error("Failed to get existing element")
	}
}

func TestChunking(t *testing.T) {
	data := make([]byte, 100*1024*1024) // 100MB
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Simulate chunking
	const chunkSize = 64 * 1024 * 1024
	var chunks [][]byte
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := make([]byte, end-i)
		copy(chunk, data[i:end])
		chunks = append(chunks, chunk)
	}

	if len(chunks) != 2 {
		t.Errorf("Expected 2 chunks, got %d", len(chunks))
	}

	// Verify hashes
	var hashes []string
	for _, chunk := range chunks {
		hash := sha256.Sum256(chunk)
		hashes = append(hashes, fmt.Sprintf("%x", hash))
	}

	root := snapshot.ComputeMerkleRoot(hashes)
	if root == "" {
		t.Error("Merkle root should not be empty")
	}
	t.Logf("Merkle root for chunks: %s", root)
}
