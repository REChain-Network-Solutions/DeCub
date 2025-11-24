package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Validator represents a validator
type Validator struct {
	ID   string
	PubKey string
}

// Consensus simulates BFT consensus with quorum signatures
type Consensus struct {
	Validators []Validator
	Threshold  int // >=2/3
}

// NewConsensus creates a new consensus instance
func NewConsensus(validators []Validator) *Consensus {
	threshold := (2 * len(validators)) / 3
	return &Consensus{Validators: validators, Threshold: threshold}
}

// SignBlock simulates signing the block by validators
func (c *Consensus) SignBlock(block Block) []string {
	var signatures []string
	for _, v := range c.Validators {
		// Simulate signature (in real impl, use crypto)
		data := fmt.Sprintf("%s%s", v.ID, HashBlock(block))
		hash := sha256.Sum256([]byte(data))
		sig := hex.EncodeToString(hash[:])
		signatures = append(signatures, sig)
	}
	return signatures
}

// VerifyQuorum checks if there are enough signatures (>=2/3)
func (c *Consensus) VerifyQuorum(signatures []string) bool {
	return len(signatures) >= c.Threshold
}

// ProposeBlock simulates proposing a new block
func (c *Consensus) ProposeBlock(height int, prevHash string, txs []Transaction, proposer string) Block {
	root, _ := BuildMerkleTree(txs)
	header := Header{
		Height:     height,
		PrevHash:   prevHash,
		MerkleRoot: root.Hash,
		Proposer:   proposer,
		Timestamp:  time.Now(),
	}
	return Block{Header: header, Txs: txs}
}
