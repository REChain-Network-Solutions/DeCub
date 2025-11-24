package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rechain/rechain/internal/cas"
	"github.com/rechain/rechain/internal/consensus"
	"github.com/rechain/rechain/internal/gossip"
	"github.com/rechain/rechain/internal/security"
	"github.com/rechain/rechain/internal/storage"
)

// Server represents the API server
type Server struct {
	consensus *consensus.Consensus
	store     storage.Store
	cas       *cas.CAS
	gossip    *gossip.GossipProtocol
	security  *security.KeyManager
	httpServer *http.Server
	router     *mux.Router
}

// NewServer creates a new API server
func NewServer(consensus *consensus.Consensus, store storage.Store, cas *cas.CAS, gossip *gossip.GossipProtocol, security *security.KeyManager) *Server {
	srv := &Server{
		consensus: consensus,
		store:     store,
		cas:       cas,
		gossip:    gossip,
		security:  security,
		router:    mux.NewRouter(),
	}

	srv.routes()

	return srv
}

// Start starts the API server
func (s *Server) Start(addr string) error {
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	log.Printf("API server starting on %s", addr)
	return s.httpServer.ListenAndServe()
}

// Stop gracefully stops the API server
func (s *Server) Stop() error {
	if s.httpServer == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}

// routes defines all API routes
func (s *Server) routes() {
	// Health check
	s.router.HandleFunc("/health", s.handleHealthCheck).Methods("GET")

	// Block operations
	s.router.HandleFunc("/blocks/latest", s.handleGetLatestBlock).Methods("GET")
	s.router.HandleFunc("/blocks/{height:[0-9]+}", s.handleGetBlock).Methods("GET")
	s.router.HandleFunc("/blocks", s.handleGetBlocks).Methods("GET")

	// Transaction operations
	s.router.HandleFunc("/txs", s.handleSubmitTx).Methods("POST")
	s.router.HandleFunc("/txs/{hash}", s.handleGetTx).Methods("GET")
	s.router.HandleFunc("/txs", s.handleGetTxs).Methods("GET")

	// CAS operations
	s.router.HandleFunc("/cas/objects", s.handleStoreObject).Methods("POST")
	s.router.HandleFunc("/cas/objects/{cid}", s.handleGetObject).Methods("GET")
	s.router.HandleFunc("/cas/objects/{cid}", s.handleDeleteObject).Methods("DELETE")
	s.router.HandleFunc("/cas/objects", s.handleListObjects).Methods("GET")

	// Gossip operations
	s.router.HandleFunc("/gossip/state", s.handleGetGossipState).Methods("GET")
	s.router.HandleFunc("/gossip/state", s.handleUpdateGossipState).Methods("POST")
	s.router.HandleFunc("/gossip/query", s.handleQueryGossip).Methods("POST")

	// Node info
	s.router.HandleFunc("/node/info", s.handleNodeInfo).Methods("GET")
	s.router.HandleFunc("/node/peers", s.handleGetPeers).Methods("GET")

	// Consensus state
	s.router.HandleFunc("/consensus/state", s.handleGetConsensusState).Methods("GET")
}

// API Response Helpers
func (s *Server) respond(w http.ResponseWriter, r *http.Request, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

func (s *Server) error(w http.ResponseWriter, r *http.Request, err error, status int) {
	s.respond(w, r, map[string]string{
		"error": err.Error(),
	}, status)
}

// Handlers
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	s.respond(w, r, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}, http.StatusOK)
}

func (s *Server) handleGetLatestBlock(w http.ResponseWriter, r *http.Request) {
	// Get latest block from storage
	// This is simplified - in production, get from consensus
	key := []byte("latest-block")
	data, err := s.store.Get(context.Background(), key)
	if err != nil {
		s.error(w, r, fmt.Errorf("failed to get latest block: %w", err), http.StatusInternalServerError)
		return
	}

	if data == nil {
		s.respond(w, r, map[string]string{"message": "No blocks yet"}, http.StatusOK)
		return
	}

	var block map[string]interface{}
	if err := json.Unmarshal(data, &block); err != nil {
		s.error(w, r, err, http.StatusInternalServerError)
		return
	}

	s.respond(w, r, block, http.StatusOK)
}

