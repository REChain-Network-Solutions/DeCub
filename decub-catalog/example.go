package main

import (
	"fmt"
	"time"
)

// Example demonstrates CRDT merge conflict resolution and delta sync
func ExampleCRDT() {
	fmt.Println("=== CRDT Catalog Example ===")

	// Create two nodes
	node1 := NewCRDTCatalog("node1")
	node2 := NewCRDTCatalog("node2")

	fmt.Println("\n--- Initial State ---")
	fmt.Printf("Node1 snapshots: %v\n", node1.QuerySnapshots(""))
	fmt.Printf("Node2 snapshots: %v\n", node2.QuerySnapshots(""))

	// Node1 adds a snapshot
	fmt.Println("\n--- Node1 adds snapshot ---")
	node1.AddSnapshot("snap1", map[string]interface{}{
		"size":     1024,
		"created":  time.Now().Format(time.RFC3339),
		"cluster":  "cluster-a",
	})

	fmt.Printf("Node1 snapshots: %v\n", node1.QuerySnapshots(""))

	// Simulate network delay, then Node2 adds different snapshot
	time.Sleep(10 * time.Millisecond)
	node2.AddSnapshot("snap2", map[string]interface{}{
		"size":     2048,
		"created":  time.Now().Format(time.RFC3339),
		"cluster":  "cluster-b",
	})

	fmt.Printf("Node2 snapshots: %v\n", node2.QuerySnapshots(""))

	// Both nodes update the same snapshot metadata (conflict scenario)
	fmt.Println("\n--- Concurrent metadata updates (conflict) ---")
	node1.UpdateSnapshotMetadata("snap1", map[string]interface{}{
		"size":     1024,
		"status":   "completed",
		"updated":  time.Now().Format(time.RFC3339),
	})

	time.Sleep(5 * time.Millisecond)
	node2.UpdateSnapshotMetadata("snap1", map[string]interface{}{
		"size":     1024,
		"status":   "failed",
		"updated":  time.Now().Format(time.RFC3339),
	})

	fmt.Println("Node1 metadata:", node1.snapshotMetadata["snap1"].Get())
	fmt.Println("Node2 metadata:", node2.snapshotMetadata["snap1"].Get())

	// Delta exchange simulation
	fmt.Println("\n--- Delta Exchange ---")

	// Node1 sends deltas to Node2
	deltas1 := node1.GenerateDelta()
	fmt.Printf("Node1 sending %d deltas to Node2\n", len(deltas1))

	for _, delta := range deltas1 {
		fmt.Printf("Applying delta: %s (%s)\n", delta.Key, delta.Type)
		node2.ApplyDelta(delta)
	}

	// Node2 sends deltas to Node1
	deltas2 := node2.GenerateDelta()
	fmt.Printf("Node2 sending %d deltas to Node1\n", len(deltas2))

	for _, delta := range deltas2 {
		fmt.Printf("Applying delta: %s (%s)\n", delta.Key, delta.Type)
		node1.ApplyDelta(delta)
	}

	fmt.Println("\n--- After Sync ---")
	fmt.Printf("Node1 snapshots: %v\n", node1.QuerySnapshots(""))
	fmt.Printf("Node2 snapshots: %v\n", node2.QuerySnapshots(""))

	fmt.Println("Node1 snap1 metadata:", node1.snapshotMetadata["snap1"].Get())
	fmt.Println("Node2 snap1 metadata:", node2.snapshotMetadata["snap1"].Get())

	// Demonstrate LWW conflict resolution
	fmt.Println("\n--- LWW Conflict Resolution ---")
	fmt.Println("Both nodes have metadata for snap1, but with different timestamps.")
	fmt.Println("The node with the later timestamp wins (LWW semantics).")
	fmt.Printf("Winner: %v\n", node1.snapshotMetadata["snap1"].Get())

	// Vector clock demonstration
	fmt.Println("\n--- Vector Clock Causality ---")
	fmt.Printf("Node1 VC: %v\n", node1.vectorClock)
	fmt.Printf("Node2 VC: %v\n", node2.vectorClock)

	// Demonstrate causal ordering
	oldDelta := &Delta{
		NodeID:      "node1",
		VectorClock: VectorClock{"node1": 1}, // Old clock
		Type:        "orset",
		Key:         "snapshots:snap3",
		Data:        map[string]interface{}{"tag": "old-tag"},
		Timestamp:   time.Now().UnixNano(),
	}

	fmt.Println("\n--- Causal Ordering ---")
	fmt.Printf("Trying to apply old delta with VC %v\n", oldDelta.VectorClock)
	applied := node2.ApplyDelta(oldDelta)
	fmt.Printf("Applied: %v (should be false due to causality)\n", applied)

	fmt.Println("\n=== Example Complete ===")
}

// RunExample runs the CRDT example
func RunExample() {
	ExampleCRDT()
}
