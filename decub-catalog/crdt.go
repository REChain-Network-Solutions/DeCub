package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// VectorClock represents a vector clock for causal ordering
type VectorClock map[string]int64

// NewVectorClock creates a new vector clock
func NewVectorClock() VectorClock {
	return make(VectorClock)
}

// Increment increments the clock for a node
func (vc VectorClock) Increment(nodeID string) {
	vc[nodeID]++
}

// Merge merges another vector clock
func (vc VectorClock) Merge(other VectorClock) {
	for node, time := range other {
		if time > vc[node] {
			vc[node] = time
		}
	}
}

// Compare compares two vector clocks
// Returns: -1 if vc < other, 0 if concurrent, 1 if vc > other
func (vc VectorClock) Compare(other VectorClock) int {
	vcLess := false
	otherLess := false

	for node := range vc {
		if vc[node] < other[node] {
			vcLess = true
		} else if vc[node] > other[node] {
			otherLess = true
		}
	}

	for node := range other {
		if vc[node] < other[node] {
			vcLess = true
		} else if vc[node] > other[node] {
			otherLess = true
		}
	}

	if vcLess && otherLess {
		return 0 // concurrent
	} else if vcLess {
		return -1 // vc < other
	} else if otherLess {
		return 1 // vc > other
	}
	return 0 // equal
}

// LWWRegister represents a Last-Write-Wins Register CRDT
type LWWRegister struct {
	value     interface{}
	timestamp int64
	nodeID    string
	mu        sync.RWMutex
}

// NewLWWRegister creates a new LWW register
func NewLWWRegister(nodeID string) *LWWRegister {
	return &LWWRegister{
		nodeID: nodeID,
	}
}

// Set sets the value with current timestamp
func (r *LWWRegister) Set(value interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.value = value
	r.timestamp = time.Now().UnixNano()
}

// Merge merges another LWW register
func (r *LWWRegister) Merge(other *LWWRegister) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if other.timestamp > r.timestamp {
		r.value = other.value
		r.timestamp = other.timestamp
		r.nodeID = other.nodeID
	}
}

// Get returns the current value
func (r *LWWRegister) Get() interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.value
}

