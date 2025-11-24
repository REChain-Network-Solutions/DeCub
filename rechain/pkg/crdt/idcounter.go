package crdt

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// IDCounter is an operation-based increment-decrement counter (CvRDT)
type IDCounter struct {
	nodeID string
	// Using RWMutex for better read performance
	mu sync.RWMutex
	// Using sync.Map for better concurrent read performance
	p sync.Map // map[string]int64 // Positive increments
	n sync.Map // map[string]int64 // Negative increments (decrements)
	// Cache for Value() to avoid recalculating
	valueCache     int64
	valueCacheTime int64 // UnixNano timestamp of last cache update
}

// NewIDCounter creates a new IDCounter
func NewIDCounter(nodeID string) *IDCounter {
	return &IDCounter{
		nodeID:         nodeID,
		p:              sync.Map{},
		n:              sync.Map{},
		valueCache:     0,
		valueCacheTime: 0,
	}
}

// Type returns the CRDT type
func (c *IDCounter) Type() CRDTType {
	return "idcounter"
}

// Increment increments the counter by the given value (must be positive)
func (c *IDCounter) Increment(by int64) {
	if by <= 0 {
		return
	}

	// Fast path: update the value without lock if possible
	if val, ok := c.p.Load(c.nodeID); ok {
		c.p.Store(c.nodeID, val.(int64)+by)
	} else {
		c.p.Store(c.nodeID, by)
	}

	// Invalidate cache
	c.invalidateCache()
}

// Decrement decrements the counter by the given value (must be positive)
func (c *IDCounter) Decrement(by int64) {
	if by <= 0 {
		return
	}

	// Fast path: update the value without lock if possible
	if val, ok := c.n.Load(c.nodeID); ok {
		c.n.Store(c.nodeID, val.(int64)+by)
	} else {
		c.n.Store(c.nodeID, by)
	}

	// Invalidate cache
	c.invalidateCache()
}

// Value returns the current value of the counter
// This method uses a cached result if available and still valid
func (c *IDCounter) Value() int64 {
	// Try to use cache first (read-lock only)
	if val := c.getCachedValue(); val != 0 || c.valueCacheTime > 0 {
		return val
	}

	// Cache miss, need to recalculate
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check cache again after acquiring lock (double-checked locking)
	if c.valueCacheTime > 0 {
		return c.valueCache
	}

	// Recalculate the value
	var sumP, sumN int64

	c.p.Range(func(key, value interface{}) bool {
		sumP += value.(int64)
		return true
	})

	c.n.Range(func(key, value interface{}) bool {
		sumN += value.(int64)
		return true
	})

	// Update cache
	c.valueCache = sumP - sumN
	c.valueCacheTime = time.Now().UnixNano()

	return c.valueCache
}

// getCachedValue returns the cached value if it's still valid
func (c *IDCounter) getCachedValue() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Cache is valid for 100ms
	if c.valueCacheTime > 0 && time.Since(time.Unix(0, c.valueCacheTime)) < 100*time.Millisecond {
		return c.valueCache
	}
	return 0
}

// invalidateCache marks the cache as invalid
func (c *IDCounter) invalidateCache() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.valueCacheTime = 0
}

// Merge merges another IDCounter into this one
func (c *IDCounter) Merge(other CRDT) error {
	otherCounter, ok := other.(*IDCounter)
	if !ok {
		return fmt.Errorf("%w: expected *IDCounter, got %T", ErrIncompatibleTypes, other)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Invalidate cache
	c.invalidateCache()

	// Merge positive counters
	otherCounter.p.Range(func(key, value interface{}) bool {
		nodeID := key.(string)
		count := value.(int64)
		
		if val, ok := c.p.Load(nodeID); !ok || count > val.(int64) {
			c.p.Store(nodeID, count)
		}
		return true
	})

	// Merge negative counters
	otherCounter.n.Range(func(key, value interface{}) bool {
		nodeID := key.(string)
		count := value.(int64)
		
		if val, ok := c.n.Load(nodeID); !ok || count > val.(int64) {
			c.n.Store(nodeID, count)
		}
		return true
	})

	return nil
}

// Value implements the CRDT interface
func (c *IDCounter) Value() interface{} {
	return c.Value()
}

// Marshal serializes the IDCounter to JSON
func (c *IDCounter) Marshal() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Convert sync.Maps to regular maps for serialization
	p := make(map[string]int64)
	c.p.Range(func(key, value interface{}) bool {
		p[key.(string)] = value.(int64)
		return true
	})

	n := make(map[string]int64)
	c.n.Range(func(key, value interface{}) bool {
		n[key.(string)] = value.(int64)
		return true
	})

	// Create a serializable struct
	data := struct {
		Type string            `json:"type"`
		P    map[string]int64 `json:"p"`
		N    map[string]int64 `json:"n"`
	}{
		Type: string(c.Type()),
		P:    p,
		N:    n,
	}

	return json.Marshal(data)
}

// Unmarshal deserializes the IDCounter from JSON
func (c *IDCounter) Unmarshal(data []byte) error {
	var aux struct {
		Type string            `json:"type"`
		P    map[string]int64 `json:"p"`
		N    map[string]int64 `json:"n"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Type != string(c.Type()) {
		return fmt.Errorf("%w: expected %s, got %s", ErrIncompatibleTypes, c.Type(), aux.Type)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Reset the maps
	c.p = sync.Map{}
	c.n = sync.Map{}

	// Copy the maps
	for k, v := range aux.P {
		c.p.Store(k, v)
	}

	for k, v := range aux.N {
		c.n.Store(k, v)
	}

	// Reset cache
	c.invalidateCache()

	return nil
}

// ApplyOperation applies an operation to the counter
func (c *IDCounter) ApplyOperation(op Operation) error {
	switch op.Type {
	case "increment":
		if value, ok := op.Value.(float64); ok {
			c.Increment(int64(value))
			return nil
		}
		return fmt.Errorf("invalid value type for increment: %T", op.Value)
	case "decrement":
		if value, ok := op.Value.(float64); ok {
			c.Decrement(int64(value))
			return nil
		}
		return fmt.Errorf("invalid value type for decrement: %T", op.Value)
	default:
		return fmt.Errorf("unknown operation type: %s", op.Type)
	}
}

// Operation represents an operation on the counter
type Operation struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}