func (s *Server) handleGetBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	heightStr := vars["height"]
	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		s.error(w, r, err, http.StatusBadRequest)
		return
	}

	// Get block by height
	key := []byte(fmt.Sprintf("block/%d", height))
	data, err := s.store.Get(context.Background(), key)
	if err != nil {
		s.error(w, r, fmt.Errorf("failed to get block: %w", err), http.StatusInternalServerError)
		return
	}

	if data == nil {
		s.error(w, r, fmt.Errorf("block not found"), http.StatusNotFound)
		return
	}

	var block map[string]interface{}
	if err := json.Unmarshal(data, &block); err != nil {
		s.error(w, r, err, http.StatusInternalServerError)
		return
	}

	s.respond(w, r, block, http.StatusOK)
}

func (s *Server) handleGetBlocks(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := uint64(10) // default
	if limitStr != "" {
		if l, err := strconv.ParseUint(limitStr, 10, 64); err == nil && l > 0 {
			limit = l
		}
	}

	// This is simplified - in production, implement proper block listing
	blocks := make([]map[string]interface{}, 0)

	// Get latest blocks (simplified)
	for i := uint64(1); i <= limit; i++ {
		key := []byte(fmt.Sprintf("block/%d", i))
		data, err := s.store.Get(context.Background(), key)
		if err != nil || data == nil {
			break
		}

		var block map[string]interface{}
		if err := json.Unmarshal(data, &block); err == nil {
			blocks = append(blocks, block)
		}
	}

	s.respond(w, r, map[string]interface{}{
		"blocks": blocks,
		"count":  len(blocks),
	}, http.StatusOK)
}

func (s *Server) handleSubmitTx(w http.ResponseWriter, r *http.Request) {
	var txReq struct {
		Type    string                 `json:"type"`
		Payload map[string]interface{} `json:"payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&txReq); err != nil {
		s.error(w, r, err, http.StatusBadRequest)
		return
	}

	// Create transaction
	tx := &consensus.Transaction{
		ID:        fmt.Sprintf("tx-%d", time.Now().UnixNano()),
		Type:      txReq.Type,
		Payload:   nil, // Serialize payload
		Timestamp: time.Now(),
		Sender:    "api-client", // In production, get from auth
	}

	payloadBytes, err := json.Marshal(txReq.Payload)
	if err != nil {
		s.error(w, r, err, http.StatusBadRequest)
		return
	}
	tx.Payload = payloadBytes

	// Add to consensus mempool
	s.consensus.AddTransaction(tx)

	s.respond(w, r, map[string]interface{}{
		"tx_id":     tx.ID,
		"status":    "submitted",
		"timestamp": tx.Timestamp.Format(time.RFC3339),
	}, http.StatusOK)
}

func (s *Server) handleGetTx(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	txHash := vars["hash"]

	// This is simplified - in production, query transaction index
	key := []byte(fmt.Sprintf("tx/%s", txHash))
	data, err := s.store.Get(context.Background(), key)
	if err != nil {
		s.error(w, r, fmt.Errorf("failed to get transaction: %w", err), http.StatusInternalServerError)
		return
	}

	if data == nil {
		s.error(w, r, fmt.Errorf("transaction not found"), http.StatusNotFound)
		return
	}

	var tx map[string]interface{}
	if err := json.Unmarshal(data, &tx); err != nil {
		s.error(w, r, err, http.StatusInternalServerError)
		return
	}

	s.respond(w, r, tx, http.StatusOK)
}

func (s *Server) handleGetTxs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := uint64(10) // default
	if limitStr != "" {
		if l, err := strconv.ParseUint(limitStr, 10, 64); err == nil && l > 0 {
			limit = l
		}
	}

	// This is simplified - in production, implement proper transaction listing
	txs := make([]map[string]interface{}, 0)

	// Get recent transactions (simplified)
	for i := uint64(1); i <= limit; i++ {
		key := []byte(fmt.Sprintf("tx/tx-%d", i))
		data, err := s.store.Get(context.Background(), key)
		if err != nil || data == nil {
			break
		}

		var tx map[string]interface{}
		if err := json.Unmarshal(data, &tx); err == nil {
			txs = append(txs, tx)
		}
	}

	s.respond(w, r, map[string]interface{}{
		"transactions": txs,
		"count":        len(txs),
	}, http.StatusOK)
}

func (s *Server) handleStoreObject(w http.ResponseWriter, r *http.Request) {
	// Parse metadata from headers
	metadata := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 && key != "Content-Type" {
			metadata[key] = values[0]
		}
	}

	// Store object in CAS
	objInfo, err := s.cas.Store(context.Background(), r.Body, metadata)
	if err != nil {
		s.error(w, r, fmt.Errorf("failed to store object: %w", err), http.StatusInternalServerError)
		return
	}

	s.respond(w, r, map[string]interface{}{
		"cid":         objInfo.CID,
		"size":        objInfo.Size,
		"chunks":      len(objInfo.Chunks),
		"merkle_root": objInfo.MerkleRoot,
		"uploaded":    objInfo.Uploaded.Format(time.RFC3339),
	}, http.StatusCreated)
}

func (s *Server) handleGetObject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cid := vars["cid"]

	// Retrieve object from CAS
	reader, err := s.cas.Retrieve(context.Background(), cid)
	if err != nil {
		s.error(w, r, fmt.Errorf("failed to retrieve object: %w", err), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	// Stream object to response
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("X-Content-ID", cid)
	io.Copy(w, reader)
}

func (s *Server) handleDeleteObject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cid := vars["cid"]

	// Delete object from CAS
	if err := s.cas.Delete(context.Background(), cid); err != nil {
		s.error(w, r, fmt.Errorf("failed to delete object: %w", err), http.StatusInternalServerError)
		return
	}

	s.respond(w, r, map[string]string{"message": "Object deleted"}, http.StatusOK)
}

func (s *Server) handleListObjects(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	prefix := r.URL.Query().Get("prefix")

	// List objects from CAS
	objects, err := s.cas.List(context.Background(), prefix)
	if err != nil {
		s.error(w, r, fmt.Errorf("failed to list objects: %w", err), http.StatusInternalServerError)
		return
	}

	s.respond(w, r, map[string]interface{}{
		"objects": objects,
		"count":   len(objects),
	}, http.StatusOK)
}

func (s *Server) handleGetGossipState(w http.ResponseWriter, r *http.Request) {
	// This is simplified - in production, get from gossip protocol
	state := make(map[string]interface{})

	// Get some sample state
	if value, exists := s.gossip.GetCRDT("example-key"); exists {
		state["example-key"] = value
	}

	s.respond(w, r, map[string]interface{}{
		"state": state,
		"peers": "unknown", // In production, get peer count
	}, http.StatusOK)
}

func (s *Server) handleUpdateGossipState(w http.ResponseWriter, r *http.Request) {
	var updateReq struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		s.error(w, r, err, http.StatusBadRequest)
		return
	}

	// Update CRDT state
	if err := s.gossip.UpdateCRDT(updateReq.Key, updateReq.Value); err != nil {
		s.error(w, r, err, http.StatusInternalServerError)
		return
	}

	s.respond(w, r, map[string]string{"message": "State updated"}, http.StatusOK)
}

func (s *Server) handleQueryGossip(w http.ResponseWriter, r *http.Request) {
	var queryReq struct {
		Key string `json:"key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&queryReq); err != nil {
		s.error(w, r, err, http.StatusBadRequest)
		return
	}

	// Query CRDT state
	if err := s.gossip.QueryCRDT(queryReq.Key); err != nil {
		s.error(w, r, err, http.StatusInternalServerError)
		return
	}

	s.respond(w, r, map[string]string{"message": "Query sent"}, http.StatusOK)
}

