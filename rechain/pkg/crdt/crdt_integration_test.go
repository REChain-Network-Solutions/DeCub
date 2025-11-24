package crdt_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/rechain/rechain/pkg/crdt"
	"github.com/rechain/rechain/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCRDTs_Integration(t *testing.T) {
	t.Run("LWWRegister_Convergence", func(t *testing.T) {
		node1 := "node-1"
		node2 := "node-2"

		// Create two replicas of the same register on different nodes
		r1 := crdt.NewLWWRegister(node1)
		r2 := crdt.NewLWWRegister(node2)

		// Node 1 sets a value
		r1.Set("value-from-node1")

		// Node 2 sets a different value
		r2.Set("value-from-node2")

		// Merge r2 into r1 (should take r2's value because it's newer)
		err := r1.Merge(r2)
		require.NoError(t, err)

		// Merge r1 into r2 (should be idempotent)
		err = r2.Merge(r1)
		require.NoError(t, err)

		// Both registers should now have the same value
		assert.Equal(t, r1.GetValue(), r2.GetValue())

		// The final value should be from node2 because it was set last
		// (assuming the timestamps are different, which they should be in real usage)
		// In this test, since we're not controlling the timestamps precisely,
		// we'll just verify that they converged to the same value
	})

	t.Run("PNCounter_Convergence", func(t *testing.T) {
		node1 := "node-1"
		node2 := "node-2"

		// Create two replicas of the same counter on different nodes
		c1 := crdt.NewPNCounter(node1)
		c2 := crdt.NewPNCounter(node2)

		// Node 1 increments by 5
		c1.Increment(5)

		// Node 2 increments by 3 and decrements by 1
		c2.Increment(3)
		c2.Decrement(1)

		// Merge c2 into c1
		err := c1.Merge(c2)
		require.NoError(t, err)

		// Merge c1 into c2 (should be idempotent)
		err = c2.Merge(c1)
		require.NoError(t, err)

		// Both counters should now have the same value
		assert.Equal(t, c1.Value(), c2.Value())
		assert.Equal(t, int64(7), c1.Value()) // 5 + 3 - 1 = 7
	})

	t.Run("ORSet_Convergence", func(t *testing.T) {
		node1 := "node-1"
		node2 := "node-2"

		// Create two replicas of the same set on different nodes
		s1 := crdt.NewORSet(node1)
		s2 := crdt.NewORSet(node2)

		// Node 1 adds "a" and "b"
		s1.Add("a")
		s1.Add("b")

		// Node 2 adds "b" and "c"
		s2.Add("b")
		s2.Add("c")

		// Node 2 removes "b"
		s2.Remove("b")

		// Merge s2 into s1
		err := s1.Merge(s2)
		require.NoError(t, err)

		// Merge s1 into s2 (should be idempotent)
		err = s2.Merge(s1)
		require.NoError(t, err)

		// Both sets should now have the same elements
		elements1 := s1.Elements()
		elements2 := s2.Elements()

		// Convert to a map for easier comparison
		elementsMap1 := make(map[interface{}]bool)
		for _, e := range elements1 {
			elementsMap1[e] = true
		}

		elementsMap2 := make(map[interface{}]bool)
		for _, e := range elements2 {
			elementsMap2[e] = true
		}

		assert.Equal(t, elementsMap1, elementsMap2)

		// The final set should contain "a" and "c" ("b" was removed by node2)
		assert.True(t, s1.Contains("a") && s1.Contains("c") && !s1.Contains("b"),
			"Expected {a, c}, got %v", elements1)
	})

	t.Run("GCounter_Convergence", func(t *testing.T) {
		node1 := "node-1"
		node2 := "node-2"

		// Create two replicas of the same counter on different nodes
		c1 := crdt.NewGCounter(node1)
		c2 := crdt.NewGCounter(node2)

		// Node 1 increments by 5
		c1.Increment(5)

		// Node 2 increments by 3
		c2.Increment(3)

		// Merge c2 into c1
		err := c1.Merge(c2)
		require.NoError(t, err)

		// Merge c1 into c2 (should be idempotent)
		err = c2.Merge(c1)
		require.NoError(t, err)

		// Both counters should now have the same value
		assert.Equal(t, c1.Value(), c2.Value())
		assert.Equal(t, int64(8), c1.Value()) // 5 + 3 = 8
	})

	t.Run("CRDT_Serialization_Roundtrip", func(t *testing.T) {
		tests := []struct {
			name  string
			crdt  crdt.CRDT
			setup func(c crdt.CRDT)
		}{
			{
				name: "LWWRegister",
				crdt: crdt.NewLWWRegister("test-node"),
				setup: func(c crdt.CRDT) {
					c.(*crdt.LWWRegister).Set("test-value")
				},
			},
			{
				name: "PNCounter",
				crdt: crdt.NewPNCounter("test-node"),
				setup: func(c crdt.CRDT) {
					c.(*crdt.PNCounter).Increment(5)
					c.(*crdt.PNCounter).Decrement(2)
				},
			},
			{
				name: "ORSet",
				crdt: crdt.NewORSet("test-node"),
				setup: func(c crdt.CRDT) {
					c.(*crdt.ORSet).Add("a")
					c.(*crdt.ORSet).Add("b")
					c.(*crdt.ORSet).Remove("a")
				},
			},
			{
				name: "GCounter",
				crdt: crdt.NewGCounter("test-node"),
				setup: func(c crdt.CRDT) {
					c.(*crdt.GCounter).Increment(10)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Setup the CRDT
				tt.setup(tt.crdt)

				// Marshal to JSON
				data, err := tt.crdt.Marshal()
				require.NoError(t, err)

				// Create a new CRDT of the same type
				newCRDT, err := crdt.New(tt.crdt.Type(), "new-node")
				require.NoError(t, err)

				// Unmarshal the data
				err = newCRDT.Unmarshal(data)
				require.NoError(t, err)

				// The new CRDT should be equal to the original
				assert.Equal(t, tt.crdt.Value(), newCRDT.Value())

				// Test JSON marshaling/unmarshaling
				jsonData, err := json.Marshal(tt.crdt)
				require.NoError(t, err)

				var jsonMap map[string]interface{}
				err = json.Unmarshal(jsonData, &jsonMap)
				require.NoError(t, err)

				// The JSON should contain the type field
				assert.Equal(t, string(tt.crdt.Type()), jsonMap["type"])
			})
		}
	})

	t.Run("CRDT_Network_Partition", func(t *testing.T) {
		// This test simulates a network partition where two nodes operate independently
		// and then merge their states after the partition is resolved.

		node1 := "node-1"
		node2 := "node-2"

		// Create two replicas of the same counter on different nodes
		c1 := crdt.NewPNCounter(node1)
		c2 := crdt.NewPNCounter(node2)

		// Both nodes start with the same state
		c1.Increment(5)
		c2.Increment(5)

		// Network partition occurs - nodes operate independently

		// Node 1 operations during partition
		c1.Increment(3)
		c1.Decrement(1)

		// Node 2 operations during partition
		c2.Increment(2)
		c2.Decrement(1)

		// Network partition is resolved - merge states
		err := c1.Merge(c2)
		require.NoError(t, err)

		err = c2.Merge(c1)
		require.NoError(t, err)

		// Both nodes should have the same state
		assert.Equal(t, c1.Value(), c2.Value())

		// The final value should be:
		// Initial: 5 (both nodes)
		// During partition:
		// - Node 1: +3 -1 = +2
		// - Node 2: +2 -1 = +1
		// Total: 5 + 2 + 1 = 8
		assert.Equal(t, int64(8), c1.Value())
	})
}
