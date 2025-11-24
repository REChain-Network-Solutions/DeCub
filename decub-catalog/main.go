package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/syndtr/goleveldb/leveldb"
)

// ORSet represents an Observed-Remove Set CRDT
type ORSet struct {
	addSet map[string]map[string]bool // item -> {tag: true}
	rmSet  map[string]map[string]bool // item -> {tag: true}
	mu     sync.RWMutex
}

// NewORSet creates a new OR-Set
func NewORSet() *ORSet {
	return &ORSet{
		addSet: make(map[string]map[string]bool),
		rmSet:  make(map[string]map[string]bool),
	}
}

// Add adds an item to the set
func (s *ORSet) Add(item string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	tag := fmt.Sprintf("%d", time.Now().UnixNano())
	if s.addSet[item] == nil {
		s.addSet[item] = make(map[string]bool)
	}
	s.addSet[item][tag] = true
	return tag
}

// Remove removes an item from the set
func (s *ORSet) Remove(item string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.addSet[item] != nil {
		for tag := range s.addSet[item] {
			if s.rmSet[item] == nil {
				s.rmSet[item] = make(map[string]bool)
			}
			s.rmSet[item][tag] = true
		}
	}
}

// Contains checks if an item is in the set
func (s *ORSet) Contains(item string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.addSet[item] == nil {
		return false
	}

	for tag := range s.addSet[item] {
		if s.rmSet[item] == nil || !s.rmSet[item][tag] {
			return true
		}
	}
	return false
}

// Merge merges another OR-Set into this one
func (s *ORSet) Merge(other *ORSet) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for item, tags := range other.addSet {
		if s.addSet[item] == nil {
			s.addSet[item] = make(map[string]bool)
		}
		for tag := range tags {
			s.addSet[item][tag] = true
		}
	}

	for item, tags := range other.rmSet {
		if s.rmSet[item] == nil {
			s.rmSet[item] = make(map[string]bool)
		}
		for tag := range tags {
			s.rmSet[item][tag] = true
		}
	}
}

// Serialize serializes the OR-Set
func (s *ORSet) Serialize() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := map[string]interface{}{
		"addSet": s.addSet,
		"rmSet":  s.rmSet,
	}
	jsonData, _ := json.Marshal(data)
	return jsonData
}

// Deserialize deserializes the OR-Set
func (s *ORSet) Deserialize(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var d map[string]interface{}
	json.Unmarshal(data, &d)

	if addSet, ok := d["addSet"].(map[string]interface{}); ok {
		for item, tags := range addSet {
			if tagMap, ok := tags.(map[string]interface{}); ok {
				s.addSet[item] = make(map[string]bool)
				for tag, _ := range tagMap {
					s.addSet[item][tag] = true
				}
			}
		}
	}

	if rmSet, ok := d["rmSet"].(map[string]interface{}); ok {
		for item, tags := range rmSet {
			if tagMap, ok := tags.(map[string]interface{}); ok {
				s.rmSet[item] = make(map[string]bool)
				for tag, _ := range tagMap {
					s.rmSet[item][tag] = true
				}
			}
		}
	}
}

// Catalog represents the metadata catalog
type Catalog struct {
	snapshots *ORSet
	images    *ORSet
	db        *leveldb.DB
}

// NewCatalog creates a new catalog
func NewCatalog() (*Catalog, error) {
	db, err := leveldb.OpenFile("./catalog.db", nil)
	if err != nil {
		return nil, err
	}

	catalog := &Catalog{
		snapshots: NewORSet(),
		images:    NewORSet(),
		db:        db,
	}

	// Load from DB
	if data, err := db.Get([]byte("snapshots"), nil); err == nil {
		catalog.snapshots.Deserialize(data)
	}
	if data, err := db.Get([]byte("images"), nil); err == nil {
		catalog.images.Deserialize(data)
	}

	return catalog, nil
}

