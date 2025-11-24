package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// CatalogMerkleNode represents a node in the catalog Merkle tree
type CatalogMerkleNode struct {
	Hash  string              `json:"hash"`
	Left  *CatalogMerkleNode  `json:"left,omitempty"`
	Right *CatalogMerkleNode  `json:"right,omitempty"`
	Data  *CatalogMerkleData  `json:"data,omitempty"`
}

// CatalogMerkleData represents the data stored in a Merkle tree leaf
type CatalogMerkleData struct {
	Type     string                 `json:"type"` // "snapshot" or "image"
	ID       string                 `json:"id"`
	Metadata map[string]interface{} `json:"metadata"`
	Version  int64                  `json:"version"` // Vector clock timestamp
}

// CatalogMerkleTree manages Merkle tree operations for the catalog
type CatalogMerkleTree struct {
	root *CatalogMerkleNode
}

// NewCatalogMerkleTree creates a new catalog Merkle tree
func NewCatalogMerkleTree() *CatalogMerkleTree {
	return &CatalogMerkleTree{}
}

// BuildFromCatalog builds the Merkle tree from catalog data
func (mt *CatalogMerkleTree) BuildFromCatalog(snapshots, images map[string]*LWWRegister) error {
	var leaves []*CatalogMerkleData

	// Add snapshots
	for id, register := range snapshots {
		if register != nil {
			data := register.Get()
			if metadata, ok := data.(map[string]interface{}); ok {
				leaves = append(leaves, &CatalogMerkleData{
					Type:     "snapshot",
					ID:       id,
					Metadata: metadata,
					Version:  register.timestamp,
				})
			}
		}
	}

	// Add images
	for id, register := range images {
		if register != nil {
			data := register.Get()
			if metadata, ok := data.(map[string]interface{}); ok {
				leaves = append(leaves, &CatalogMerkleData{
					Type:     "image",
					ID:       id,
					Metadata: metadata,
					Version:  register.timestamp,
				})
			}
		}
	}

	// Sort leaves for consistent tree building
	sort.Slice(leaves, func(i, j int) bool {
		if leaves[i].Type != leaves[j].Type {
			return leaves[i].Type < leaves[j].Type
		}
		return leaves[i].ID < leaves[j].ID
	})

	root, err := mt.buildTree(leaves)
	if err != nil {
		return err
	}

	mt.root = root
	return nil
}

// buildTree recursively builds the Merkle tree from leaves
func (mt *CatalogMerkleTree) buildTree(leaves []*CatalogMerkleData) (*CatalogMerkleNode, error) {
	if len(leaves) == 0 {
		return nil, nil
	}

	if len(leaves) == 1 {
		// Leaf node
		dataBytes, err := json.Marshal(leaves[0])
		if err != nil {
			return nil, fmt.Errorf("failed to marshal leaf data: %w", err)
		}

		hash := sha256.Sum256(dataBytes)
		return &CatalogMerkleNode{
			Hash: hex.EncodeToString(hash[:]),
			Data: leaves[0],
		}, nil
	}

	// Internal node
	mid := len(leaves) / 2
	left, err := mt.buildTree(leaves[:mid])
	if err != nil {
		return nil, err
	}

	right, err := mt.buildTree(leaves[mid:])
	if err != nil {
		return nil, err
	}

	// Combine hashes
	var combined string
	if left != nil {
		combined += left.Hash
	}
	if right != nil {
		combined += right.Hash
	}

	hash := sha256.Sum256([]byte(combined))
	return &CatalogMerkleNode{
		Hash:  hex.EncodeToString(hash[:]),
		Left:  left,
		Right: right,
	}, nil
}

// GetRootHash returns the root hash of the Merkle tree
func (mt *CatalogMerkleTree) GetRootHash() string {
	if mt.root == nil {
		return ""
	}
	return mt.root.Hash
}

// CompareRoot compares this tree's root with another root hash
func (mt *CatalogMerkleTree) CompareRoot(otherRoot string) bool {
	return mt.GetRootHash() == otherRoot
}

