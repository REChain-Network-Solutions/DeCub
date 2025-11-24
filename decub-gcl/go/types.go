package main

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// Transaction represents a transaction in the block
type Transaction struct {
	TxID   string `json:"tx_id"`
	Type   string `json:"type"`
	Origin string `json:"origin"`
	Payload string `json:"payload"`
	Sig    string `json:"sig"`
}

// Header represents the block header
type Header struct {
	Height     int       `json:"height"`
	PrevHash   string    `json:"prev_hash"`
	MerkleRoot string    `json:"merkle_root"`
	Proposer   string    `json:"proposer"`
	Timestamp  time.Time `json:"timestamp"`
}

// Block represents a block in the ledger
type Block struct {
	Header Header        `json:"header"`
	Txs    []Transaction `json:"txs"`
}

// MerkleNode represents a node in the Merkle tree
type MerkleNode struct {
	Hash  string
	Left  *MerkleNode
	Right *MerkleNode
}

// MerkleProof represents a proof for a transaction
type MerkleProof struct {
	Hashes []string `json:"hashes"`
	Index  int      `json:"index"`
}

// HashTransaction computes the hash of a transaction
func HashTransaction(tx Transaction) string {
	data := tx.TxID + tx.Type + tx.Origin + tx.Payload + tx.Sig
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// HashBlock computes the hash of a block
func HashBlock(block Block) string {
	data := block.Header.PrevHash + block.Header.MerkleRoot + block.Header.Proposer + block.Header.Timestamp.String()
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
