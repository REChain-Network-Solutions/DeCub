package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"

	"github.com/google/uuid"
)

// KeyManager manages encryption keys
type KeyManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewKeyManager creates a new key manager
func NewKeyManager() (*KeyManager, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	return &KeyManager{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}, nil
}

// EncryptData encrypts data with AES-GCM
func (km *KeyManager) EncryptData(plaintext []byte) ([]byte, []byte, error) {
	// Generate random key for AES
	key := make([]byte, 32) // 256-bit key
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, nil, fmt.Errorf("failed to generate AES key: %w", err)
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Encrypt the AES key with RSA
	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, km.publicKey, key, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt AES key: %w", err)
	}

	return ciphertext, encryptedKey, nil
}

// DecryptData decrypts data with AES-GCM
func (km *KeyManager) DecryptData(ciphertext, encryptedKey []byte) ([]byte, error) {
	// Decrypt the AES key
	key, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, km.privateKey, encryptedKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES key: %w", err)
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// SignData signs data with RSA-PSS
func (km *KeyManager) SignData(data []byte) ([]byte, error) {
	hashed := sha256.Sum256(data)
	signature, err := rsa.SignPSS(rand.Reader, km.privateKey, 0, hashed[:], nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}
	return signature, nil
}

// VerifySignature verifies an RSA-PSS signature
func (km *KeyManager) VerifySignature(data, signature []byte) error {
	hashed := sha256.Sum256(data)
	return rsa.VerifyPSS(km.publicKey, 0, hashed[:], signature, nil)
}

// GenerateNonce generates a random nonce
func GenerateNonce(size int) ([]byte, error) {
	nonce := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	return nonce, nil
}

// TransactionSigner handles transaction signing
type TransactionSigner struct {
	keyManager *KeyManager
	nodeID     string
}

// NewTransactionSigner creates a new transaction signer
func NewTransactionSigner(nodeID string) (*TransactionSigner, error) {
	km, err := NewKeyManager()
	if err != nil {
		return nil, err
	}

	return &TransactionSigner{
		keyManager: km,
		nodeID:     nodeID,
	}, nil
}

// SignTransaction signs a transaction
func (ts *TransactionSigner) SignTransaction(txID string, txData []byte) ([]byte, error) {
	// Create signing payload
	payload := fmt.Sprintf("%s:%s:%s", ts.nodeID, txID, string(txData))

	return ts.keyManager.SignData([]byte(payload))
}

// VerifyTransaction verifies a transaction signature
func (ts *TransactionSigner) VerifyTransaction(txID string, txData, signature []byte, signerPublicKey *rsa.PublicKey) error {
	payload := fmt.Sprintf("%s:%s:%s", ts.nodeID, txID, string(txData))

	hashed := sha256.Sum256([]byte(payload))
	return rsa.VerifyPSS(signerPublicKey, 0, hashed[:], signature, nil)
}

// HSMManager provides HSM integration stubs
type HSMManager struct {
	connected bool
}

// NewHSMManager creates a new HSM manager
func NewHSMManager(hsmAddress string) (*HSMManager, error) {
	// Stub implementation - in production, connect to actual HSM
	log.Printf("Connecting to HSM at %s (stub)", hsmAddress)

	return &HSMManager{
		connected: true, // Pretend we're connected
	}, nil
}

// SignWithHSM signs data using HSM (stub)
func (hsm *HSMManager) SignWithHSM(data []byte) ([]byte, error) {
	if !hsm.connected {
		return nil, fmt.Errorf("HSM not connected")
	}

	// Stub - in production, use actual HSM signing
	log.Printf("Signing with HSM (stub): %d bytes", len(data))

	// Generate a fake signature for demo
	signature := make([]byte, 256)
	if _, err := io.ReadFull(rand.Reader, signature); err != nil {
		return nil, err
	}

	return signature, nil
}

// GenerateKeyWithHSM generates a key using HSM (stub)
func (hsm *HSMManager) GenerateKeyWithHSM(keyID string) error {
	if !hsm.connected {
		return fmt.Errorf("HSM not connected")
	}

	log.Printf("Generating key %s with HSM (stub)", keyID)
	return nil
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	CertFile string
	KeyFile  string
	CAFile   string
}

// LoadTLSConfig loads TLS configuration (stub)
func LoadTLSConfig(certFile, keyFile, caFile string) (*TLSConfig, error) {
	// Stub - in production, load actual certificates
	return &TLSConfig{
		CertFile: certFile,
		KeyFile:  keyFile,
		CAFile:   caFile,
	}, nil
}

// ValidateCertificate validates a certificate (stub)
func ValidateCertificate(certPEM []byte) error {
	// Stub - in production, perform actual certificate validation
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("invalid PEM block")
	}

	_, err := x509.ParseCertificate(block.Bytes)
	return err
}

// GenerateCertID generates a unique certificate ID
func GenerateCertID() string {
	return uuid.New().String()
}

// AuditLogger logs security events
type AuditLogger struct {
	enabled bool
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(enabled bool) *AuditLogger {
	return &AuditLogger{enabled: enabled}
}

// LogSecurityEvent logs a security event
func (al *AuditLogger) LogSecurityEvent(eventType, details string) {
	if !al.enabled {
		return
	}

	log.Printf("SECURITY EVENT [%s]: %s", eventType, details)
}

// LogAccess logs an access event
func (al *AuditLogger) LogAccess(resource, action, userID string) {
	if !al.enabled {
		return
	}

	log.Printf("ACCESS: %s %s by %s", action, resource, userID)
}
