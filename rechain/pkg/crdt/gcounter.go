package crdt

import (
	"encoding/json"
	"fmt"
	"sync"
)

// GCounter is a Grow-only Counter CRDT
type GCounter struct {
	nodeID string
	mu     sync.RWMutex
	counts map[string]int64 // nodeID -> count
}

// NewGCounter creates a new GCounter
func NewGCounter(nodeID string) *GCounter {
	return &GCounter{
		nodeID: nodeID,
		counts: make(map[string]int64),
	}
}

// Type returns the CRDT type
func (c *GCounter) Type() CRDTType {
	return "gcounter"
}

// Increment increments the counter by the given value (must be positive)
func (c *GCounter) Increment(by int64) {
	if by <= 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.counts[c.nodeID] += by
}

// Value returns the current value of the counter
func (c *GCounter) Value() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var total int64
	for _, count := range c.counts {
		total += count
	}

	return total
}

// Merge merges another GCounter into this one
func (c *GCounter) Merge(other CRDT) error {
	otherCounter, ok := other.(*GCounter)
	if !ok {
		return fmt.Errorf("%w: expected *GCounter, got %T", ErrIncompatibleTypes, other)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	otherCounter.mu.RLock()
	defer otherCounter.mu.RUnlock()

	// Take the maximum count for each node
	for nodeID, count := range otherCounter.counts {
		if count > c.counts[nodeID] {
			c.counts[nodeID] = count
		}
	}

	return nil
}

// Value implements the CRDT interface
func (c *GCounter) Value() interface{} {
	return c.Value()
}

// Marshal serializes the GCounter to JSON
func (c *GCounter) Marshal() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data := struct {
		Type   string          `json:"type"`
		Counts map[string]int64 `json:"counts"`
	}{
		Type:   string(c.Type()),
		Counts: c.counts,
	}

	return json.Marshal(data)
}

// Unmarshal deserializes the GCounter from JSON
func (c *GCounter) Unmarshal(data []byte) error {
	var aux struct {
		Type   string          `json:"type"`
		Counts map[string]int64 `json:"counts"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Type != string(c.Type()) {
		return fmt.Errorf("%w: expected %s, got %s", ErrIncompatibleTypes, c.Type(), aux.Type)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.counts = aux.Counts

	return nil
}
