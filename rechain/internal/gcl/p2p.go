package gcl

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

// P2PServer handles the peer-to-peer communication
type P2PServer struct {
	node    *enode.Node
	server  *p2p.Server
	config  *Config
	privKey *ecdsa.PrivateKey

	peers     map[enode.ID]*p2p.Peer
	peersLock sync.RWMutex
}

// NewP2PServer creates a new P2P server
func NewP2PServer(config *Config) (*P2PServer, error) {
	// Generate a new private key for the node
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	srv := &P2PServer{
		config:  config,
		privKey: privKey,
		peers:   make(map[enode.ID]*p2p.Peer),
	}

	// Create the P2P server configuration
	serverConfig := p2p.Config{
		PrivateKey:      privKey,
		Name:            "rechain-gcl",
		ListenAddr:      fmt.Sprintf(":%d", config.Port),
		Protocols:       srv.makeProtocols(),
		NodeDatabase:    "", // In-memory node database for now
		BootstrapNodes:  []*enode.Node{},
		StaticNodes:     []*enode.Node{},
		TrustedNodes:    []*enode.Node{},
		NetRestrict:     nil,
		NoDiscovery:     false,
		DialRatio:       3,
		MaxPeers:        50,
		MaxPendingPeers: 50,
	}

	// Add bootstrap nodes if any
	for _, seed := range config.Seeds {
		node, err := enode.Parse(enode.ValidSchemes, seed)
		if err != nil {
			log.Printf("Failed to parse seed node %s: %v", seed, err)
			continue
		}
		serverConfig.BootstrapNodes = append(serverConfig.BootstrapNodes, node)
	}

	srv.server = &p2p.Server{Config: serverConfig}

	return srv, nil
}

// Start starts the P2P server
func (s *P2PServer) Start() error {
	if err := s.server.Start(); err != nil {
		return fmt.Errorf("failed to start P2P server: %w", err)
	}

	// Start the discovery service
	s.server.AddPeer = func(node *enode.Node) {
		log.Printf("New peer connected: %s", node)
	}

	s.server.DelPeer = func(node *enode.Node) {
		log.Printf("Peer disconnected: %s", node)
	}

	log.Printf("P2P server started, node ID: %s", s.server.Self())
	return nil
}

// Stop stops the P2P server
func (s *P2PServer) Stop() error {
	s.server.Stop()
	return nil
}

// makeProtocols creates the supported protocols
func (s *P2PServer) makeProtocols() []p2p.Protocol {
	return []p2p.Protocol{
		{
			Name:    "rechain",
			Version: 1,
			Length:  16,
			Run:     s.handlePeer,
		},
	}
}

// handlePeer handles a connected peer
func (s *P2PServer) handlePeer(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
	// Add to peer list
	s.peersLock.Lock()
	s.peers[peer.ID()] = peer
	s.peersLock.Unlock()

	// Remove from peer list when done
	defer func() {
		s.peersLock.Lock()
		delete(s.peers, peer.ID())
		s.peersLock.Unlock()
	}()

	// Main message loop
	for {
		msg, err := rw.ReadMsg()
		if err != nil {
			return err
		}

		// Handle the message
		switch msg.Code {
		// Add message handlers here
		default:
			log.Printf("Unknown message code: %d", msg.Code)
		}

		msg.Discard()
	}
}

// Broadcast sends a message to all connected peers
func (s *P2PServer) Broadcast(msg p2p.Msg) error {
	s.peersLock.RLock()
	defer s.peersLock.RUnlock()

	for _, peer := range s.peers {
		if err := p2p.Send(peer.RW(), msg.Code, msg.Payload); err != nil {
			log.Printf("Failed to send message to peer %s: %v", peer.ID(), err)
		}
	}

	return nil
}
