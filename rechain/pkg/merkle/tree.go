package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// Node represents a node in the Merkle tree
type Node struct {
	Left   *Node
	Right  *Node
	Hash   []byte
	IsLeaf bool
	Key    []byte
	Value  []byte
}

// Tree represents a Merkle tree
type Tree struct {
	Root *Node
	leafs []*Node
}

// NewTree creates a new Merkle tree from a map of key-value pairs
func NewTree(data map[string][]byte) (*Tree, error) {
	if len(data) == 0 {
		return nil, errors.New("cannot create tree with no data")
	}

	var leafs []*Node

	// Create leaf nodes
	for k, v := range data {
		hash := hash(append([]byte(k), v...))
		leafs = append(leafs, &Node{
			Hash:   hash,
			IsLeaf: true,
			Key:    []byte(k),
			Value:  v,
		})
	}

	// Sort leafs by key for consistent ordering
	sortNodes(leafs)

	// Build the tree
	root := buildTree(leafs)

	return &Tree{
		Root:  root,
		leafs: leafs,
	}, nil
}

// buildTree recursively builds the Merkle tree from leaf nodes
func buildTree(nodes []*Node) *Node {
	if len(nodes) == 1 {
		return nodes[0]
	}

	var newLevel []*Node

	// Pair up nodes
	for i := 0; i < len(nodes); i += 2 {
		left := nodes[i]
		var right *Node

		// If there's an odd number of nodes, duplicate the last one
		if i+1 == len(nodes) {
			right = left
		} else {
			right = nodes[i+1]
		}

		// Create parent node
		parent := &Node{
			Left:  left,
			Right: right,
			Hash:  hash(append(left.Hash, right.Hash...)),
		}

		newLevel = append(newLevel, parent)
	}

	return buildTree(newLevel)
}

// GetProof returns the Merkle proof for a given key
func (t *Tree) GetProof(key []byte) ([][]byte, error) {
	var proof [][]byte
	var targetNode *Node

	// Find the target leaf node
	for _, node := range t.leafs {
		if string(node.Key) == string(key) {
			targetNode = node
			break
		}
	}

	if targetNode == nil {
		return nil, errors.New("key not found in tree")
	}

	// Build the proof by traversing up the tree
	current := targetNode
	for current != t.Root {
		parent := t.findParent(current)
		if parent == nil {
			break
		}

		if parent.Left == current {
			// If current is left child, add right sibling to proof
			proof = append(proof, parent.Right.Hash)
		} else {
			// If current is right child, add left sibling to proof
			proof = append(proof, parent.Left.Hash)
		}

		current = parent
	}

	return proof, nil
}

// VerifyProof verifies a Merkle proof for a key-value pair
func VerifyProof(rootHash []byte, key, value []byte, proof [][]byte) bool {
	hash := hash(append(key, value...))

	for _, sibling := range proof {
		hash = hash(append(hash, sibling...))
	}

	return string(hash) == string(rootHash)
}

// findParent finds the parent of a node in the tree
func (t *Tree) findParent(node *Node) *Node {
	if t.Root == nil || node == t.Root {
		return nil
	}
	return t.findParentHelper(t.Root, node)
}

func (t *Tree) findParentHelper(current, target *Node) *Node {
	if current == nil {
		return nil
	}

	if current.Left == target || current.Right == target {
		return current
	}

	if parent := t.findParentHelper(current.Left, target); parent != nil {
		return parent
	}

	return t.findParentHelper(current.Right, target)
}

// hash computes the SHA-256 hash of the input data
func hash(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

// sortNodes sorts nodes by their key for consistent ordering
func sortNodes(nodes []*Node) {
	// Simple bubble sort for demonstration
	// In production, use a more efficient sort
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if string(nodes[i].Key) > string(nodes[j].Key) {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
}

// RootHash returns the root hash of the Merkle tree
func (t *Tree) RootHash() string {
	if t.Root == nil {
		return ""
	}
	return hex.EncodeToString(t.Root.Hash)
}

// Get returns the value for a given key if it exists in the tree
func (t *Tree) Get(key []byte) ([]byte, bool) {
	for _, node := range t.leafs {
		if string(node.Key) == string(key) {
			return node.Value, true
		}
	}
	return nil, false
}
