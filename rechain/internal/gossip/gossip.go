package gossip

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
)

// GossipProtocol implements epidemic broadcast for metadata synchronization
type GossipProtocol struct {
	host       host.Host
	peers      map[peer.ID]*PeerInfo
	peersMutex sync.RWMutex

	// Message handling
	incoming chan *Message
	outgoing chan *Message

	// CRDT state
	crdtState map[string]interface{}
	stateMutex sync.RWMutex

	// Configuration
	fanout      int           // Number of peers to send to initially
	gossipInterval time.Duration
	antiEntropyInterval time.Duration

	quit chan struct{}
}

// PeerInfo holds information about a connected peer
type PeerInfo struct {
	ID       peer.ID
	LastSeen time.Time
	Score    int // Peer reputation score
}

// Message represents a gossip message
type Message struct {
	ID        string
	Type      MessageType
	Payload   []byte
	Timestamp time.Time
	Sender    peer.ID
	TTL       int // Time to live
}

// MessageType defines the type of gossip message
type MessageType int

const (
	UpdateMessage MessageType = iota
	QueryMessage
	ResponseMessage
	AntiEntropyMessage
)

// NewGossipProtocol creates a new gossip protocol instance
func NewGossipProtocol(listenAddr string) (*GossipProtocol, error) {
	// Create libp2p host
	host, err := libp2p.New(
		libp2p.ListenAddrStrings(listenAddr),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	gp := &GossipProtocol{
		host:       host,
		peers:      make(map[peer.ID]*PeerInfo),
		incoming:   make(chan *Message, 1000),
		outgoing:   make(chan *Message, 1000),
		crdtState:  make(map[string]interface{}),
		fanout:     3,
		gossipInterval: 1 * time.Second,
		antiEntropyInterval: 30 * time.Second,
		quit:       make(chan struct{}),
	}

	// Set up stream handler
	host.SetStreamHandler(protocol.ID("/rechain/gossip/1.0.0"), gp.handleStream)

	// Start background processes
	go gp.processMessages()
	go gp.gossipLoop()
	go gp.antiEntropyLoop()

	log.Printf("Gossip protocol started on %s", host.ID())
	return gp, nil
}

// Start starts the gossip protocol
func (gp *GossipProtocol) Start() error {
	log.Println("Gossip protocol running")
	return nil
}

// Stop stops the gossip protocol
func (gp *GossipProtocol) Stop() error {
	close(gp.quit)
	return gp.host.Close()
}

// AddPeer adds a peer to the known peers list
func (gp *GossipProtocol) AddPeer(peerAddr string) error {
	addr, err := multiaddr.NewMultiaddr(peerAddr)
	if err != nil {
		return fmt.Errorf("invalid peer address: %w", err)
	}

	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return fmt.Errorf("failed to parse peer info: %w", err)
	}

	// Connect to peer
	if err := gp.host.Connect(context.Background(), *peerInfo); err != nil {
		return fmt.Errorf("failed to connect to peer: %w", err)
	}

	gp.peersMutex.Lock()
	gp.peers[peerInfo.ID] = &PeerInfo{
		ID:       peerInfo.ID,
		LastSeen: time.Now(),
		Score:    0,
	}
	gp.peersMutex.Unlock()

	log.Printf("Added peer: %s", peerInfo.ID)
	return nil
}

// Broadcast broadcasts a message to peers
func (gp *GossipProtocol) Broadcast(msgType MessageType, payload []byte) error {
	msg := &Message{
		ID:        generateMessageID(),
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now(),
		Sender:    gp.host.ID(),
		TTL:       10, // Default TTL
	}

	select {
	case gp.outgoing <- msg:
		return nil
	default:
		return fmt.Errorf("outgoing message queue full")
	}
}

// UpdateCRDT updates the local CRDT state and gossips the update
func (gp *GossipProtocol) UpdateCRDT(key string, value interface{}) error {
	gp.stateMutex.Lock()
	gp.crdtState[key] = value
	gp.stateMutex.Unlock()

	// Create update message
	update := map[string]interface{}{
		"key":   key,
		"value": value,
	}
	payload, err := json.Marshal(update)
	if err != nil {
		return err
	}

	return gp.Broadcast(UpdateMessage, payload)
}

// GetCRDT gets a value from the CRDT state
func (gp *GossipProtocol) GetCRDT(key string) (interface{}, bool) {
	gp.stateMutex.RLock()
	defer gp.stateMutex.RUnlock()
	value, exists := gp.crdtState[key]
	return value, exists
}

