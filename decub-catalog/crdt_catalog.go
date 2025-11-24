package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/syndtr/goleveldb/leveldb"
)

// CRDTService represents the CRDT catalog service
type CRDTService struct {
	catalog *CRDTCatalog
	db      *leveldb.DB
	mu      sync.RWMutex
}

// NewCRDTService creates a new CRDT service
func NewCRDTService(nodeID string) (*CRDTService, error) {
	db, err := leveldb.OpenFile("./crdt_catalog.db", nil)
	if err != nil {
		return nil, err
	}

	service := &CRDTService{
		catalog: NewCRDTCatalog(nodeID),
		db:      db,
	}

	// Load persisted state
	service.loadState()

	return service, nil
}

// loadState loads the catalog state from database
func (s *CRDTService) loadState() {
	// Load vector clock
	if data, err := s.db.Get([]byte("vector_clock"), nil); err == nil {
		json.Unmarshal(data, &s.catalog.vectorClock)
	}

	// Load OR-Sets
	if data, err := s.db.Get([]byte("snapshots"), nil); err == nil {
		s.catalog.snapshots.Deserialize(data)
	}
	if data, err := s.db.Get([]byte("images"), nil); err == nil {
		s.catalog.images.Deserialize(data)
	}

	// Load metadata (simplified - in production, use proper serialization)
}

// saveState persists the catalog state
func (s *CRDTService) saveState() {
	// Save vector clock
	if vcData, err := json.Marshal(s.catalog.vectorClock); err == nil {
		s.db.Put([]byte("vector_clock"), vcData, nil)
	}

	// Save OR-Sets
	s.db.Put([]byte("snapshots"), s.catalog.snapshots.Serialize(), nil)
	s.db.Put([]byte("images"), s.catalog.images.Serialize(), nil)
}

// AddSnapshot adds a snapshot with metadata
func (s *CRDTService) AddSnapshot(snapshotID string, metadata map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.catalog.AddSnapshot(snapshotID, metadata)
	s.saveState()
}

// RemoveSnapshot removes a snapshot
func (s *CRDTService) RemoveSnapshot(snapshotID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.catalog.RemoveSnapshot(snapshotID)
	s.saveState()
}

// UpdateSnapshotMetadata updates snapshot metadata
func (s *CRDTService) UpdateSnapshotMetadata(snapshotID string, metadata map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.catalog.UpdateSnapshotMetadata(snapshotID, metadata)
	s.saveState()
}

// AddImage adds an image with metadata
func (s *CRDTService) AddImage(imageID string, metadata map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.catalog.AddImage(imageID, metadata)
	s.saveState()
}

// QueryCatalog queries the catalog
func (s *CRDTService) QueryCatalog(queryType, query string) []map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	switch queryType {
	case "snapshots":
		return s.catalog.QuerySnapshots(query)
	case "images":
		return s.catalog.QueryImages(query)
	default:
		return []map[string]interface{}{}
	}
}

// GetDeltas returns pending deltas for gossip
func (s *CRDTService) GetDeltas() []*Delta {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.catalog.GenerateDelta()
}

// ApplyDelta applies a received delta
func (s *CRDTService) ApplyDelta(delta *Delta) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	applied := s.catalog.ApplyDelta(delta)
	if applied {
		s.saveState()
	}
	return applied
}

// ClearDeltas clears processed deltas
func (s *CRDTService) ClearDeltas() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.catalog.ClearDeltas()
}

// Close closes the service
func (s *CRDTService) Close() error {
	return s.db.Close()
}

// HTTP Handlers

func (s *CRDTService) handleAddSnapshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	snapshotID := vars["id"]

	var metadata map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&metadata); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	s.AddSnapshot(snapshotID, metadata)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "added", "id": snapshotID})
}

func (s *CRDTService) handleRemoveSnapshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	snapshotID := vars["id"]

	s.RemoveSnapshot(snapshotID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "removed", "id": snapshotID})
}

func (s *CRDTService) handleUpdateSnapshotMetadata(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	snapshotID := vars["id"]

	var metadata map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&metadata); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	s.UpdateSnapshotMetadata(snapshotID, metadata)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated", "id": snapshotID})
}

func (s *CRDTService) handleAddImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["id"]

	var metadata map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&metadata); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	s.AddImage(imageID, metadata)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "added", "id": imageID})
}

func (s *CRDTService) handleQuery(w http.ResponseWriter, r *http.Request) {
	queryType := r.URL.Query().Get("type")
	query := r.URL.Query().Get("q")

	results := s.QueryCatalog(queryType, query)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *CRDTService) handleGetDeltas(w http.ResponseWriter, r *http.Request) {
	deltas := s.GetDeltas()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deltas)
}

func (s *CRDTService) handleApplyDelta(w http.ResponseWriter, r *http.Request) {
	var delta Delta
	if err := json.NewDecoder(r.Body).Decode(&delta); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	applied := s.ApplyDelta(&delta)
	response := map[string]bool{"applied": applied}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *CRDTService) handleClearDeltas(w http.ResponseWriter, r *http.Request) {
	s.ClearDeltas()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
}

func main() {
	nodeID := "node1" // In production, generate unique node ID

	service, err := NewCRDTService(nodeID)
	if err != nil {
		log.Fatalf("Failed to create CRDT service: %v", err)
	}
	defer service.Close()

	r := mux.NewRouter()

	// Snapshot operations
	r.HandleFunc("/snapshots/add/{id}", service.handleAddSnapshot).Methods("POST")
	r.HandleFunc("/snapshots/remove/{id}", service.handleRemoveSnapshot).Methods("DELETE")
	r.HandleFunc("/snapshots/metadata/{id}", service.handleUpdateSnapshotMetadata).Methods("PUT")

	// Image operations
	r.HandleFunc("/images/add/{id}", service.handleAddImage).Methods("POST")

	// Query operations
	r.HandleFunc("/catalog/query", service.handleQuery).Methods("GET")

	// CRDT operations for gossip
	r.HandleFunc("/crdt/delta", service.handleGetDeltas).Methods("GET")
	r.HandleFunc("/crdt/delta", service.handleApplyDelta).Methods("POST")
	r.HandleFunc("/crdt/delta/clear", service.handleClearDeltas).Methods("POST")

	fmt.Printf("CRDT Catalog service starting on :8080 (Node ID: %s)\n", nodeID)
	log.Fatal(http.ListenAndServe(":8080", r))
}
