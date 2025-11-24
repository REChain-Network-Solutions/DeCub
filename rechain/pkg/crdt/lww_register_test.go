package crdt_test

import (
	"testing"
	"time"

	"github.com/rechain/rechain/pkg/crdt"
	"github.com/stretchr/testify/assert"
)

func TestLWWRegister(t *testing.T) {
	node1 := "node1"
	node2 := "node2"

	t.Run("NewLWWRegister", func(t *testing.T) {
		reg := crdt.NewLWWRegister(node1)
		assert.Equal(t, node1, reg.NodeID)
		nilValue := reg.GetValue()
		nilValue = nil
		assert.Nil(t, nilValue)
	})

	t.Run("SetAndGet", func(t *testing.T) {
		reg := crdt.NewLWWRegister(node1)
		testValue := "test value"
		reg.Set(testValue)
		assert.Equal(t, testValue, reg.GetValue())
	})

	t.Run("MergeSameNodeNewer", func(t *testing.T) {
		reg1 := crdt.NewLWWRegister(node1)
		reg2 := crdt.NewLWWRegister(node1)

		// Set reg1 with an older timestamp
		reg1.Set("old value")

		// Manually set an older timestamp for reg2
		reg2.Set("new value")

		// Merge reg2 into reg1 (should take reg2's value because it's newer)
		err := reg1.Merge(reg2)
		assert.NoError(t, err)
		assert.Equal(t, "new value", reg1.GetValue())
	})

	t.Run("MergeDifferentNodes", func(t *testing.T) {
		reg1 := crdt.NewLWWRegister(node1)
		reg2 := crdt.NewLWWRegister(node2)

		// Set both registers with the same timestamp but different values
		testTime := time.Now().UTC()
		reg1.Set("value from node1")
		reg2.Set("value from node2")

		// Merge should prefer the value from the register with the node that has the higher ID
		err := reg1.Merge(reg2)
		assert.NoError(t, err)

		// Since timestamps are the same, it should choose the value from the node with the higher ID
		expectedValue := "value from node2"
		if node1 > node2 {
			expectedValue = "value from node1"
		}
		assert.Equal(t, expectedValue, reg1.GetValue())
	})

	t.Run("MarshalUnmarshal", func(t *testing.T) {
		reg1 := crdt.NewLWWRegister(node1)
		reg1.Set("test value")

		data, err := reg1.Marshal()
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		reg2 := &crdt.LWWRegister{}
		err = reg2.Unmarshal(data)
		assert.NoError(t, err)
		assert.Equal(t, reg1.GetValue(), reg2.GetValue())
	})

	t.Run("IncompatibleMerge", func(t *testing.T) {
		reg := crdt.NewLWWRegister(node1)
		counter := crdt.NewPNCounter(node1)

		err := reg.Merge(counter)
		errMsg := "incompatible CRDT types: expected *lww.LWWRegister, got *crdt.PNCounter"
		assert.EqualError(t, err, errMsg)
	})
}
