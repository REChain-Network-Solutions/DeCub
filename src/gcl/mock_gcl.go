package gcl

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Tx represents a transaction
type Tx struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Payload   string    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

// MockGCL simulates a single-node Tendermint-like ledger
type MockGCL struct {
	mu   sync.RWMutex
	txs  map[string]*Tx
	port string
}

// NewMockGCL creates a new mock GCL
func NewMockGCL(port string) *MockGCL {
	return &MockGCL{
		txs:  make(map[string]*Tx),
		port: port,
	}
}

// PublishTx publishes a tx to the mock GCL
func (g *MockGCL) PublishTx(txType, payload string) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	txID := uuid.New().String()
	tx := &Tx{
		ID:        txID,
		Type:      txType,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	g.txs[txID] = tx
	log.Printf("Published tx: %s (%s)", txID, txType)
	return txID, nil
}

// GetTx retrieves a tx by ID
func (g *MockGCL) GetTx(txID string) (*Tx, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	tx, exists := g.txs[txID]
	return tx, exists
}

// StartServer starts the REST server for the mock GCL
func (g *MockGCL) StartServer() {
	r := mux.NewRouter()
	r.HandleFunc("/tx", g.publishTxHandler).Methods("POST")
	r.HandleFunc("/tx/{id}", g.getTxHandler).Methods("GET")
	r.HandleFunc("/status", g.statusHandler).Methods("GET")

	log.Printf("Mock GCL server starting on port %s", g.port)
	http.ListenAndServe(":"+g.port, r)
}

func (g *MockGCL) publishTxHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type    string `json:"type"`
		Payload string `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	txID, err := g.PublishTx(req.Type, req.Payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]string{"tx_id": txID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (g *MockGCL) getTxHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	txID := vars["id"]

	tx, exists := g.GetTx(txID)
	if !exists {
		http.Error(w, "Tx not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx)
}

func (g *MockGCL) statusHandler(w http.ResponseWriter, r *http.Request) {
	g.mu.RLock()
	count := len(g.txs)
	g.mu.RUnlock()

	resp := map[string]interface{}{
		"status":    "running",
		"tx_count":  count,
		"timestamp": time.Now(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Global instance for easy access
var GlobalMockGCL *MockGCL

// PublishTx publishes a tx to mock GCL (global function for compatibility)
func PublishTx(txType, payload string) (string, error) {
	if GlobalMockGCL == nil {
		return "", fmt.Errorf("mock GCL not initialized")
	}
	return GlobalMockGCL.PublishTx(txType, payload)
}