func (s *Server) handleNodeInfo(w http.ResponseWriter, r *http.Request) {
	// Get node information
	info := map[string]interface{}{
		"version":       "0.1.0",
		"network":       "rechain-mainnet",
		"consensus":     "bft",
		"start_time":    time.Now().Format(time.RFC3339), // In production, track actual start time
		"peers":        0, // In production, get from gossip/p2p
		"latest_block": 0, // In production, get from consensus
	}
	s.respond(w, r, info, http.StatusOK)
}

func (s *Server) handleGetPeers(w http.ResponseWriter, r *http.Request) {
	// This is simplified - in production, get from gossip/p2p layer
	peers := []map[string]interface{}{
		{
			"id":         "peer-1",
			"address":    "127.0.0.1:26656",
			"last_seen":  time.Now().Format(time.RFC3339),
			"connected":  true,
		},
	}

	s.respond(w, r, map[string]interface{}{
		"peers": peers,
		"count": len(peers),
	}, http.StatusOK)
}

func (s *Server) handleGetConsensusState(w http.ResponseWriter, r *http.Request) {
	// Get consensus state
	state := map[string]interface{}{
		"height":      0, // In production, get from consensus
		"round":       0,
		"step":        "unknown",
		"proposer":    "unknown",
		"validators":  []string{"node-1"},
		"mempool_size": len(s.consensus.GetMempool()),
	}
	s.respond(w, r, state, http.StatusOK)
}
