package crdt

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
)

// ORSet is an Observed-Removed Set CRDT
type ORSet struct {
	nodeID string
	mu     sync.RWMutex
	adds   map[interface{}]map[string]struct{} // value -> set of add tags
	dels   map[interface{}]map[string]struct{} // value -> set of remove tags
}

// NewORSet creates a new ORSet
func NewORSet(nodeID string) *ORSet {
	return &ORSet{
		nodeID: nodeID,
		adds:   make(map[interface{}]map[string]struct{}),
		dels:   make(map[interface{}]map[string]struct{}),
	}
}

// Type returns the CRDT type
func (s *ORSet) Type() CRDTType {
	return "orset"
}

// Add adds an element to the set
func (s *ORSet) Add(element interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tag := s.generateTag()

	if _, exists := s.adds[element]; !exists {
		s.adds[element] = make(map[string]struct{})
	}

	s.adds[element][tag] = struct{}{}
}

// Remove removes an element from the set
func (s *ORSet) Remove(element interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Can only remove elements that have been added
	if adds, exists := s.adds[element]; exists {
		for tag := range adds {
			if _, exists := s.dels[element]; !exists {
				s.dels[element] = make(map[string]struct{})
			}
			s.dels[element][tag] = struct{}{}
		}
	}
}

// Contains checks if an element is in the set
func (s *ORSet) Contains(element interface{}) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	adds, added := s.adds[element]
	if !added {
		return false
	}

	dels, deleted := s.dels[element]
	if !deleted {
		return true
	}

	// An element is in the set if it has at least one add tag that is not in the remove set
	for tag := range adds {
		if _, exists := dels[tag]; !exists {
			return true
		}
	}

	return false
}

// Elements returns all elements in the set
func (s *ORSet) Elements() []interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var elements []interface{}

	for element := range s.adds {
		if s.Contains(element) {
			elements = append(elements, element)
		}
	}

	return elements
}

// Merge merges another ORSet into this one
func (s *ORSet) Merge(other CRDT) error {
	otherSet, ok := other.(*ORSet)
	if !ok {
		return fmt.Errorf("%w: expected *ORSet, got %T", ErrIncompatibleTypes, other)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Merge adds
	for element, tags := range otherSet.adds {
		if _, exists := s.adds[element]; !exists {
			s.adds[element] = make(map[string]struct{})
		}

		for tag := range tags {
			s.adds[element][tag] = struct{}{}
		}
	}

	// Merge deletes
	for element, tags := range otherSet.dels {
		if _, exists := s.dels[element]; !exists {
			s.dels[element] = make(map[string]struct{})
		}

		for tag := range tags {
			s.dels[element][tag] = struct{}{}
		}
	}

	return nil
}

// Value returns the current value of the set
func (s *ORSet) Value() interface{} {
	return s.Elements()
}

// Marshal serializes the ORSet to JSON
func (s *ORSet) Marshal() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Convert maps to serializable types
	adds := make(map[string][]string)
	for elem, tags := range s.adds {
		key := fmt.Sprint(elem)
		adds[key] = make([]string, 0, len(tags))
		for tag := range tags {
			adds[key] = append(adds[key], tag)
		}
		sort.Strings(adds[key])
	}

	dels := make(map[string][]string)
	for elem, tags := range s.dels {
		key := fmt.Sprint(elem)
		dels[key] = make([]string, 0, len(tags))
		for tag := range tags {
			dels[key] = append(dels[key], tag)
		}
		sort.Strings(dels[key])
	}

	data := struct {
		Type string              `json:"type"`
		Adds map[string][]string `json:"adds"`
		Dels map[string][]string `json:"dels,omitempty"`
	}{
		Type: string(s.Type()),
		Adds: adds,
		Dels: dels,
	}

	return json.Marshal(data)
}

// Unmarshal deserializes the ORSet from JSON
func (s *ORSet) Unmarshal(data []byte) error {
	var aux struct {
		Type string              `json:"type"`
		Adds map[string][]string `json:"adds"`
		Dels map[string][]string `json:"dels"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Type != string(s.Type()) {
		return fmt.Errorf("%w: expected %s, got %s", ErrIncompatibleTypes, s.Type(), aux.Type)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Convert back to internal representation
	s.adds = make(map[interface{}]map[string]struct{})
	for elemStr, tags := range aux.Adds {
		s.adds[elemStr] = make(map[string]struct{})
		for _, tag := range tags {
			s.adds[elemStr][tag] = struct{}{}
		}
	}

	s.dels = make(map[interface{}]map[string]struct{})
	for elemStr, tags := range aux.Dels {
		s.dels[elemStr] = make(map[string]struct{})
		for _, tag := range tags {
			s.dels[elemStr][tag] = struct{}{}
		}
	}

	return nil
}

// generateTag generates a unique tag for an operation
func (s *ORSet) generateTag() string {
	return fmt.Sprintf("%s-%d", s.nodeID, time.Now().UnixNano())
}