// GenerateProof generates a Merkle proof for a specific catalog item
func (mt *CatalogMerkleTree) GenerateProof(itemType, itemID string) (*MerkleProof, error) {
	if mt.root == nil {
		return nil, fmt.Errorf("Merkle tree not built")
	}

	// Find the leaf node for this item
	leafNode, path := mt.findLeaf(mt.root, itemType, itemID, []string{})
	if leafNode == nil {
		return nil, fmt.Errorf("item %s:%s not found in Merkle tree", itemType, itemID)
	}

	proof := &MerkleProof{
		RootHash:  mt.GetRootHash(),
		LeafHash:  leafNode.Hash,
		Proof:     path,
		Index:     0, // Simplified - would need proper indexing
		NumLeaves: 1, // Simplified
	}

	return proof, nil
}

// findLeaf recursively finds a leaf node and builds the proof path
func (mt *CatalogMerkleTree) findLeaf(node *CatalogMerkleNode, itemType, itemID string, path []string) (*CatalogMerkleNode, []string) {
	if node == nil {
		return nil, path
	}

	// Check if this is the target leaf
	if node.Data != nil && node.Data.Type == itemType && node.Data.ID == itemID {
		return node, path
	}

	// Search left subtree
	if node.Left != nil {
		if found, newPath := mt.findLeaf(node.Left, itemType, itemID, path); found != nil {
			// Add sibling hash to path
			if node.Right != nil {
				newPath = append(newPath, node.Right.Hash)
			}
			return found, newPath
		}
	}

	// Search right subtree
	if node.Right != nil {
		if found, newPath := mt.findLeaf(node.Right, itemType, itemID, path); found != nil {
			// Add sibling hash to path
			if node.Left != nil {
				newPath = append(newPath, node.Left.Hash)
			}
			return found, newPath
		}
	}

	return nil, path
}

// VerifyProof verifies a Merkle proof
func (mt *CatalogMerkleTree) VerifyProof(proof *MerkleProof) bool {
	if proof == nil {
		return false
	}

	// Start with the leaf hash
	currentHash := proof.LeafHash

	// Apply proof hashes
	for _, siblingHash := range proof.Proof {
		// Simplified verification - in practice, need proper left/right ordering
		combined := currentHash + siblingHash
		hash := sha256.Sum256([]byte(combined))
		currentHash = hex.EncodeToString(hash[:])
	}

	return currentHash == proof.RootHash
}

// Serialize serializes the Merkle tree to JSON
func (mt *CatalogMerkleTree) Serialize() ([]byte, error) {
	return json.Marshal(mt.root)
}

// Deserialize deserializes the Merkle tree from JSON
func (mt *CatalogMerkleTree) Deserialize(data []byte) error {
	return json.Unmarshal(data, &mt.root)
}

// GetTreeStats returns statistics about the Merkle tree
func (mt *CatalogMerkleTree) GetTreeStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["root_hash"] = mt.GetRootHash()
	stats["height"] = mt.calculateHeight(mt.root)
	stats["leaf_count"] = mt.countLeaves(mt.root)
	return stats
}

// calculateHeight calculates the height of the tree
func (mt *CatalogMerkleTree) calculateHeight(node *CatalogMerkleNode) int {
	if node == nil {
		return 0
	}

	leftHeight := mt.calculateHeight(node.Left)
	rightHeight := mt.calculateHeight(node.Right)

	if leftHeight > rightHeight {
		return leftHeight + 1
	}
	return rightHeight + 1
}

// countLeaves counts the number of leaf nodes
func (mt *CatalogMerkleTree) countLeaves(node *CatalogMerkleNode) int {
	if node == nil {
		return 0
	}

	if node.Data != nil {
		return 1
	}

	return mt.countLeaves(node.Left) + mt.countLeaves(node.Right)
}

// MerkleProof represents a Merkle proof (simplified version)
type MerkleProof struct {
	RootHash  string   `json:"root_hash"`
	LeafHash  string   `json:"leaf_hash"`
	Proof     []string `json:"proof"`
	Index     uint64   `json:"index"`
	NumLeaves uint64   `json:"num_leaves"`
}
