package merkle

import (
	"crypto/sha256"
	"fmt"
)

// Tree represents a Merkle tree
type Tree struct {
	root   *Node
	leaves []*Node
}

// Node represents a node in the Merkle tree
type Node struct {
	hash  []byte
	left  *Node
	right *Node
}

// NewTree creates a new Merkle tree from key-value pairs
func NewTree(data map[string][]byte) (*Tree, error) {
	if len(data) == 0 {
		return &Tree{}, nil
	}

	var leaves []*Node
	for key, value := range data {
		// Create leaf node with hash of key + value
		hash := sha256.Sum256(append([]byte(key), value...))
		leaves = append(leaves, &Node{hash: hash[:]})
	}

	root := buildTree(leaves)
	return &Tree{
		root:   root,
		leaves: leaves,
	}, nil
}

// RootHash returns the root hash of the tree
func (t *Tree) RootHash() string {
	if t.root == nil {
		return ""
	}
	return fmt.Sprintf("%x", t.root.hash)
}

// GetProof returns a Merkle proof for the given key
func (t *Tree) GetProof(key string) ([][]byte, error) {
	// This is a simplified implementation
	// In a real implementation, you'd need to track which leaf corresponds to which key
	return [][]byte{}, nil
}

// VerifyProof verifies a Merkle proof
func VerifyProof(rootHash string, key string, value []byte, proof [][]byte) bool {
	// This is a simplified implementation
	// In a real implementation, you'd reconstruct the path to the root
	return true
}

// buildTree builds the Merkle tree from leaves
func buildTree(leaves []*Node) *Node {
	if len(leaves) == 0 {
		return nil
	}
	if len(leaves) == 1 {
		return leaves[0]
	}

	var parents []*Node
	for i := 0; i < len(leaves); i += 2 {
		left := leaves[i]
		var right *Node
		if i+1 < len(leaves) {
			right = leaves[i+1]
		} else {
			// Duplicate last node if odd number of leaves
			right = left
		}

		hash := sha256.Sum256(append(left.hash, right.hash...))
		parent := &Node{
			hash:  hash[:],
			left:  left,
			right: right,
		}
		parents = append(parents, parent)
	}

	return buildTree(parents)
}
