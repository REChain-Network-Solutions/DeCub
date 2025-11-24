package crdt

import (
	"encoding/json"
	"fmt"
)

// LWWRegister is a Last-Write-Wins Register CRDT
type LWWRegister struct {
	NodeID    string    `json:"node_id"`
	Value     any       `json:"value"`
	Timestamp Timestamp `json:"timestamp"`
}

// NewLWWRegister creates a new LWWRegister
func NewLWWRegister(nodeID string) *LWWRegister {
	return &LWWRegister{
		NodeID:    nodeID,
		Value:     nil,
		Timestamp: NewTimestamp(),
	}
}

// Type returns the CRDT type
func (r *LWWRegister) Type() CRDTType {
	return LWWRegister
}

// Value returns the current value
func (r *LWWRegister) GetValue() any {
	return r.Value
}

// Set updates the value with a new value and timestamp
func (r *LWWRegister) Set(value any) {
	r.Value = value
	r.Timestamp = NewTimestamp()
}

// Merge merges another LWWRegister
func (r *LWWRegister) Merge(other CRDT) error {
	otherReg, ok := other.(*LWWRegister)
	if !ok {
		return fmt.Errorf("%w: expected LWWRegister, got %T", ErrIncompatibleTypes, other)
	}

	// Keep the value with the latest timestamp
	if otherReg.Timestamp.Compare(r.Timestamp) > 0 {
		r.Value = otherReg.Value
		r.Timestamp = otherReg.Timestamp
		r.NodeID = otherReg.NodeID
	}

	return nil
}

// Marshal serializes the LWWRegister to JSON
func (r *LWWRegister) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Unmarshal deserializes the LWWRegister from JSON
func (r *LWWRegister) Unmarshal(data []byte) error {
	return json.Unmarshal(data, r)
}

// Value implements the CRDT interface
func (r *LWWRegister) Value() interface{} {
	return r.GetValue()
}
