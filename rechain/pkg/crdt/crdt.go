package crdt

import (
	"encoding/json"
	"time"
)

// CRDTType represents the type of CRDT
type CRDTType string

const (
	// LWWRegister is a Last-Write-Wins Register
	LWWRegister CRDTType = "lww"
	// PNCounter is a Positive-Negative Counter
	PNCounter CRDTType = "pncounter"
)

// CRDT is the interface that all CRDT implementations must satisfy
type CRDT interface {
	// Type returns the type of the CRDT
	Type() CRDTType

	// Value returns the current value of the CRDT
	Value() interface{}

	// Merge merges another CRDT of the same type
	Merge(other CRDT) error

	// Marshal serializes the CRDT to bytes
	Marshal() ([]byte, error)

	// Unmarshal deserializes the CRDT from bytes
	Unmarshal(data []byte) error
}

// New creates a new CRDT instance of the specified type
func New(t CRDTType, nodeID string) (CRDT, error) {
	switch t {
	case LWWRegister:
		return NewLWWRegister(nodeID), nil
	case PNCounter:
		return NewPNCounter(nodeID), nil
	default:
		return nil, ErrUnknownCRDTType
	}
}

// Timestamp is a wrapper around time.Time that implements json.Marshaler and json.Unmarshaler
type Timestamp struct {
	time.Time
}

// NewTimestamp creates a new Timestamp with the current time
func NewTimestamp() Timestamp {
	return Timestamp{Time: time.Now().UTC()}
}

// MarshalJSON implements json.Marshaler
func (t Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time.UnixNano())
}

// UnmarshalJSON implements json.Unmarshaler
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	var nano int64
	if err := json.Unmarshal(data, &nano); err != nil {
		return err
	}
	t.Time = time.Unix(0, nano).UTC()
	return nil
}

// Compare compares two timestamps
// Returns -1 if t < other, 0 if t == other, 1 if t > other
func (t Timestamp) Compare(other Timestamp) int {
	switch {
	case t.Before(other.Time):
		return -1
	case t.After(other.Time):
		return 1
	default:
		return 0
	}
}

// Errors
var (
	ErrIncompatibleTypes = "incompatible CRDT types"
	ErrUnknownCRDTType  = "unknown CRDT type"
)
