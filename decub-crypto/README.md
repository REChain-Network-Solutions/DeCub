# DeCube Crypto Library

A comprehensive Go library for cryptographic operations in the DeCube distributed system, providing mTLS connections, Ed25519 digital signatures, Merkle proof verification, and key rotation mechanisms.

## Features

- **Mutual TLS (mTLS)**: Secure connections between DeCube nodes with client and server certificate authentication
- **Ed25519 Signatures**: Fast, secure digital signatures for data integrity and authentication
- **Merkle Proof Verification**: Efficient verification of snapshots and transactions against Merkle roots
- **Key Rotation**: Secure key rotation through signed global transactions

## Installation

```bash
go get github.com/decubate/decub-crypto
```

## Usage

### mTLS Connections

```go
package main

import (
    "crypto/tls"
    decubcrypto "github.com/decubate/decub-crypto"
)

func main() {
    // Load certificates
    clientCert, err := decubcrypto.LoadTLSCertificates("client.crt", "client.key")
    if err != nil {
        panic(err)
    }

    caCertPool, err := decubcrypto.LoadCACertificate("ca.crt")
    if err != nil {
        panic(err)
    }

    // Create client TLS config
    clientTLSConfig := decubcrypto.CreateClientTLSConfig(clientCert, caCertPool)

    // Use with HTTP client
    client := &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: clientTLSConfig,
        },
    }
}
```

### Ed25519 Digital Signatures

```go
package main

import (
    decubcrypto "github.com/decubate/decub-crypto"
)

func main() {
    // Generate key pair
    keyPair, err := decubcrypto.GenerateEd25519KeyPair()
    if err != nil {
        panic(err)
    }

    message := []byte("Hello, DeCube!")

    // Sign message
    signature, err := decubcrypto.SignWithEd25519(keyPair.PrivateKey, message)
    if err != nil {
        panic(err)
    }

    // Verify signature
    valid := decubcrypto.VerifyEd25519Signature(keyPair.PublicKey, message, signature)
    fmt.Printf("Signature valid: %t\n", valid)
}
```

### Merkle Proof Verification

```go
package main

import (
    decubcrypto "github.com/decubate/decub-crypto"
)

func main() {
    // Create snapshot data
    snapshots := []map[string]interface{}{
        {"id": "snap1", "size": 1024},
        {"id": "snap2", "size": 2048},
    }

    // Build Merkle tree
    var leafHashes []string
    for _, snap := range snapshots {
        data, _ := json.Marshal(snap)
        leafHashes = append(leafHashes, decubcrypto.ComputeLeafHash(data))
    }

    rootHash, _ := decubcrypto.BuildMerkleTree(leafHashes)

    // Generate and verify proof
    proof, _ := decubcrypto.GenerateMerkleProof(leafHashes, 0)
    valid := decubcrypto.VerifyMerkleProof(*proof)
}
```

### Key Rotation

```go
package main

import (
    decubcrypto "github.com/decubate/decub-crypto"
)

func main() {
    // Create key rotation manager
    keyPair, _ := decubcrypto.GenerateEd25519KeyPair()
    manager := decubcrypto.NewKeyRotationManager("node-1", keyPair)

    // Generate new key pair
    newKeyPair, _ := decubcrypto.GenerateEd25519KeyPair()

    // Create rotation transaction
    tx, err := manager.CreateKeyRotationTransaction(newKeyPair, "Scheduled rotation")
    if err != nil {
        panic(err)
    }

    // Verify transaction
    err = decubcrypto.VerifyKeyRotationTransaction(tx)
    if err != nil {
        panic(err)
    }
}
```

## Example

Run the complete example:

```bash
go run example.go
```

This demonstrates:
- Ed25519 signature creation and verification
- Merkle proof generation and verification for snapshots
- Key rotation with signed transactions
- Transaction proof verification

## API Reference

### TLS Functions

- `LoadTLSCertificates(certFile, keyFile string) (tls.Certificate, error)`
- `LoadCACertificate(caFile string) (*x509.CertPool, error)`
- `CreateClientTLSConfig(clientCert tls.Certificate, caCertPool *x509.CertPool) *tls.Config`
- `CreateServerTLSConfig(serverCert tls.Certificate, caCertPool *x509.CertPool) *tls.Config`
- `VerifyMutualTLS(conn *tls.Conn) error`

### Signature Functions

- `GenerateEd25519KeyPair() (*Ed25519KeyPair, error)`
- `SignWithEd25519(privateKey ed25519.PrivateKey, data []byte) ([]byte, error)`
- `VerifyEd25519Signature(publicKey ed25519.PublicKey, data []byte, signature []byte) bool`
- `SignWithEd25519Base64(privateKey ed25519.PrivateKey, data []byte) (string, error)`
- `VerifyEd25519SignatureBase64(publicKey ed25519.PublicKey, data []byte, signatureBase64 string) (bool, error)`

### Merkle Proof Functions

- `BuildMerkleTree(leafHashes []string) (string, error)`
- `GenerateMerkleProof(leafHashes []string, targetIndex uint64) (*MerkleProof, error)`
- `VerifyMerkleProof(proof MerkleProof) bool`
- `VerifySnapshotProof(snapshotProof SnapshotProof, blockHeaderRoot string) error`
- `VerifyTransactionProof(txProof TransactionProof, expectedRoot string) error`

### Key Rotation Functions

- `NewKeyRotationManager(nodeID string, initialKeyPair *Ed25519KeyPair) *KeyRotationManager`
- `CreateKeyRotationTransaction(newKeyPair *Ed25519KeyPair, reason string) (*KeyRotationTransaction, error)`
- `VerifyKeyRotationTransaction(transaction *KeyRotationTransaction) error`
- `ApplyKeyRotation(transaction *KeyRotationTransaction) error`

## Security Considerations

- Always validate certificate chains and expiration dates
- Use sufficiently long keys (Ed25519 provides 128-bit security)
- Implement proper key rotation schedules
- Store private keys securely (encrypted at rest)
- Verify all cryptographic operations return success

## License

This library is part of the DeCube project. See project license for details.
