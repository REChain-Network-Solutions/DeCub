package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/rechain/rechain/src/catalog"
	"github.com/rechain/rechain/src/gcl"
	"github.com/rechain/rechain/src/snapshot"
)

var (
	snapshotCatalog *catalog.ORSet
	snapshotMutex   sync.RWMutex
	snapshots       = make(map[string]*snapshot.SnapshotMetadata)
)

func init() {
	snapshotCatalog = catalog.NewORSet()
}

func StartServer() {
	r := mux.NewRouter()
	r.HandleFunc("/snapshot/create", createSnapshotHandler).Methods("POST")
	r.HandleFunc("/snapshot/{id}", getSnapshotHandler).Methods("GET")
	r.HandleFunc("/snapshot/{id}/restore", restoreSnapshotHandler).Methods("POST")
	r.HandleFunc("/gcl/tx", publishTxHandler).Methods("POST")
	r.HandleFunc("/catalog/query", queryCatalogHandler).Methods("GET")
	r.HandleFunc("/catalog/add", addToCatalogHandler).Methods("POST")
	http.ListenAndServe(":8080", r)
}

func createSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Cluster string `json:"cluster"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Hardcoded for PoC; in real, from env/config
	etcdEndpoint := "http://localhost:2379"
	minioEndpoint := "localhost:9000"
	accessKey := "rechain"
	secretKey := "rechain123"
	bucket := "rechain-snapshots"

	meta, err := snapshot.CreateSnapshot(etcdEndpoint, minioEndpoint, accessKey, secretKey, bucket, req.Cluster)
	if err != nil {
		log.Printf("Failed to create snapshot: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Publish to GCL
	payload := fmt.Sprintf("snapshot-%s-%s", meta.ID, meta.MerkleRoot)
	txID, err := gcl.PublishTx("snapshot", payload)
	if err != nil {
		log.Printf("Failed to publish tx: %v", err)
		// Continue anyway
	} else {
		log.Printf("Published snapshot tx: %s", txID)
	}

	// Add to catalog
	snapshotMutex.Lock()
	snapshots[meta.ID] = meta
	snapshotCatalog.Add(meta.ID, meta)
	snapshotMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meta)
}

func getSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	snapshotMutex.RLock()
	meta, exists := snapshots[id]
	snapshotMutex.RUnlock()

	if !exists {
		http.Error(w, "Snapshot not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meta)
}

func restoreSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	snapshotMutex.RLock()
	meta, exists := snapshots[id]
	snapshotMutex.RUnlock()

	if !exists {
		http.Error(w, "Snapshot not found", http.StatusNotFound)
		return
	}

	// Hardcoded for PoC
	etcdEndpoint := "http://localhost:2379"
	minioEndpoint := "localhost:9000"
	accessKey := "rechain"
	secretKey := "rechain123"
	bucket := "rechain-snapshots"

	err := snapshot.RestoreSnapshot(meta, etcdEndpoint, minioEndpoint, accessKey, secretKey, bucket)
	if err != nil {
		log.Printf("Failed to restore snapshot: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Snapshot restored successfully"))
}

func publishTxHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type    string `json:"type"`
		Payload string `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	txID, err := gcl.PublishTx(req.Type, req.Payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]string{"tx_id": txID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func queryCatalogHandler(w http.ResponseWriter, r *http.Request) {
	result := snapshotCatalog.Query()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func addToCatalogHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	snapshotCatalog.Add(req.Key, req.Value)
	w.Write([]byte("Added to catalog"))
}