// Delta represents a CRDT delta for gossip
type Delta struct {
	NodeID      string                 `json:"node_id"`
	VectorClock VectorClock            `json:"vector_clock"`
	Type        string                 `json:"type"` // "orset" or "lww"
	Key         string                 `json:"key"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   int64                  `json:"timestamp"`
}

// NewDelta creates a new delta
func NewDelta(nodeID string, vc VectorClock, deltaType, key string, data map[string]interface{}) *Delta {
	return &Delta{
		NodeID:      nodeID,
		VectorClock: vc,
		Type:        deltaType,
		Key:         key,
		Data:        data,
		Timestamp:   time.Now().UnixNano(),
	}
}

// ORSetDelta represents changes to an OR-Set
type ORSetDelta struct {
	Adds []string `json:"adds"`
	Rms  []string `json:"rms"`
}

// LWWDelta represents changes to LWW registers
type LWWDelta struct {
	Updates map[string]interface{} `json:"updates"`
}

// CRDTCatalog represents the CRDT-backed catalog
type CRDTCatalog struct {
	nodeID      string
	vectorClock VectorClock

	// OR-Sets for sets
	snapshots *ORSet
	images    *ORSet

	// LWW Registers for metadata
	snapshotMetadata map[string]*LWWRegister // snapshotID -> metadata register
	imageMetadata    map[string]*LWWRegister // imageID -> metadata register

	// Pending deltas for gossip
	deltas []*Delta

	mu sync.RWMutex
}

// NewCRDTCatalog creates a new CRDT catalog
func NewCRDTCatalog(nodeID string) *CRDTCatalog {
	return &CRDTCatalog{
		nodeID:           nodeID,
		vectorClock:      NewVectorClock(),
		snapshots:        NewORSet(),
		images:           NewORSet(),
		snapshotMetadata: make(map[string]*LWWRegister),
		imageMetadata:    make(map[string]*LWWRegister),
		deltas:           make([]*Delta, 0),
	}
}

// AddSnapshot adds a snapshot with metadata
func (c *CRDTCatalog) AddSnapshot(snapshotID string, metadata map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Add to OR-Set
	tag := c.snapshots.Add(snapshotID)

	// Update metadata LWW register
	if c.snapshotMetadata[snapshotID] == nil {
		c.snapshotMetadata[snapshotID] = NewLWWRegister(c.nodeID)
	}
	c.snapshotMetadata[snapshotID].Set(metadata)

	// Update vector clock
	c.vectorClock.Increment(c.nodeID)

	// Create delta
	deltaData := map[string]interface{}{
		"tag":      tag,
		"metadata": metadata,
	}
	delta := NewDelta(c.nodeID, c.vectorClock, "orset", "snapshots:"+snapshotID, deltaData)
	c.deltas = append(c.deltas, delta)

	fmt.Printf("Added snapshot %s with tag %s\n", snapshotID, tag)
}

// RemoveSnapshot removes a snapshot
func (c *CRDTCatalog) RemoveSnapshot(snapshotID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.snapshots.Remove(snapshotID)

	// Update vector clock
	c.vectorClock.Increment(c.nodeID)

	// Create delta
	deltaData := map[string]interface{}{
		"removed": true,
	}
	delta := NewDelta(c.nodeID, c.vectorClock, "orset", "snapshots:"+snapshotID+":remove", deltaData)
	c.deltas = append(c.deltas, delta)

	fmt.Printf("Removed snapshot %s\n", snapshotID)
}

// UpdateSnapshotMetadata updates snapshot metadata
func (c *CRDTCatalog) UpdateSnapshotMetadata(snapshotID string, metadata map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.snapshotMetadata[snapshotID] == nil {
		c.snapshotMetadata[snapshotID] = NewLWWRegister(c.nodeID)
	}
	c.snapshotMetadata[snapshotID].Set(metadata)

	// Update vector clock
	c.vectorClock.Increment(c.nodeID)

	// Create delta
	delta := NewDelta(c.nodeID, c.vectorClock, "lww", "snapshot_metadata:"+snapshotID, metadata)
	c.deltas = append(c.deltas, delta)

	fmt.Printf("Updated metadata for snapshot %s\n", snapshotID)
}

// AddImage adds an image with metadata
func (c *CRDTCatalog) AddImage(imageID string, metadata map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tag := c.images.Add(imageID)

	if c.imageMetadata[imageID] == nil {
		c.imageMetadata[imageID] = NewLWWRegister(c.nodeID)
	}
	c.imageMetadata[imageID].Set(metadata)

	c.vectorClock.Increment(c.nodeID)

	deltaData := map[string]interface{}{
		"tag":      tag,
		"metadata": metadata,
	}
	delta := NewDelta(c.nodeID, c.vectorClock, "orset", "images:"+imageID, deltaData)
	c.deltas = append(c.deltas, delta)

	fmt.Printf("Added image %s with tag %s\n", imageID, tag)
}

// QuerySnapshots returns all snapshots with metadata
func (c *CRDTCatalog) QuerySnapshots(query string) []map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var results []map[string]interface{}

	// In a real implementation, this would iterate through the OR-Set
	// For demo, return some sample data
	if query == "" || query == "snap1" {
		results = append(results, map[string]interface{}{
			"id":       "snap1",
			"metadata": c.snapshotMetadata["snap1"].Get(),
		})
	}
	if query == "" || query == "snap2" {
		results = append(results, map[string]interface{}{
			"id":       "snap2",
			"metadata": c.snapshotMetadata["snap2"].Get(),
		})
	}

	return results
}

// QueryImages returns all images with metadata
func (c *CRDTCatalog) QueryImages(query string) []map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var results []map[string]interface{}

	if query == "" || query == "img1" {
		results = append(results, map[string]interface{}{
			"id":       "img1",
			"metadata": c.imageMetadata["img1"].Get(),
		})
	}

	return results
}

// GenerateDelta returns pending deltas for gossip
func (c *CRDTCatalog) GenerateDelta() []*Delta {
	c.mu.RLock()
	defer c.mu.RUnlock()

	deltas := make([]*Delta, len(c.deltas))
	copy(deltas, c.deltas)
	return deltas
}

// ApplyDelta applies a received delta
func (c *CRDTCatalog) ApplyDelta(delta *Delta) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if delta is already applied (causal ordering)
	comparison := c.vectorClock.Compare(delta.VectorClock)
	if comparison > 0 {
		// Our clock is ahead, ignore this delta
		return false
	}

	// Update our vector clock
	c.vectorClock.Merge(delta.VectorClock)
	c.vectorClock.Increment(c.nodeID)

	// Apply the delta based on type
	switch delta.Type {
	case "orset":
		c.applyORSetDelta(delta)
	case "lww":
		c.applyLWWDelta(delta)
	}

	return true
}

// applyORSetDelta applies an OR-Set delta
func (c *CRDTCatalog) applyORSetDelta(delta *Delta) {
	parts := strings.Split(delta.Key, ":")
	if len(parts) < 2 {
		return
	}

	setType := parts[0]
	itemID := parts[1]

	switch setType {
	case "snapshots":
		if len(parts) == 3 && parts[2] == "remove" {
			c.snapshots.Remove(itemID)
		} else {
			if tag, ok := delta.Data["tag"].(string); ok {
				c.snapshots.addWithTag(itemID, tag)
			}
			if metadata, ok := delta.Data["metadata"].(map[string]interface{}); ok {
				if c.snapshotMetadata[itemID] == nil {
					c.snapshotMetadata[itemID] = NewLWWRegister(delta.NodeID)
				}
				c.snapshotMetadata[itemID].Set(metadata)
			}
		}
	case "images":
		if tag, ok := delta.Data["tag"].(string); ok {
			c.images.addWithTag(itemID, tag)
		}
		if metadata, ok := delta.Data["metadata"].(map[string]interface{}); ok {
			if c.imageMetadata[itemID] == nil {
				c.imageMetadata[itemID] = NewLWWRegister(delta.NodeID)
			}
			c.imageMetadata[itemID].Set(metadata)
		}
	}
}

// applyLWWDelta applies an LWW delta
func (c *CRDTCatalog) applyLWWDelta(delta *Delta) {
	parts := strings.Split(delta.Key, ":")
	if len(parts) < 2 {
		return
	}

	fieldType := parts[0]
	itemID := parts[1]

	switch fieldType {
	case "snapshot_metadata":
		if c.snapshotMetadata[itemID] == nil {
			c.snapshotMetadata[itemID] = NewLWWRegister(delta.NodeID)
		}
		c.snapshotMetadata[itemID].Merge(&LWWRegister{
			value:     delta.Data,
			timestamp: delta.Timestamp,
			nodeID:    delta.NodeID,
		})
	}
}

// ClearDeltas clears processed deltas
func (c *CRDTCatalog) ClearDeltas() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deltas = c.deltas[:0]
}

// addWithTag adds an item with a specific tag (for delta application)
func (s *ORSet) addWithTag(item, tag string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.addSet[item] == nil {
		s.addSet[item] = make(map[string]bool)
	}
	s.addSet[item][tag] = true
}
