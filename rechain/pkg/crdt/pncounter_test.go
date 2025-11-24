package crdt_test

import (
	"testing"

	"github.com/rechain/rechain/pkg/crdt"
	"github.com/stretchr/testify/assert"
)

func TestPNCounter(t *testing.T) {
	node1 := "node1"
	node2 := "node2"

	t.Run("NewPNCounter", func(t *testing.T) {
		counter := crdt.NewPNCounter(node1)
		assert.Equal(t, int64(0), counter.Value())
	})

	t.Run("Increment", func(t *testing.T) {
		counter := crdt.NewPNCounter(node1)
		counter.Increment(5)
		assert.Equal(t, int64(5), counter.Value())

		// Should not allow negative increments
		counter.Increment(-3)
		assert.Equal(t, int64(5), counter.Value())
	})

	t.Run("Decrement", func(t *testing.T) {
		counter := crdt.NewPNCounter(node1)
		counter.Increment(10)
		counter.Decrement(3)
		assert.Equal(t, int64(7), counter.Value())

		// Should not allow negative decrements
		counter.Decrement(-2)
		assert.Equal(t, int64(7), counter.Value())
	})

	t.Run("Merge", func(t *testing.T) {
		// Create two counters from different nodes
		counter1 := crdt.NewPNCounter(node1)
		counter2 := crdt.NewPNCounter(node2)

		// Node1 increments by 5
		counter1.Increment(5)

		// Node2 increments by 3 and decrements by 1
		counter2.Increment(3)
		counter2.Decrement(1)

		// Merge node2 into node1
		err := counter1.Merge(counter2)
		assert.NoError(t, err)

		// Should have 5 (node1) + 3 (node2) - 1 (node2) = 7
		assert.Equal(t, int64(7), counter1.Value())

		// Merge node1 into node2
		err = counter2.Merge(counter1)
		assert.NoError(t, err)

		// Both should have the same value after bidirectional merge
		assert.Equal(t, counter1.Value(), counter2.Value())
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		counter := crdt.NewPNCounter(node1)

		// Simulate concurrent increments and decrements
		for i := 0; i < 100; i++ {
			go func() {
				counter.Increment(1)
			}()
		}

		// The exact value isn't deterministic due to concurrency,
		// but we can check if it's within expected bounds
		// Note: In practice, you'd use proper synchronization for testing concurrency
	})

	t.Run("MarshalUnmarshal", func(t *testing.T) {
		counter1 := crdt.NewPNCounter(node1)
		counter1.Increment(5)
		counter1.Decrement(2)

		data, err := counter1.Marshal()
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		counter2 := &crdt.PNCounter{}
		err = counter2.Unmarshal(data)
		assert.NoError(t, err)
		assert.Equal(t, counter1.Value(), counter2.Value())
	})

	t.Run("IncompatibleMerge", func(t *testing.T) {
		counter := crdt.NewPNCounter(node1)
		reg := crdt.NewLWWRegister(node1)

		err := counter.Merge(reg)
		errMsg := "incompatible CRDT types: expected *crdt.PNCounter, got *crdt.LWWRegister"
		assert.EqualError(t, err, errMsg)
	})
}
