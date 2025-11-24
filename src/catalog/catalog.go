package catalog

import (
	"sync"
	"time"
)

// ORSet represents an Observed-Remove Set for CRDT metadata
type ORSet struct {
	mu      sync.RWMutex
	elements map[string]*Element
}

type Element struct {
	Value     interface{}
	Timestamp time.Time
	Removed   bool
}

// NewORSet creates a new OR-Set
func NewORSet() *ORSet {
	return &ORSet{
		elements: make(map[string]*Element),
	}
}

// Add adds an element to the set
func (s *ORSet) Add(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.elements[key] = &Element{
		Value:     value,
		Timestamp: time.Now(),
		Removed:   false,
	}
}

// Remove marks an element as removed
func (s *ORSet) Remove(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if elem, exists := s.elements[key]; exists {
		elem.Removed = true
	}
}

// Query returns all non-removed elements
func (s *ORSet) Query() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]interface{})
	for key, elem := range s.elements {
		if !elem.Removed {
			result[key] = elem.Value
		}
	}
	return result
}

// Get returns a specific element if not removed
func (s *ORSet) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if elem, exists := s.elements[key]; exists && !elem.Removed {
		return elem.Value, true
	}
	return nil, false
}

// Merge merges another OR-Set (simplified, assumes no conflicts)
func (s *ORSet) Merge(other *ORSet) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, elem := range other.elements {
		if existing, exists := s.elements[key]; !exists || elem.Timestamp.After(existing.Timestamp) {
			s.elements[key] = elem
		}
	}
}
