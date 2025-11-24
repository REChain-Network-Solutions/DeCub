package crdt_test

import (
	"testing"

	"github.com/rechain/rechain/pkg/crdt"
	"github.com/stretchr/testify/assert"
)

func TestIDCounter(t *testing.T) {
	node1 := "node1"
	node2 := "node2"

	t.Run("NewIDCounter", func(t *testing.T) {
		counter := crdt.NewIDCounter(node1)
		assert.Equal(t, int64(0), counter.Value())
	})

	t.Run("Increment", func(t *testing.T) {
		counter := crdt.NewIDCounter(node1)
		counter.Increment(5)
		assert.Equal(t, int64(5), counter.Value())

		// Should ignore non-positive values
		counter.Increment(0)
		counter.Increment(-1)
		assert.Equal(t, int64(5), counter.Value())
	})

	t.Run("Decrement", func(t *testing.T) {
		counter := crdt.NewIDCounter(node1)
		counter.Increment(10)
		counter.Decrement(3)
		assert.Equal(t, int64(7), counter.Value())

		// Should ignore non-positive values
		counter.Decrement(0)
		counter.Decrement(-1)
		assert.Equal(t, int64(7), counter.Value())
	})

	t.Run("Merge", func(t *testing.T) {
		counter1 := crdt.NewIDCounter(node1)
		counter2 := crdt.NewIDCounter(node2)

		// Counter 1: +5 -2 = 3
		counter1.Increment(5)
		counter1.Decrement(2)

		// Counter 2: +3 -1 = 2
		counter2.Increment(3)
		counter2.Decrement(1)

		// Merge counter2 into counter1
		err := counter1.Merge(counter2)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), counter1.Value()) // 5 - 2 + 3 - 1 = 5

		// Merge counter1 into counter2 (should be idempotent)
		err = counter2.Merge(counter1)
		assert.NoError(t, err)
		assert.Equal(t, counter1.Value(), counter2.Value())
	})

	t.Run("ApplyOperation", func(t *testing.T) {
		counter := crdt.NewIDCounter(node1)

		// Test increment operation
		incOp := crdt.Operation{Type: "increment", Value: float64(5)}
		err := counter.ApplyOperation(incOp)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), counter.Value())

		// Test decrement operation
		decOp := crdt.Operation{Type: "decrement", Value: float64(2)}
		err = counter.ApplyOperation(decOp)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), counter.Value())

		// Test invalid operation type
		invalidOp := crdt.Operation{Type: "invalid", Value: float64(1)}
		err = counter.ApplyOperation(invalidOp)
		assert.Error(t, err)
		assert.Equal(t, "unknown operation type: invalid", err.Error())

		// Test invalid value type
		invalidValueOp := crdt.Operation{Type: "increment", Value: "not-a-number"}
		err = counter.ApplyOperation(invalidValueOp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid value type for increment")
	})

	t.Run("MarshalUnmarshal", func(t *testing.T) {
		counter1 := crdt.NewIDCounter(node1)
		counter1.Increment(5)
		counter1.Decrement(2)

		data, err := counter1.Marshal()
		assert.NoError(t, err)

		counter2 := &crdt.IDCounter{}
		err = counter2.Unmarshal(data)
		assert.NoError(t, err)

		assert.Equal(t, counter1.Value(), counter2.Value())
	})

	t.Run("IncompatibleMerge", func(t *testing.T) {
		counter := crdt.NewIDCounter(node1)
		set := crdt.NewORSet(node1)

		err := counter.Merge(set)
		errMsg := "incompatible CRDT types: expected *crdt.IDCounter, got *crdt.ORSet"
		assert.EqualError(t, err, errMsg)
	})
}
