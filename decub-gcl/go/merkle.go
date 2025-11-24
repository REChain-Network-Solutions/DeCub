package main

import (
	"crypto/sha256"
	"encoding/hex"
)

// BuildMerkleTree builds a Merkle tree from transactions and returns the root hash
func BuildMerkleTree(txs []Transaction) (*MerkleNode, string) {
	if len(txs) == 0 {
		return nil, ""
	}

	nodes := make([]*MerkleNode, len(txs))
	for i, tx := range txs {
		hash := HashTransaction(tx)
		nodes[i] = &MerkleNode{Hash: hash}
	}

	for len(nodes) > 1 {
		var newNodes []*MerkleNode
		for i := 0; i < len(nodes); i += 2 {
			left := nodes[i]
			var right *MerkleNode
			if i+1 < len(nodes) {
				right = nodes[i+1]
			} else {
				right = left // Duplicate for odd number
			}
			hash := sha256.Sum256([]byte(left.Hash + right.Hash))
			newNodes = append(newNodes, &MerkleNode{Hash: hex.EncodeToString(hash[:]), Left: left, Right: right})
		}
		nodes = newNodes
	}

	return nodes[0], nodes[0].Hash
}

// GenerateMerkleProof generates a Merkle proof for a transaction at the given index
func GenerateMerkleProof(root *MerkleNode, index int) MerkleProof {
	var proof MerkleProof
	proof.Index = index
	current := root
	for current.Left != nil || current.Right != nil {
		if index%2 == 0 {
			if current.Right != nil {
				proof.Hashes = append(proof.Hashes, current.Right.Hash)
			}
		} else {
			if current.Left != nil {
				proof.Hashes = append(proof.Hashes, current.Left.Hash)
			}
		}
		if index%2 == 0 {
			current = current.Left
		} else {
			current = current.Right
		}
		index /= 2
	}
	return proof
}