// QueryCRDT queries for CRDT state from peers
func (gp *GossipProtocol) QueryCRDT(key string) error {
	query := map[string]string{"key": key}
	payload, err := json.Marshal(query)
	if err != nil {
		return err
	}

	return gp.Broadcast(QueryMessage, payload)
}

// gossipLoop periodically gossips recent updates
func (gp *GossipProtocol) gossipLoop() {
	ticker := time.NewTicker(gp.gossipInterval)
	defer ticker.Stop()

	for {
		select {
		case <-gp.quit:
			return
		case <-ticker.C:
			gp.performGossip()
		}
	}
}

// performGossip sends recent updates to random peers
func (gp *GossipProtocol) performGossip() {
	gp.peersMutex.RLock()
	peerIDs := make([]peer.ID, 0, len(gp.peers))
	for id := range gp.peers {
		peerIDs = append(peerIDs, id)
	}
	gp.peersMutex.RUnlock()

	if len(peerIDs) == 0 {
		return
	}

	// Select fanout peers randomly
	selectedPeers := selectRandomPeers(peerIDs, gp.fanout)

	// Send recent state updates
	gp.stateMutex.RLock()
	if len(gp.crdtState) > 0 {
		payload, _ := json.Marshal(gp.crdtState)
		msg := &Message{
			ID:        generateMessageID(),
			Type:      UpdateMessage,
			Payload:   payload,
			Timestamp: time.Now(),
			Sender:    gp.host.ID(),
			TTL:       5,
		}

		for _, peerID := range selectedPeers {
			gp.sendMessage(peerID, msg)
		}
	}
	gp.stateMutex.RUnlock()
}

// antiEntropyLoop performs periodic anti-entropy with random peers
func (gp *GossipProtocol) antiEntropyLoop() {
	ticker := time.NewTicker(gp.antiEntropyInterval)
	defer ticker.Stop()

	for {
		select {
		case <-gp.quit:
			return
		case <-ticker.C:
			gp.performAntiEntropy()
		}
	}
}

// performAntiEntropy performs anti-entropy synchronization
func (gp *GossipProtocol) performAntiEntropy() {
	gp.peersMutex.RLock()
	peerIDs := make([]peer.ID, 0, len(gp.peers))
	for id := range gp.peers {
		peerIDs = append(peerIDs, id)
	}
	gp.peersMutex.RUnlock()

	if len(peerIDs) == 0 {
		return
	}

	// Select one random peer for anti-entropy
	selectedPeer := selectRandomPeers(peerIDs, 1)[0]

	// Send anti-entropy message with current state hash
	gp.stateMutex.RLock()
	stateHash := gp.computeStateHash()
	gp.stateMutex.RUnlock()

	antiEntropyMsg := map[string]string{
		"state_hash": stateHash,
	}
	payload, _ := json.Marshal(antiEntropyMsg)

	msg := &Message{
		ID:        generateMessageID(),
		Type:      AntiEntropyMessage,
		Payload:   payload,
		Timestamp: time.Now(),
		Sender:    gp.host.ID(),
		TTL:       3,
	}

	gp.sendMessage(selectedPeer, msg)
}

// computeStateHash computes a simple hash of the current state
func (gp *GossipProtocol) computeStateHash() string {
	// Simplified - in production, use Merkle tree root
	return fmt.Sprintf("%d", len(gp.crdtState))
}

// processMessages processes incoming messages
func (gp *GossipProtocol) processMessages() {
	for {
		select {
		case <-gp.quit:
			return
		case msg := <-gp.incoming:
			gp.handleMessage(msg)
		}
	}
}

// handleMessage handles an incoming message
func (gp *GossipProtocol) handleMessage(msg *Message) {
	// Update peer last seen
	gp.peersMutex.Lock()
	if peer, exists := gp.peers[msg.Sender]; exists {
		peer.LastSeen = time.Now()
	}
	gp.peersMutex.Unlock()

	switch msg.Type {
	case UpdateMessage:
		gp.handleUpdateMessage(msg)
	case QueryMessage:
		gp.handleQueryMessage(msg)
	case ResponseMessage:
		gp.handleResponseMessage(msg)
	case AntiEntropyMessage:
		gp.handleAntiEntropyMessage(msg)
	}
}

