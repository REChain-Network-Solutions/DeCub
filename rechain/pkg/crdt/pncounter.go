package crdt

import (
	"encoding/json"
	"fmt"
	"sync"
)

// PNCounter is a Positive-Negative Counter CRDT
type PNCounter struct {
	nodeID string
	mu     sync.RWMutex
	P      map[string]int64 `json:"p"` // Positive increments
	N      map[string]int64 `json:"n"` // Negative increments (decrements)
}

// NewPNCounter creates a new PNCounter
func NewPNCounter(nodeID string) *PNCounter {
	return &PNCounter{
		nodeID: nodeID,
		P:      make(map[string]int64),
		N:      make(map[string]int64),
	}
}

// Type returns the CRDT type
func (c *PNCounter) Type() CRDTType {
	return PNCounter
}

// Increment increments the counter by the given value (must be positive)
func (c *PNCounter) Increment(by int64) {
	if by <= 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.P[c.nodeID] += by
}

// Decrement decrements the counter by the given value (must be positive)
func (c *PNCounter) Decrement(by int64) {
	if by <= 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.N[c.nodeID] += by
}

// Value returns the current value of the counter
func (c *PNCounter) Value() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var sumP, sumN int64

	for _, v := range c.P {
		sumP += v
	}

	for _, v := range c.N {
		sumN += v
	}

	return sumP - sumN
}

// Merge merges another PNCounter
func (c *PNCounter) Merge(other CRDT) error {
	otherCounter, ok := other.(*PNCounter)
	if !ok {
		return fmt.Errorf("%w: expected PNCounter, got %T", ErrIncompatibleTypes, other)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Merge positive counters
	for nodeID, value := range otherCounter.P {
		if value > c.P[nodeID] {
			c.P[nodeID] = value
		}
	}

	// Merge negative counters
	for nodeID, value := range otherCounter.N {
		if value > c.N[nodeID] {
			c.N[nodeID] = value
		}
	}

	return nil
}

// Marshal serializes the PNCounter to JSON
func (c *PNCounter) Marshal() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return json.Marshal(struct {
		Type CRDTType          `json:"type"`
		P    map[string]int64 `json:"p"`
		N    map[string]int64 `json:"n"`
	}{
		Type: PNCounter,
		P:    c.P,
		N:    c.N,
	})
}

// Unmarshal deserializes the PNCounter from JSON
func (c *PNCounter) Unmarshal(data []byte) error {
	var aux struct {
		Type CRDTType          `json:"type"`
		P    map[string]int64 `json:"p"`
		N    map[string]int64 `json:"n"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Type != PNCounter {
		return fmt.Errorf("%w: expected PNCounter, got %s", ErrIncompatibleTypes, aux.Type)
	}

	c.P = aux.P
	c.N = aux.N

	return nil
}