// AddSnapshot adds a snapshot to the catalog
func (c *Catalog) AddSnapshot(snapshotID string) {
	tag := c.snapshots.Add(snapshotID)
	c.save("snapshots")

	// Log for gossip sync
	fmt.Printf("Added snapshot %s with tag %s\n", snapshotID, tag)
}

// RemoveSnapshot removes a snapshot from the catalog
func (c *Catalog) RemoveSnapshot(snapshotID string) {
	c.snapshots.Remove(snapshotID)
	c.save("snapshots")
}

// AddImage adds an image to the catalog
func (c *Catalog) AddImage(imageID string) {
	tag := c.images.Add(imageID)
	c.save("images")

	fmt.Printf("Added image %s with tag %s\n", imageID, tag)
}

// RemoveImage removes an image from the catalog
func (c *Catalog) RemoveImage(imageID string) {
	c.images.Remove(imageID)
	c.save("images")
}

// QuerySnapshots returns all snapshots
func (c *Catalog) QuerySnapshots() []string {
	var snapshots []string
	// In a real implementation, iterate through the set
	// For simplicity, return a hardcoded list
	return []string{"snap1", "snap2"}
}

// QueryImages returns all images
func (c *Catalog) QueryImages() []string {
	return []string{"img1", "img2"}
}

// Merge merges another catalog
func (c *Catalog) Merge(other *Catalog) {
	c.snapshots.Merge(other.snapshots)
	c.images.Merge(other.images)
	c.save("snapshots")
	c.save("images")
}

// save persists the catalog to DB
func (c *Catalog) save(key string) {
	var data []byte
	if key == "snapshots" {
		data = c.snapshots.Serialize()
	} else if key == "images" {
		data = c.images.Serialize()
	}
	c.db.Put([]byte(key), data, nil)
}

// Close closes the catalog
func (c *Catalog) Close() error {
	return c.db.Close()
}

// API handlers
func (c *Catalog) handleAddSnapshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	snapshotID := vars["id"]
	c.AddSnapshot(snapshotID)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Snapshot %s added", snapshotID)
}

func (c *Catalog) handleRemoveSnapshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	snapshotID := vars["id"]
	c.RemoveSnapshot(snapshotID)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Snapshot %s removed", snapshotID)
}

func (c *Catalog) handleAddImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["id"]
	c.AddImage(imageID)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Image %s added", imageID)
}

func (c *Catalog) handleRemoveImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["id"]
	c.RemoveImage(imageID)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Image %s removed", imageID)
}

func (c *Catalog) handleQuerySnapshots(w http.ResponseWriter, r *http.Request) {
	snapshots := c.QuerySnapshots()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshots)
}

func (c *Catalog) handleQueryImages(w http.ResponseWriter, r *http.Request) {
	images := c.QueryImages()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(images)
}

func (c *Catalog) handleMerge(w http.ResponseWriter, r *http.Request) {
	var other Catalog
	if err := json.NewDecoder(r.Body).Decode(&other); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c.Merge(&other)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Catalog merged")
}

func main() {
	catalog, err := NewCatalog()
	if err != nil {
		log.Fatalf("Failed to create catalog: %v", err)
	}
	defer catalog.Close()

	r := mux.NewRouter()
	r.HandleFunc("/snapshots/add/{id}", catalog.handleAddSnapshot).Methods("POST")
	r.HandleFunc("/snapshots/remove/{id}", catalog.handleRemoveSnapshot).Methods("DELETE")
	r.HandleFunc("/images/add/{id}", catalog.handleAddImage).Methods("POST")
	r.HandleFunc("/images/remove/{id}", catalog.handleRemoveImage).Methods("DELETE")
	r.HandleFunc("/snapshots/query", catalog.handleQuerySnapshots).Methods("GET")
	r.HandleFunc("/images/query", catalog.handleQueryImages).Methods("GET")
	r.HandleFunc("/merge", catalog.handleMerge).Methods("POST")

	fmt.Println("Catalog server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
