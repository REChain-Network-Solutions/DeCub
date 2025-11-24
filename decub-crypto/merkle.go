package decubcrypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// MerkleProof represents a Merkle proof for verification
type MerkleProof struct {
	RootHash  string   `json:"root_hash"`
	LeafHash  string   `json:"leaf_hash"`
	Proof     []string `json:"proof"`
	Index     uint64   `json:"index"`
	NumLeaves uint64   `json:"num_leaves"`
}

// SnapshotProof represents a proof for a snapshot
type SnapshotProof struct {
	MerkleProof
	SnapshotID string                 `json:"snapshot_id"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// TransactionProof represents a proof for a transaction
type TransactionProof struct {
	MerkleProof
	TxID   string `json:"tx_id"`
	TxData []byte `json:"tx_data"`
}

// ComputeSHA256Hash computes SHA256 hash of data
func ComputeSHA256Hash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// ComputeLeafHash computes hash for a leaf node
func ComputeLeafHash(data []byte) string {
	return ComputeSHA256Hash(data)
}

// ComputeInternalHash computes hash for internal nodes
func ComputeInternalHash(left, right string) string {
	combined := left + right
	return ComputeSHA256Hash([]byte(combined))
}

// BuildMerkleTree builds a Merkle tree from leaf hashes and returns the root
func BuildMerkleTree(leafHashes []string) (string, error) {
	if len(leafHashes) == 0 {
		return "", fmt.Errorf("cannot build Merkle tree with empty leaves")
	}

	currentLevel := make([]string, len(leafHashes))
	copy(currentLevel, leafHashes)

	for len(currentLevel) > 1 {
		var nextLevel []string

		for i := 0; i < len(currentLevel); i += 2 {
			left := currentLevel[i]
			var right string
			if i+1 < len(currentLevel) {
				right = currentLevel[i+1]
			} else {
				// Duplicate last hash if odd number
				right = left
			}
			parentHash := ComputeInternalHash(left, right)
			nextLevel = append(nextLevel, parentHash)
		}

		currentLevel = nextLevel
	}

	return currentLevel[0], nil
}

// VerifyMerkleProof verifies a Merkle proof against a root hash
func VerifyMerkleProof(proof MerkleProof) bool {
	if proof.Index >= proof.NumLeaves {
		return false
	}

	currentHash := proof.LeafHash
	index := proof.Index

	for _, siblingHash := range proof.Proof {
		if index%2 == 0 {
			// Left child
			currentHash = ComputeInternalHash(currentHash, siblingHash)
		} else {
			// Right child
			currentHash = ComputeInternalHash(siblingHash, currentHash)
		}
		index /= 2
	}

	return currentHash == proof.RootHash
}

// GenerateMerkleProof generates a Merkle proof for a leaf at given index
func GenerateMerkleProof(leafHashes []string, targetIndex uint64) (*MerkleProof, error) {
	if targetIndex >= uint64(len(leafHashes)) {
		return nil, fmt.Errorf("target index out of bounds")
	}

	numLeaves := uint64(len(leafHashes))
	rootHash, err := BuildMerkleTree(leafHashes)
	if err != nil {
		return nil, err
	}

	proof := &MerkleProof{
		RootHash:  rootHash,
		LeafHash:  leafHashes[targetIndex],
		Proof:     []string{},
		Index:     targetIndex,
		NumLeaves: numLeaves,
	}

	// Build the tree and collect proof
	currentLevel := make([]string, len(leafHashes))
	copy(currentLevel, leafHashes)

	index := targetIndex

	for len(currentLevel) > 1 {
		var nextLevel []string
		siblingIndex := index ^ 1 // Flip the least significant bit to get sibling

		if siblingIndex < uint64(len(currentLevel)) {
			proof.Proof = append(proof.Proof, currentLevel[siblingIndex])
		} else {
			// If no sibling exists, duplicate the current hash
			proof.Proof = append(proof.Proof, currentLevel[index])
		}

		// Build next level
		for i := 0; i < len(currentLevel); i += 2 {
			left := currentLevel[i]
			var right string
			if i+1 < len(currentLevel) {
				right = currentLevel[i+1]
			} else {
				right = left
			}
			parentHash := ComputeInternalHash(left, right)
			nextLevel = append(nextLevel, parentHash)
		}

		currentLevel = nextLevel
		index /= 2
	}

	return proof, nil
}

// VerifySnapshotProof verifies a snapshot proof against a block header
func VerifySnapshotProof(snapshotProof SnapshotProof, blockHeaderRoot string) error {
	// Verify the Merkle proof
	if !VerifyMerkleProof(snapshotProof.MerkleProof) {
		return fmt.Errorf("invalid Merkle proof for snapshot %s", snapshotProof.SnapshotID)
	}

	// Verify against block header root
	if snapshotProof.RootHash != blockHeaderRoot {
		return fmt.Errorf("snapshot proof root hash does not match block header root")
	}

	return nil
}

// VerifyTransactionProof verifies a transaction proof
func VerifyTransactionProof(txProof TransactionProof, expectedRoot string) error {
	// Verify the Merkle proof
	if !VerifyMerkleProof(txProof.MerkleProof) {
		return fmt.Errorf("invalid Merkle proof for transaction %s", txProof.TxID)
	}

	// Verify against expected root
	if txProof.RootHash != expectedRoot {
		return fmt.Errorf("transaction proof root hash does not match expected root")
	}

	// Verify transaction data hash matches leaf hash
	computedLeafHash := ComputeLeafHash(txProof.TxData)
	if computedLeafHash != txProof.LeafHash {
		return fmt.Errorf("transaction data hash does not match proof leaf hash")
	}

	return nil
}
