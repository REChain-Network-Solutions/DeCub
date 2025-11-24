package decubcrypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// Ed25519KeyPair represents an Ed25519 public/private key pair
type Ed25519KeyPair struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

// GenerateEd25519KeyPair generates a new Ed25519 key pair
func GenerateEd25519KeyPair() (*Ed25519KeyPair, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Ed25519 key pair: %w", err)
	}

	return &Ed25519KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}

// SignWithEd25519 signs data using Ed25519 private key
func SignWithEd25519(privateKey ed25519.PrivateKey, data []byte) ([]byte, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: expected %d, got %d", ed25519.PrivateKeySize, len(privateKey))
	}

	signature := ed25519.Sign(privateKey, data)
	return signature, nil
}

// VerifyEd25519Signature verifies an Ed25519 signature
func VerifyEd25519Signature(publicKey ed25519.PublicKey, data []byte, signature []byte) bool {
	if len(publicKey) != ed25519.PublicKeySize {
		return false
	}
	if len(signature) != ed25519.SignatureSize {
		return false
	}

	return ed25519.Verify(publicKey, data, signature)
}

// SignWithEd25519Base64 signs data and returns base64-encoded signature
func SignWithEd25519Base64(privateKey ed25519.PrivateKey, data []byte) (string, error) {
	signature, err := SignWithEd25519(privateKey, data)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifyEd25519SignatureBase64 verifies a base64-encoded Ed25519 signature
func VerifyEd25519SignatureBase64(publicKey ed25519.PublicKey, data []byte, signatureBase64 string) (bool, error) {
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}

	return VerifyEd25519Signature(publicKey, data, signature), nil
}

// EncodeEd25519PublicKey encodes public key to base64
func EncodeEd25519PublicKey(publicKey ed25519.PublicKey) string {
	return base64.StdEncoding.EncodeToString(publicKey)
}

// DecodeEd25519PublicKey decodes base64-encoded public key
func DecodeEd25519PublicKey(encodedKey string) (ed25519.PublicKey, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	if len(keyBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: expected %d, got %d", ed25519.PublicKeySize, len(keyBytes))
	}
	return ed25519.PublicKey(keyBytes), nil
}
