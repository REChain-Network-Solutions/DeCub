package decubcrypto

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"time"
)

// KeyRotationTransaction represents a signed global transaction for key rotation
type KeyRotationTransaction struct {
	TransactionID   string            `json:"transaction_id"`
	Timestamp       int64             `json:"timestamp"`
	NodeID          string            `json:"node_id"`
	OldPublicKey    string            `json:"old_public_key"`    // base64 encoded
	NewPublicKey    string            `json:"new_public_key"`    // base64 encoded
	RotationReason  string            `json:"rotation_reason"`
	SequenceNumber  uint64            `json:"sequence_number"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	Signature       string            `json:"signature"` // base64 encoded signature of the transaction
}

// KeyRotationManager manages key rotation operations
type KeyRotationManager struct {
	currentKeyPair *Ed25519KeyPair
	nodeID         string
	sequenceNumber uint64
}

// NewKeyRotationManager creates a new key rotation manager
func NewKeyRotationManager(nodeID string, initialKeyPair *Ed25519KeyPair) *KeyRotationManager {
	return &KeyRotationManager{
		currentKeyPair: initialKeyPair,
		nodeID:         nodeID,
		sequenceNumber: 0,
	}
}

// CreateKeyRotationTransaction creates a signed key rotation transaction
func (krm *KeyRotationManager) CreateKeyRotationTransaction(newKeyPair *Ed25519KeyPair, reason string) (*KeyRotationTransaction, error) {
	krm.sequenceNumber++

	transaction := &KeyRotationTransaction{
		TransactionID:  fmt.Sprintf("%s-%d-%d", krm.nodeID, time.Now().Unix(), krm.sequenceNumber),
		Timestamp:      time.Now().Unix(),
		NodeID:         krm.nodeID,
		OldPublicKey:   EncodeEd25519PublicKey(krm.currentKeyPair.PublicKey),
		NewPublicKey:   EncodeEd25519PublicKey(newKeyPair.PublicKey),
		RotationReason: reason,
		SequenceNumber: krm.sequenceNumber,
		Metadata:       make(map[string]string),
	}

	// Sign the transaction
	txData, err := json.Marshal(map[string]interface{}{
		"transaction_id":  transaction.TransactionID,
		"timestamp":       transaction.Timestamp,
		"node_id":         transaction.NodeID,
		"old_public_key":  transaction.OldPublicKey,
		"new_public_key":  transaction.NewPublicKey,
		"rotation_reason": transaction.RotationReason,
		"sequence_number": transaction.SequenceNumber,
		"metadata":        transaction.Metadata,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transaction for signing: %w", err)
	}

	signature, err := SignWithEd25519Base64(krm.currentKeyPair.PrivateKey, txData)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	transaction.Signature = signature

	// Update current key pair after successful transaction creation
	krm.currentKeyPair = newKeyPair

	return transaction, nil
}

// VerifyKeyRotationTransaction verifies a key rotation transaction
func VerifyKeyRotationTransaction(transaction *KeyRotationTransaction) error {
	// Decode the old public key for verification
	oldPublicKey, err := DecodeEd25519PublicKey(transaction.OldPublicKey)
	if err != nil {
		return fmt.Errorf("failed to decode old public key: %w", err)
	}

	// Prepare transaction data for verification (same as signing)
	txData, err := json.Marshal(map[string]interface{}{
		"transaction_id":  transaction.TransactionID,
		"timestamp":       transaction.Timestamp,
		"node_id":         transaction.NodeID,
		"old_public_key":  transaction.OldPublicKey,
		"new_public_key":  transaction.NewPublicKey,
		"rotation_reason": transaction.RotationReason,
		"sequence_number": transaction.SequenceNumber,
		"metadata":        transaction.Metadata,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal transaction for verification: %w", err)
	}

	// Verify signature
	valid, err := VerifyEd25519SignatureBase64(oldPublicKey, txData, transaction.Signature)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid signature for key rotation transaction")
	}

	// Validate transaction fields
	if transaction.NodeID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	if transaction.Timestamp <= 0 {
		return fmt.Errorf("invalid timestamp")
	}
	if transaction.SequenceNumber == 0 {
		return fmt.Errorf("sequence number must be greater than 0")
	}

	// Validate public keys
	newPublicKey, err := DecodeEd25519PublicKey(transaction.NewPublicKey)
	if err != nil {
		return fmt.Errorf("failed to decode new public key: %w", err)
	}
	if len(newPublicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid new public key size")
	}

	return nil
}

// ApplyKeyRotation applies a verified key rotation transaction
func (krm *KeyRotationManager) ApplyKeyRotation(transaction *KeyRotationTransaction) error {
	// Verify the transaction first
	if err := VerifyKeyRotationTransaction(transaction); err != nil {
		return fmt.Errorf("transaction verification failed: %w", err)
	}

	// Decode new public key
	newPublicKey, err := DecodeEd25519PublicKey(transaction.NewPublicKey)
	if err != nil {
		return fmt.Errorf("failed to decode new public key: %w", err)
	}

	// Update key pair (assuming the private key is already updated)
	krm.currentKeyPair.PublicKey = newPublicKey
	krm.sequenceNumber = transaction.SequenceNumber

	return nil
}

// GetCurrentPublicKey returns the current public key
func (krm *KeyRotationManager) GetCurrentPublicKey() ed25519.PublicKey {
	return krm.currentKeyPair.PublicKey
}

// GetCurrentSequenceNumber returns the current sequence number
func (krm *KeyRotationManager) GetCurrentSequenceNumber() uint64 {
	return krm.sequenceNumber
}
