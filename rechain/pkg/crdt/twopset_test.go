package crdt_test

import (
	"testing"

	"github.com/rechain/rechain/pkg/crdt"
	"github.com/stretchr/testify/assert"
)

func TestTwoPhaseSet(t *testing.T) {
	node1 := "node1"
	node2 := "node2"

	t.Run("NewTwoPhaseSet", func(t *testing.T) {
		set := crdt.NewTwoPhaseSet(node1)
		assert.Equal(t, 0, len(set.Elements()))
	})

	t.Run("AddAndContains", func(t *testing.T) {
		set := crdt.NewTwoPhaseSet(node1)
		set.Add("a")
		assert.True(t, set.Contains("a"))
		assert.False(t, set.Contains("b"))
	})

	t.Run("Remove", func(t *testing.T) {
		set := crdt.NewTwoPhaseSet(node1)
		set.Add("a")
		set.Remove("a")
		assert.False(t, set.Contains("a"))
	})

	t.Run("ReAddAfterRemove", func(t *testing.T) {
		set := crdt.NewTwoPhaseSet(node1)
		set.Add("a")
		set.Remove("a")
		set.Add("a") // Should not be added back
		assert.False(t, set.Contains("a"), "Element should not be re-added after removal")
	})

	t.Run("Merge", func(t *testing.T) {
		set1 := crdt.NewTwoPhaseSet(node1)
		set2 := crdt.NewTwoPhaseSet(node2)

		// Set 1: add "a" and "b"
		set1.Add("a")
		set1.Add("b")

		// Set 2: add "b" and "c", then remove "b"
		set2.Add("b")
		set2.Add("c")
		set2.Remove("b")

		// Merge set2 into set1
		err := set1.Merge(set2)
		assert.NoError(t, err)

		// Check elements
		elements := set1.Elements()
		assert.Len(t, elements, 2)
		assert.Contains(t, elements, "a")
		assert.Contains(t, elements, "c")
		assert.NotContains(t, elements, "b")

		// Merge set1 into set2 (should be idempotent)
		err = set2.Merge(set1)
		assert.NoError(t, err)
		assert.ElementsMatch(t, set1.Elements(), set2.Elements())
	})

	t.Run("MarshalUnmarshal", func(t *testing.T) {
		set1 := crdt.NewTwoPhaseSet(node1)
		set1.Add("a")
		set1.Add("b")
		set1.Remove("a")

		data, err := set1.Marshal()
		assert.NoError(t, err)

		set2 := &crdt.TwoPhaseSet{}
		err = set2.Unmarshal(data)
		assert.NoError(t, err)

		assert.ElementsMatch(t, set1.Elements(), set2.Elements())
	})

	t.Run("IncompatibleMerge", func(t *testing.T) {
		set := crdt.NewTwoPhaseSet(node1)
		counter := crdt.NewPNCounter(node1)

		err := set.Merge(counter)
		errMsg := "incompatible CRDT types: expected *crdt.TwoPhaseSet, got *crdt.PNCounter"
		assert.EqualError(t, err, errMsg)
	})
}