// handleUpdateMessage handles a state update message
func (gp *GossipProtocol) handleUpdateMessage(msg *Message) {
	var update map[string]interface{}
	if err := json.Unmarshal(msg.Payload, &update); err != nil {
		log.Printf("Failed to unmarshal update message: %v", err)
		return
	}

	// Merge update into local state (simplified CRDT merge)
	gp.stateMutex.Lock()
	for key, value := range update {
		gp.crdtState[key] = value
	}
	gp.stateMutex.Unlock()

	log.Printf("Applied update from %s: %v", msg.Sender, update)
}

// handleQueryMessage handles a query message
func (gp *GossipProtocol) handleQueryMessage(msg *Message) {
	var query map[string]string
	if err := json.Unmarshal(msg.Payload, &query); err != nil {
		log.Printf("Failed to unmarshal query message: %v", err)
		return
	}

	key, exists := query["key"]
	if !exists {
		return
	}

	// Get value and send response
	if value, found := gp.GetCRDT(key); found {
		response := map[string]interface{}{
			"key":   key,
			"value": value,
		}
		payload, _ := json.Marshal(response)

		responseMsg := &Message{
			ID:        generateMessageID(),
			Type:      ResponseMessage,
			Payload:   payload,
			Timestamp: time.Now(),
			Sender:    gp.host.ID(),
			TTL:       5,
		}

		gp.sendMessage(msg.Sender, responseMsg)
	}
}

// handleResponseMessage handles a response message
func (gp *GossipProtocol) handleResponseMessage(msg *Message) {
	var response map[string]interface{}
	if err := json.Unmarshal(msg.Payload, &response); err != nil {
		log.Printf("Failed to unmarshal response message: %v", err)
		return
	}

	log.Printf("Received response from %s: %v", msg.Sender, response)
}

// handleAntiEntropyMessage handles an anti-entropy message
func (gp *GossipProtocol) handleAntiEntropyMessage(msg *Message) {
	var antiEntropy map[string]string
	if err := json.Unmarshal(msg.Payload, &antiEntropy); err != nil {
		log.Printf("Failed to unmarshal anti-entropy message: %v", err)
		return
	}

	peerStateHash := antiEntropy["state_hash"]
	localStateHash := gp.computeStateHash()

	if peerStateHash != localStateHash {
		// State differs - send current state for reconciliation
		gp.stateMutex.RLock()
		payload, _ := json.Marshal(gp.crdtState)
		gp.stateMutex.RUnlock()

		reconcileMsg := &Message{
			ID:        generateMessageID(),
			Type:      UpdateMessage,
			Payload:   payload,
			Timestamp: time.Now(),
			Sender:    gp.host.ID(),
			TTL:       3,
		}

		gp.sendMessage(msg.Sender, reconcileMsg)
		log.Printf("Sent state reconciliation to %s", msg.Sender)
	}
}

// handleStream handles incoming streams
func (gp *GossipProtocol) handleStream(s network.Stream) {
	defer s.Close()

	// Read message from stream
	var msg Message
	if err := json.NewDecoder(s).Decode(&msg); err != nil {
		log.Printf("Failed to decode message: %v", err)
		return
	}

	// Add to incoming queue
	select {
	case gp.incoming <- &msg:
	default:
		log.Println("Incoming message queue full, dropping message")
	}
}

// sendMessage sends a message to a specific peer
func (gp *GossipProtocol) sendMessage(peerID peer.ID, msg *Message) {
	s, err := gp.host.NewStream(context.Background(), peerID, protocol.ID("/rechain/gossip/1.0.0"))
	if err != nil {
		log.Printf("Failed to create stream to %s: %v", peerID, err)
		return
	}
	defer s.Close()

	if err := json.NewEncoder(s).Encode(msg); err != nil {
		log.Printf("Failed to send message to %s: %v", peerID, err)
	}
}

// selectRandomPeers selects n random peers from the list
func selectRandomPeers(peers []peer.ID, n int) []peer.ID {
	if len(peers) <= n {
		return peers
	}

	selected := make([]peer.ID, n)
	for i := 0; i < n; i++ {
		randomIndex := make([]byte, 1)
		rand.Read(randomIndex)
		index := int(randomIndex[0]) % len(peers)
		selected[i] = peers[index]
		// Remove selected peer to avoid duplicates
		peers = append(peers[:index], peers[index+1:]...)
	}

	return selected
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}
