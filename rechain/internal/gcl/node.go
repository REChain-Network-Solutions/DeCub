package gcl

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/rechain/rechain/internal/storage"
)

// Node represents a GCL node in the network
type Node struct {
	store  storage.Store
	config *Config

	// Add other node-related fields here
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Config holds the GCL node configuration
type Config struct {
	Port   int
	Seeds  []string
	NodeID string
}

// NewNode creates a new GCL node
func NewNode(store storage.Store) (*Node, error) {
	// Default config
	config := &Config{
		Port:   26656,
		Seeds:  []string{},
		NodeID: "local-node", // In production, this should be a unique identifier
	}

	return &Node{
		store:  store,
		config: config,
	}, nil
}

// Start starts the GCL node
func (n *Node) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	n.cancel = cancel

	// Start the node's main loop
	n.wg.Add(1)
	go n.run(ctx)

	log.Printf("GCL node started on port %d", n.config.Port)
	return nil
}

// Stop gracefully stops the GCL node
func (n *Node) Stop() error {
	if n.cancel != nil {
		n.cancel()
	}

	n.wg.Wait()
	return nil
}

// run is the main event loop for the GCL node
func (n *Node) run(ctx context.Context) {
	defer n.wg.Done()

	// Start the P2P server
	p2pServer, err := NewP2PServer(n.config)
	if err != nil {
		log.Printf("Failed to start P2P server: %v", err)
		return
	}

	// Start the consensus service
	consensus, err := NewConsensus(n.store, p2pServer)
	if err != nil {
		log.Printf("Failed to start consensus: %v", err)
		return
	}

	// Start the API server
	apiServer := NewAPIServer(consensus, n.store)
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()

	// Wait for shutdown
	<-ctx.Done()

	// Shutdown sequence
	if err := apiServer.Stop(); err != nil {
		log.Printf("Error stopping API server: %v", err)
	}

	if err := consensus.Stop(); err != nil {
		log.Printf("Error stopping consensus: %v", err)
	}

	if err := p2pServer.Stop(); err != nil {
		log.Printf("Error stopping P2P server: %v", err)
	}
}
