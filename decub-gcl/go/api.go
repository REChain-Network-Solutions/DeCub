package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var (
	ledger   []Block
	ledgerMu sync.RWMutex
	cons     *Consensus
)

// SubmitTx handles POST /gcl/tx
func SubmitTx(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tx Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Simulate adding to pending txs, for simplicity add to new block immediately
	ledgerMu.Lock()
	height := len(ledger) + 1
	var prevHash string
	if height > 1 {
		prevHash = HashBlock(ledger[height-2])
	}
	block := cons.ProposeBlock(height, prevHash, []Transaction{tx}, "validator1")
	sigs := cons.SignBlock(block)
	if cons.VerifyQuorum(sigs) {
		ledger = append(ledger, block)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Transaction submitted, block %d created", height)
	} else {
		http.Error(w, "Consensus failed", http.StatusInternalServerError)
	}
	ledgerMu.Unlock()
}

// GetBlock handles GET /gcl/block/{height}
func GetBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/gcl/block/")
	height, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid height", http.StatusBadRequest)
		return
	}

	ledgerMu.RLock()
	if height < 1 || height > len(ledger) {
		ledgerMu.RUnlock()
		http.Error(w, "Block not found", http.StatusNotFound)
		return
	}
	block := ledger[height-1]
	ledgerMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(block)
}

// GetProof handles GET /gcl/proof/{tx_id}
func GetProof(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/gcl/proof/")
	txID := path

	ledgerMu.RLock()
	defer ledgerMu.RUnlock()

	for _, block := range ledger {
		for i, tx := range block.Txs {
			if tx.TxID == txID {
				root, _ := BuildMerkleTree(block.Txs)
				proof := GenerateMerkleProof(root, i)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(proof)
				return
			}
		}
	}

	http.Error(w, "Transaction not found", http.StatusNotFound)
}
