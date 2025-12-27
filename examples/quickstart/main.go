package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/REChain-Network-Solutions/DeCub/rechain/pkg/crdt"
)

func main() {
	fmt.Println("DeCube Quick Start Example")
	fmt.Println("=========================")

	// Example 1: CRDT Operations
	fmt.Println("\n1. CRDT Operations Example")
	demoCRDT()

	// Example 2: Merkle Tree
	fmt.Println("\n2. Merkle Tree Example")
	demoMerkleTree()

	// Example 3: Basic Operations
	fmt.Println("\n3. Basic Operations Example")
	demoBasicOperations()
}

func demoCRDT() {
	// Create a new OR-Set CRDT
	orset := crdt.NewORSet()

	// Add elements
	orset.Add("item1")
	orset.Add("item2")
	orset.Add("item3")

	// Check membership
	fmt.Printf("  Contains 'item1': %v\n", orset.Contains("item1"))
	fmt.Printf("  Contains 'item4': %v\n", orset.Contains("item4"))

	// Remove an element
	orset.Remove("item2")
	fmt.Printf("  Contains 'item2' after removal: %v\n", orset.Contains("item2"))

	// Get all elements
	elements := orset.Elements()
	fmt.Printf("  All elements: %v\n", elements)
}

func demoMerkleTree() {
	// This is a placeholder - implement actual Merkle tree demo
	// when the merkle package is available
	fmt.Println("  Merkle tree operations would be demonstrated here")
	fmt.Println("  See rechain/pkg/merkle for implementation")
}

func demoBasicOperations() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("  Basic operations example:")
	fmt.Println("  - Connect to DeCube cluster")
	fmt.Println("  - Create snapshot")
	fmt.Println("  - Query catalog")
	fmt.Println("  - Sync with gossip protocol")

	// Simulate async operation
	select {
	case <-time.After(1 * time.Second):
		fmt.Println("  ✓ Operations completed")
	case <-ctx.Done():
		log.Fatal("  ✗ Operation timeout")
	}
}

