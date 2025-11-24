package main

import (
	"encoding/json"
	"fmt"
	"log"

	decubcrypto "github.com/decubate/decub-crypto"
)

func main() {
	fmt.Println("DeCube Crypto Library Example")
	fmt.Println("============================")

	// Example 1: Ed25519 Signatures
	fmt.Println("\n1. Ed25519 Digital Signatures")
	keyPair, err := decubcrypto.GenerateEd25519KeyPair()
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}

	message := []byte("Hello, DeCube!")
	signature, err := decubcrypto.SignWithEd25519(keyPair.PrivateKey, message)
	if err != nil {
		log.Fatalf("Failed to sign message: %v", err)
	}

	valid := decubcrypto.VerifyEd25519Signature(keyPair.PublicKey, message, signature)
	fmt.Printf("Message: %s\n", message)
	fmt.Printf("Signature valid: %t\n", valid)

	// Example 2: Merkle Proof Verification for Snapshot
	fmt.Println("\n2. Merkle Proof Verification for Snapshot")

	// Create some snapshot data
	snapshots := []map[string]interface{}{
		{"id": "snap1", "size": 1024, "cluster": "prod"},
		{"id": "snap2", "size": 2048, "cluster": "dev"},
		{"id": "snap3", "size": 512, "cluster": "test"},
	}

	// Convert to JSON and hash
	var leafHashes []string
	for _, snap := range snapshots {
		data, _ := json.Marshal(snap)
		leafHashes = append(leafHashes, decubcrypto.ComputeLeafHash(data))
	}

	// Build Merkle tree
	rootHash, err := decubcrypto.BuildMerkleTree(leafHashes)
	if err != nil {
		log.Fatalf("Failed to build Merkle tree: %v", err)
	}
	fmt.Printf("Merkle Root: %s\n", rootHash)

	// Generate proof for first snapshot
	proof, err := decubcrypto.GenerateMerkleProof(leafHashes, 0)
	if err != nil {
		log.Fatalf("Failed to generate proof: %v", err)
	}

	// Create snapshot proof
	snapshotProof := decubcrypto.SnapshotProof{
		MerkleProof: *proof,
		SnapshotID:  "snap1",
		Metadata:    snapshots[0],
	}

	// Verify snapshot proof against block header (simulated)
	blockHeaderRoot := rootHash // In real scenario, this comes from block header
	err = decubcrypto.VerifySnapshotProof(snapshotProof, blockHeaderRoot)
	if err != nil {
		log.Fatalf("Snapshot proof verification failed: %v", err)
	}
	fmt.Printf("Snapshot proof verified successfully for snapshot: %s\n", snapshotProof.SnapshotID)

	// Example 3: Key Rotation
	fmt.Println("\n3. Key Rotation with Signed Transaction")

	// Create key rotation manager
	rotationManager := decubcrypto.NewKeyRotationManager("node-1", keyPair)

	// Generate new key pair for rotation
	newKeyPair, err := decubcrypto.GenerateEd25519KeyPair()
	if err != nil {
		log.Fatalf("Failed to generate new key pair: %v", err)
	}

	// Create key rotation transaction
	rotationTx, err := rotationManager.CreateKeyRotationTransaction(newKeyPair, "Scheduled key rotation")
	if err != nil {
		log.Fatalf("Failed to create rotation transaction: %v", err)
	}

	fmt.Printf("Key rotation transaction created:\n")
	fmt.Printf("  Transaction ID: %s\n", rotationTx.TransactionID)
	fmt.Printf("  Node ID: %s\n", rotationTx.NodeID)
	fmt.Printf("  Sequence: %d\n", rotationTx.SequenceNumber)
	fmt.Printf("  Reason: %s\n", rotationTx.RotationReason)

	// Verify the rotation transaction
	err = decubcrypto.VerifyKeyRotationTransaction(rotationTx)
	if err != nil {
		log.Fatalf("Key rotation transaction verification failed: %v", err)
	}
	fmt.Println("Key rotation transaction verified successfully")

	// Example 4: Transaction Proof Verification
	fmt.Println("\n4. Transaction Proof Verification")

	// Create some transaction data
	transactions := [][]byte{
		[]byte("tx1: transfer 100 coins"),
		[]byte("tx2: create account"),
		[]byte("tx3: update metadata"),
	}

	// Hash transactions
	var txHashes []string
	for _, tx := range transactions {
		txHashes = append(txHashes, decubcrypto.ComputeLeafHash(tx))
	}

	// Build Merkle tree for transactions
	txRootHash, err := decubcrypto.BuildMerkleTree(txHashes)
	if err != nil {
		log.Fatalf("Failed to build transaction Merkle tree: %v", err)
	}

	// Generate proof for first transaction
	txProof, err := decubcrypto.GenerateMerkleProof(txHashes, 0)
	if err != nil {
		log.Fatalf("Failed to generate transaction proof: %v", err)
	}

	// Create transaction proof
	transactionProof := decubcrypto.TransactionProof{
		MerkleProof: *txProof,
		TxID:        "tx1",
		TxData:      transactions[0],
	}

	// Verify transaction proof
	err = decubcrypto.VerifyTransactionProof(transactionProof, txRootHash)
	if err != nil {
		log.Fatalf("Transaction proof verification failed: %v", err)
	}
	fmt.Printf("Transaction proof verified successfully for transaction: %s\n", transactionProof.TxID)

	fmt.Println("\nAll examples completed successfully!")
}
