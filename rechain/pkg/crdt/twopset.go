package crdt

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// TwoPhaseSet is a state-based two-phase set CRDT (CmRDT)
// It allows an element to be added and removed, but not re-added after removal
type TwoPhaseSet struct {
	nodeID string
	// Using RWMutex for better read performance
	mu       sync.RWMutex
	// Using sync.Map for better concurrent read performance
	added    sync.Map // map[interface{}]struct{} // Elements that have been added
	removed  sync.Map // map[interface{}]struct{} // Elements that have been removed
	// Cache for Elements() to avoid allocations on frequent calls
	elementsCache     []interface{}
	elementsCacheTime int64 // UnixNano timestamp of last cache update
}

// NewTwoPhaseSet creates a new TwoPhaseSet
func NewTwoPhaseSet(nodeID string) *TwoPhaseSet {
	return &TwoPhaseSet{
		nodeID:           nodeID,
		added:            sync.Map{},
		removed:          sync.Map{},
		elementsCache:    make([]interface{}, 0, 16), // Pre-allocate some space
		elementsCacheTime: 0,
	}
}

// Type returns the CRDT type
func (s *TwoPhaseSet) Type() CRDTType {
	return "2pset"
}

// Add adds an element to the set if it hasn't been removed
func (s *TwoPhaseSet) Add(element interface{}) {
	// Fast path: check if already added without lock
	if _, exists := s.added.Load(element); exists {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring lock
	if _, removed := s.removed.Load(element); !removed {
		s.added.Store(element, struct{}{})
		// Invalidate cache
		s.elementsCache = nil
		s.elementsCacheTime = 0
	}
}

// Remove removes an element from the set
func (s *TwoPhaseSet) Remove(element interface{}) {
	// Fast path: check if already removed without lock
	if _, exists := s.removed.Load(element); exists {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring lock
	if _, exists := s.added.Load(element); exists {
		s.removed.Store(element, struct{}{})
		// Invalidate cache
		s.elementsCache = nil
		s.elementsCacheTime = 0
	}
}

// Contains checks if an element is in the set
// This is optimized for read performance using RWMutex
func (s *TwoPhaseSet) Contains(element interface{}) bool {
	// Fast path: check if it's in the removed set first (no lock needed for Load)
	if _, removed := s.removed.Load(element); removed {
		return false
	}

	// Check if it's in the added set (no lock needed for Load)
	_, added := s.added.Load(element)
	return added
}

// Elements returns all elements in the set
// This method uses a cached result if available and still valid
func (s *TwoPhaseSet) Elements() []interface{} {
	// Try to use cache first (read-lock only)
	if cache := s.getCachedElements(); cache != nil {
		return cache
	}

	// Cache miss, need to rebuild
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check cache again after acquiring lock (double-checked locking)
	if s.elementsCache != nil {
		return s.elementsCache
	}

	// Rebuild cache
	var elements []interface{}
	s.added.Range(func(key, value interface{}) bool {
		if _, removed := s.removed.Load(key); !removed {
			elements = append(elements, key)
		}
		return true
	})

	// Update cache
	s.elementsCache = elements
	s.elementsCacheTime = time.Now().UnixNano()

	return elements
}

// getCachedElements returns the cached elements if they're still valid
func (s *TwoPhaseSet) getCachedElements() []interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Cache is valid for 100ms
	if s.elementsCache != nil && time.Since(time.Unix(0, s.elementsCacheTime)) < 100*time.Millisecond {
		return s.elementsCache
	}
	return nil
}

// Merge merges another TwoPhaseSet into this one
func (s *TwoPhaseSet) Merge(other CRDT) error {
	otherSet, ok := other.(*TwoPhaseSet)
	if !ok {
		return fmt.Errorf("%w: expected *TwoPhaseSet, got %T", ErrIncompatibleTypes, other)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Invalidate cache
	s.elementsCache = nil
	s.elementsCacheTime = 0

	// Process added elements
	otherSet.added.Range(func(key, value interface{}) bool {
		s.added.Store(key, struct{}{})
		// If it was removed in the other set, remove it from added
		if _, removed := otherSet.removed.Load(key); removed {
			s.added.Delete(key)
		}
		return true
	})

	// Process removed elements
	otherSet.removed.Range(func(key, value interface{}) bool {
		s.removed.Store(key, struct{}{})
		s.added.Delete(key)
		return true
	})

	return nil
}

// Value returns the current value of the set
func (s *TwoPhaseSet) Value() interface{} {
	return s.Elements()
}

// Marshal serializes the TwoPhaseSet to JSON
func (s *TwoPhaseSet) Marshal() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Pre-allocate slices with estimated capacity
	added := make([]interface{}, 0, 16)
	s.added.Range(func(key, value interface{}) bool {
		added = append(added, key)
		return true
	})

	removed := make([]interface{}, 0, 8)
	s.removed.Range(func(key, value interface{}) bool {
		removed = append(removed, key)
		return true
	})

	data := struct {
		Type    string        `json:"type"`
		Added   []interface{} `json:"added"`
		Removed []interface{} `json:"removed,omitempty"`
	}{
		Type:    string(s.Type()),
		Added:   added,
		Removed: removed,
	}

	return json.Marshal(data)
}

// Unmarshal deserializes the TwoPhaseSet from JSON
func (s *TwoPhaseSet) Unmarshal(data []byte) error {
	var aux struct {
		Type    string        `json:"type"`
		Added   []interface{} `json:"added"`
		Removed []interface{} `json:"removed"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Type != string(s.Type()) {
		return fmt.Errorf("%w: expected %s, got %s", ErrIncompatibleTypes, s.Type(), aux.Type)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Reset the maps
	s.added = sync.Map{}
	s.removed = sync.Map{}

	// Add elements
	for _, elem := range aux.Added {
		s.added.Store(elem, struct{}{})
	}

	// Remove elements
	for _, elem := range aux.Removed {
		s.removed.Store(elem, struct{}{})
		s.added.Delete(elem)
	}

	// Reset cache
	s.elementsCache = nil
	s.elementsCacheTime = 0

	return nil
}
