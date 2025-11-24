# DeCube Crypto Library Implementation

## Overview
Create a Go library for DeCube that handles cryptographic operations including mTLS connections, Ed25519 signatures, Merkle proof verification, and key rotation.

## Tasks

### 1. Project Setup
- [ ] Create decub-crypto directory
- [ ] Initialize go.mod with proper module name
- [ ] Set up basic package structure

### 2. mTLS Connections
- [ ] Implement TLS certificate loading functions
- [ ] Create client TLS config with client certificate
- [ ] Create server TLS config with server certificate and CA
- [ ] Add functions for mutual TLS handshake verification

### 3. Ed25519 Digital Signatures
- [ ] Implement key pair generation
- [ ] Create Sign function for Ed25519
- [ ] Create Verify function for Ed25519 signatures
- [ ] Add signature encoding/decoding (base64)

### 4. Merkle Proof Verification
- [ ] Define Merkle proof structures for snapshots and transactions
- [ ] Implement proof verification against Merkle root
- [ ] Add support for both snapshot and transaction proofs
- [ ] Include hash calculation functions

### 5. Key Rotation
- [ ] Define global transaction structure for key rotation
- [ ] Implement signed key rotation transactions
- [ ] Add verification of rotation transactions
- [ ] Include key transition logic

### 6. Example Implementation
- [ ] Create example demonstrating snapshot proof verification
- [ ] Show verification against block header
- [ ] Include sample data and usage

### 7. Testing and Documentation
- [ ] Add unit tests for all functions
- [ ] Write comprehensive documentation
- [ ] Add usage examples in README
