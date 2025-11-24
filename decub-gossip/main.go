package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"github.com/syndtr/goleveldb/leveldb"
)

// LWWRegister represents a Last-Write-Wins register for CRDT
type LWWRegister struct {
	value     interface{}
	timestamp int64
}

// NewLWWRegister creates a new LWW register
func NewLWWRegister(value interface{}) *LWWRegister {
	return &LWWRegister{
		value:     value,
		timestamp: time.Now().UnixNano(),
	}
}

// Set sets the value with current timestamp
func (r *LWWRegister) Set(value interface{}) {
	r.value = value
	r.timestamp = time.Now().UnixNano()
}

// Get returns the current value
func (r *LWWRegister) Get() interface{} {
	return r.value
}

// Merge merges with another register using LWW rule
func (r *LWWRegister) Merge(other *LWWRegister) {
	if other.timestamp > r.timestamp {
		r.value = other.value
		r.timestamp = other.timestamp
	}
}

// Delta represents a CRDT delta for gossip
type Delta struct {
	NodeID      string                 `json:"node_id"`
	VectorClock map[string]int64       `json:"vector_clock"`
	Type        string                 `json:"type"` // "orset" or "lww"
	Key         string                 `json:"key"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   int64                  `json:"timestamp"`
}

// GossipNode represents a gossip node for catalog synchronization
type GossipNode struct {
	host        host.Host
	pubsub      *pubsub.PubSub
	catalog     *CatalogCRDT // Integrated CRDT catalog
	db          *leveldb.DB
	merkleTree  *CatalogMerkleTree
	config      *GossipConfig
	catalogAddr string
	merkleRoot  string
	mu          sync.RWMutex
}

// GossipConfig holds configuration for gossip synchronization
type GossipConfig struct {
	NodeID              string        `json:"node_id"`
	ListenAddr          string        `json:"listen_addr"`
	InitialPeers        []string      `json:"initial_peers"`
	SyncInterval        time.Duration `json:"sync_interval"`
	AntiEntropyInterval time.Duration `json:"anti_entropy_interval"`
	CatalogAddr         string        `json:"catalog_addr"`
}

// NewGossipConfig creates default gossip configuration
func NewGossipConfig() *GossipConfig {
	return &GossipConfig{
		NodeID:              "node-" + fmt.Sprintf("%d", time.Now().Unix()),
		ListenAddr:          "/ip4/0.0.0.0/tcp/0",
		InitialPeers:        []string{},
		SyncInterval:        10 * time.Second,
		AntiEntropyInterval: 30 * time.Second,
		CatalogAddr:         "http://localhost:8080",
	}
}

// CatalogCRDT represents the CRDT-backed catalog (simplified interface)
type CatalogCRDT struct {
	nodeID      string
	vectorClock map[string]int64
	snapshots   map[string]*LWWRegister
	images      map[string]*LWWRegister
	deltas      []*Delta
	mu          sync.RWMutex
}

// NewCatalogCRDT creates a new catalog CRDT
func NewCatalogCRDT(nodeID string) *CatalogCRDT {
	return &CatalogCRDT{
		nodeID:      nodeID,
		vectorClock: make(map[string]int64),
		snapshots:   make(map[string]*LWWRegister),
		images:      make(map[string]*LWWRegister),
		deltas:      make([]*Delta, 0),
	}
}

// AddSnapshot adds a snapshot to the catalog
func (c *CatalogCRDT) AddSnapshot(id string, metadata map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.snapshots[id] = NewLWWRegister(metadata)
	c.vectorClock[c.nodeID]++

	delta := &Delta{
		NodeID:      c.nodeID,
		VectorClock: c.vectorClock,
		Type:        "lww",
		Key:         "snapshots:" + id,
		Data:        map[string]interface{}{"metadata": metadata},
		Timestamp:   time.Now().UnixNano(),
	}
	c.deltas = append(c.deltas, delta)
}

// GetDeltas returns pending deltas
func (c *CatalogCRDT) GetDeltas() []*Delta {
	c.mu.RLock()
	defer c.mu.RUnlock()

	deltas := make([]*Delta, len(c.deltas))
	copy(deltas, c.deltas)
	return deltas
}

// ApplyDelta applies a received delta
func (c *CatalogCRDT) ApplyDelta(delta *Delta) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simple causal ordering check
	if delta.VectorClock[delta.NodeID] <= c.vectorClock[delta.NodeID] {
		return false // Already applied
	}

	// Update vector clock
	for node, time := range delta.VectorClock {
		if time > c.vectorClock[node] {
			c.vectorClock[node] = time
		}
	}

	// Apply delta based on type
	switch delta.Type {
	case "lww":
		if strings.HasPrefix(delta.Key, "snapshots:") {
			id := strings.TrimPrefix(delta.Key, "snapshots:")
			if metadata, ok := delta.Data["metadata"].(map[string]interface{}); ok {
				if existing, exists := c.snapshots[id]; exists {
					existing.Merge(NewLWWRegister(metadata))
				} else {
					c.snapshots[id] = NewLWWRegister(metadata)
				}
			}
		}
	}

	return true
}

// ClearDeltas clears processed deltas
func (c *CatalogCRDT) ClearDeltas() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deltas = c.deltas[:0]
}

// GetState returns the current catalog state for Merkle calculation
func (c *CatalogCRDT) GetState() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	state := make(map[string]interface{})
	for id, reg := range c.snapshots {
		state["snapshot:"+id] = reg.Get()
	}
	for id, reg := range c.images {
		state["image:"+id] = reg.Get()
	}
	return state
}

// MerkleNode represents a node in the Merkle tree
type MerkleNode struct {
	Hash  string
	Left  *MerkleNode
	Right *MerkleNode
}

// BuildMerkleTree builds a Merkle tree from data
func BuildMerkleTree(data []string) *MerkleNode {
	if len(data) == 0 {
		return nil
	}

	nodes := make([]*MerkleNode, len(data))
	for i, d := range data {
		hash := sha256.Sum256([]byte(d))
		nodes[i] = &MerkleNode{Hash: hex.EncodeToString(hash[:])}
	}

	for len(nodes) > 1 {
		var newNodes []*MerkleNode
		for i := 0; i < len(nodes); i += 2 {
			left := nodes[i]
			var right *MerkleNode
			if i+1 < len(nodes) {
				right = nodes[i+1]
			} else {
				right = left // Duplicate for odd number
			}
			hash := sha256.Sum256([]byte(left.Hash + right.Hash))
			newNodes = append(newNodes, &MerkleNode{Hash: hex.EncodeToString(hash[:]), Left: left, Right: right})
		}
		nodes = newNodes
	}

	return nodes[0]
}

// NewGossipNode creates a new gossip node
func NewGossipNode(config *GossipConfig) (*GossipNode, error) {
	// Generate a new private key
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		return nil, err
	}

	// Create libp2p host
	host, err := libp2p.New(
		libp2p.ListenAddrStrings(config.ListenAddr),
		libp2p.Identity(priv),
	)
	if err != nil {
		return nil, err
	}

	// Create pubsub
	ps, err := pubsub.NewGossipSub(context.Background(), host)
	if err != nil {
		return nil, err
	}

	// Open LevelDB
	db, err := leveldb.OpenFile("./gossip.db", nil)
	if err != nil {
		return nil, err
	}

	catalog := NewCatalogCRDT(config.NodeID)
	merkleTree := NewCatalogMerkleTree()

	node := &GossipNode{
		host:        host,
		pubsub:      ps,
		catalog:     catalog,
		db:          db,
		merkleTree:  merkleTree,
		config:      config,
		catalogAddr: config.CatalogAddr,
	}

	// Subscribe to topics
	node.subscribeToTopics()

	// Connect to initial peers
	for _, peerAddr := range config.InitialPeers {
		go func(addr string) {
			if err := node.Connect(addr); err != nil {
				log.Printf("Failed to connect to initial peer %s: %v", addr, err)
			}
		}(peerAddr)
	}

	return node, nil
}

// subscribeToTopics subscribes to gossip topics
func (n *GossipNode) subscribeToTopics() {
	// Delta sync topic
	deltaTopic, err := n.pubsub.Join("decub/delta")
	if err != nil {
		log.Printf("Failed to join delta topic: %v", err)
		return
	}

	sub, err := deltaTopic.Subscribe()
	if err != nil {
		log.Printf("Failed to subscribe to delta: %v", err)
		return
	}

	go n.handleDeltas(sub)

	// Anti-entropy topic
	antiEntropyTopic, err := n.pubsub.Join("decub/anti-entropy")
	if err != nil {
		log.Printf("Failed to join anti-entropy topic: %v", err)
		return
	}

	subAE, err := antiEntropyTopic.Subscribe()
	if err != nil {
		log.Printf("Failed to subscribe to anti-entropy: %v", err)
		return
	}

	go n.handleAntiEntropy(subAE)
}

// handleDeltas handles incoming delta messages
func (n *GossipNode) handleDeltas(sub *pubsub.Subscription) {
	ticker := time.NewTicker(n.config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send pending deltas
			deltas := n.catalog.GetDeltas()
			if len(deltas) > 0 {
				data, _ := json.Marshal(deltas)
				n.publish("decub/delta", data)
			}

		default:
			msg, err := sub.Next(context.Background())
			if err != nil {
				log.Printf("Delta subscription error: %v", err)
				continue
			}

			if msg.ReceivedFrom == n.host.ID() {
				continue // Ignore own messages
			}

			var deltas []*Delta
			if err := json.Unmarshal(msg.Data, &deltas); err != nil {
				log.Printf("Failed to unmarshal deltas: %v", err)
				continue
			}

			// Apply received deltas
			for _, delta := range deltas {
				applied := n.catalog.ApplyDelta(delta)
				if applied {
					log.Printf("Applied delta: %s (%s)", delta.Key, delta.Type)
				}
			}

			// Clear processed deltas
			n.catalog.ClearDeltas()
		}
	}
}

// handleAntiEntropy handles anti-entropy messages
func (n *GossipNode) handleAntiEntropy(sub *pubsub.Subscription) {
	ticker := time.NewTicker(n.config.AntiEntropyInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send Merkle root for anti-entropy
			state := n.catalog.GetState()
			stateData, _ := json.Marshal(state)
			root := BuildMerkleTree([]string{string(stateData)})
			if root != nil {
				n.merkleRoot = root.Hash
				data, _ := json.Marshal(map[string]string{"merkle_root": root.Hash})
				n.publish("decub/anti-entropy", data)
			}

		default:
			msg, err := sub.Next(context.Background())
			if err != nil {
				log.Printf("Anti-entropy subscription error: %v", err)
				continue
			}

			if msg.ReceivedFrom == n.host.ID() {
				continue
			}

			var aeMsg map[string]interface{}
			if err := json.Unmarshal(msg.Data, &aeMsg); err != nil {
				log.Printf("Failed to unmarshal anti-entropy: %v", err)
				continue
			}

			// Check if it's a Merkle root message
			if merkleRoot, ok := aeMsg["merkle_root"].(string); ok {
				if merkleRoot != n.merkleRoot {
					log.Printf("Merkle root mismatch detected, requesting full sync")
					// Request full state sync
					n.publish("decub/anti-entropy", []byte(`{"sync_request": true}`))
				}
			}

			// Check if it's a sync request or full state
			if _, ok := aeMsg["sync_request"]; ok {
				// Send full state
				state := n.catalog.GetState()
				data, _ := json.Marshal(state)
				n.publish("decub/anti-entropy", data)
			} else if _, ok := aeMsg["snapshot:snap1"]; ok {
				// Received full state, apply it
				for key, value := range aeMsg {
					if strings.HasPrefix(key, "snapshot:") {
						id := strings.TrimPrefix(key, "snapshot:")
						if metadata, ok := value.(map[string]interface{}); ok {
							n.catalog.snapshots[id] = metadata
						}
					}
				}
				log.Printf("Applied full state sync")
			}
		}
	}
}

// publish publishes a message to a topic
func (n *GossipNode) publish(topic string, data []byte) {
	t, err := n.pubsub.Join(topic)
	if err != nil {
		log.Printf("Failed to join topic %s: %v", topic, err)
		return
	}

	if err := t.Publish(context.Background(), data); err != nil {
		log.Printf("Failed to publish to %s: %v", topic, err)
	}
}

// Connect connects to a peer
func (n *GossipNode) Connect(addr string) error {
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return err
	}

	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return err
	}

	return n.host.Connect(context.Background(), *info)
}

// GetStatus returns gossip node status
func (n *GossipNode) GetStatus() map[string]interface{} {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return map[string]interface{}{
		"node_id":       n.catalog.nodeID,
		"merkle_root":  n.merkleRoot,
		"peers":        len(n.host.Peerstore().Peers()),
		"snapshots":    len(n.catalog.snapshots),
		"pending_deltas": len(n.catalog.deltas),
	}
}

// Close closes the gossip node
func (n *GossipNode) Close() error {
	n.db.Close()
	return n.host.Close()
}

func main() {
	config := LoadConfigFromEnv()

	// Override with command line args if provided
	if len(os.Args) > 1 {
		config.ListenAddr = os.Args[1]
	}
	if len(os.Args) > 2 {
		config.NodeID = os.Args[2]
	}
	if len(os.Args) > 3 {
		config.InitialPeers = []string{os.Args[3]}
	}

	if err := config.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	node, err := NewGossipNode(config)
	if err != nil {
		log.Fatalf("Failed to create gossip node: %v", err)
	}
	defer node.Close()

	fmt.Printf("Gossip node started at %s (Node ID: %s)\n", node.host.Addrs()[0], config.NodeID)

	// Start periodic Merkle root broadcasting
	go node.startMerkleBroadcast()

	// Add some test data
	node.catalog.AddSnapshot("test-snap", map[string]interface{}{
		"size": 1024,
		"cluster": "test",
	})

	// Keep running
	select {}
}

// startMerkleBroadcast periodically updates and broadcasts Merkle root
func (n *GossipNode) startMerkleBroadcast() {
	ticker := time.NewTicker(n.config.AntiEntropyInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Update Merkle tree from catalog state
			state := n.catalog.GetState()
			if len(state) > 0 {
				err := n.merkleTree.BuildFromCatalog(n.catalog.snapshots, n.catalog.images)
				if err != nil {
					log.Printf("Failed to build Merkle tree: %v", err)
					continue
				}

				rootHash := n.merkleTree.GetRootHash()
				if rootHash != "" {
					data, _ := json.Marshal(map[string]string{"merkle_root": rootHash})
					n.publish("decub/anti-entropy", data)
					log.Printf("Broadcasted Merkle root: %s", rootHash[:8]+"...")
				}
			}
		}
	}
}
