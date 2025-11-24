package consensus

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rechain/rechain/internal/gcl"
	"github.com/rechain/rechain/internal/storage"
)

// Consensus implements the BFT consensus algorithm (Tendermint-style)
type Consensus struct {
	store     storage.Store
	p2p       *gcl.P2PServer
	config    *Config
	proposals chan *Proposal
	blocks    chan *Block
	quit      chan struct{}

	height    uint64
	round     int32
	step      Step
	locked    *Block
	validated *Block

	voted map[uint32]bool // Track votes in current round
	votes  []*Vote        // Collected votes for current round/step

	votingMutex sync.Mutex

	// Validator set (simplified - all connected peers are validators)
	validators []string
	validatorIndex int

	// Timing
	timeoutPrevote   time.Duration
	timeoutPrecommit time.Duration
	timeoutCommit    time.Duration

	// Mempool for transactions
	mempool []*Transaction
}

// Step represents the current step in the consensus round
type Step int

const (
	Propose Step = iota
	Prevote
	Precommit
	Commit
)

// Config holds consensus configuration
type Config struct {
	NodeID        string
	BlockInterval time.Duration
	Timeout       time.Duration
}

// Transaction represents a transaction to be included in a block
type Transaction struct {
	ID        string
	Type      string
	Payload   []byte
	Timestamp time.Time
	Sender    string
	Signature []byte
}

// NewConsensus creates a new consensus instance
func NewConsensus(store storage.Store, p2p *gcl.P2PServer) (*Consensus, error) {
	c := &Consensus{
		store:     store,
		p2p:       p2p,
		proposals: make(chan *Proposal, 100),
		blocks:    make(chan *Block, 100),
		quit:      make(chan struct{}),
		voted:     make(map[uint32]bool),
		config: &Config{
			BlockInterval: 1 * time.Second,
			Timeout:       5 * time.Second,
		},
		timeoutPrevote:   3 * time.Second,
		timeoutPrecommit: 3 * time.Second,
		timeoutCommit:    1 * time.Second,
		validators:       []string{"node-1"}, // Simplified
		validatorIndex:   0,
		mempool:          make([]*Transaction, 0),
	}

	// Start the consensus loop
	go c.run()

	return c, nil
}

// Start starts the consensus process
func (c *Consensus) Start() error {
	log.Println("Consensus engine started")
	return nil
}

// Stop stops the consensus process
func (c *Consensus) Stop() error {
	close(c.quit)
	return nil
}

// AddTransaction adds a transaction to the mempool
func (c *Consensus) AddTransaction(tx *Transaction) {
	c.votingMutex.Lock()
	defer c.votingMutex.Unlock()
	c.mempool = append(c.mempool, tx)
	log.Printf("Added transaction %s to mempool", tx.ID)
}

// GetMempool returns current transactions in mempool
func (c *Consensus) GetMempool() []*Transaction {
	c.votingMutex.Lock()
	defer c.votingMutex.Unlock()
	return append([]*Transaction{}, c.mempool...)
}

// Propose proposes a new block
func (c *Consensus) Propose(block *Block) error {
	c.proposals <- &Proposal{
		Block: block,
		Round: c.round,
	}
	return nil
}

// run is the main consensus loop
func (c *Consensus) run() {
	ticker := time.NewTicker(c.config.BlockInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.quit:
			return

		case <-ticker.C:
			// Start new height
			c.startNewHeight()

		case prop := <-c.proposals:
			c.handleProposal(prop)

		case block := <-c.blocks:
			c.handleBlock(block)
		}
	}
}

// startNewHeight starts a new consensus height
func (c *Consensus) startNewHeight() {
	c.votingMutex.Lock()
	defer c.votingMutex.Unlock()

	c.height++
	c.round = 0
	c.step = Propose
	c.locked = nil
	c.validated = nil
	c.voted = make(map[uint32]bool)
	c.votes = nil

	log.Printf("Starting new height: %d", c.height)

	// If we're the proposer for this round, propose a new block
	if c.isProposer() {
		block := c.createProposal()
		c.Propose(block)
	}

	// Start timeout for propose step
	go c.startTimeout(Propose, c.timeoutPrevote)
}

// isProposer checks if the current node is the proposer for the current round
func (c *Consensus) isProposer() bool {
	// Simple round-robin proposer selection
	proposerIndex := (int(c.height) + int(c.round)) % len(c.validators)
	return c.validators[proposerIndex] == c.config.NodeID
}

// createProposal creates a new block proposal
func (c *Consensus) createProposal() *Block {
	// Get transactions from mempool
	txs := c.GetMempool()

	// Create a new block with transactions
	block := &Block{
		Height:    c.height,
		Round:     c.round,
		Timestamp: time.Now(),
		Txs:       make([][]byte, len(txs)),
		LastHash:  c.getLastBlockHash(),
		StateHash: c.getStateHash(),
	}

	// Serialize transactions
	for i, tx := range txs {
		txBytes, _ := json.Marshal(tx)
		block.Txs[i] = txBytes
	}

	log.Printf("Created proposal for height %d with %d transactions", c.height, len(txs))
	return block
}

// getLastBlockHash returns the hash of the last committed block
func (c *Consensus) getLastBlockHash() []byte {
	// Simplified - in production, get from committed blocks
	if c.height == 1 {
		return make([]byte, 32) // Genesis block hash
	}
	// Get from storage
	key := []byte(fmt.Sprintf("block/%d", c.height-1))
	hash, _ := c.store.Get(context.Background(), key)
	if hash == nil {
		return make([]byte, 32)
	}
	return hash
}

// getStateHash returns the current state hash
func (c *Consensus) getStateHash() []byte {
	// Simplified - in production, get from MerkleStore
	hash := sha256.Sum256([]byte(fmt.Sprintf("state-%d", c.height)))
	return hash[:]
}

// startTimeout starts a timeout for the given step
func (c *Consensus) startTimeout(step Step, duration time.Duration) {
	time.Sleep(duration)

	c.votingMutex.Lock()
	defer c.votingMutex.Unlock()

	if c.step == step {
		log.Printf("Timeout for step %v at height %d, round %d", step, c.height, c.round)
		c.advanceToNextStep()
	}
}

// handleProposal handles a new proposal
func (c *Consensus) handleProposal(proposal *Proposal) {
	// Validate the proposal
	if !c.validateProposal(proposal) {
		log.Printf("Invalid proposal for height %d", proposal.Block.Height)
		return
	}

	log.Printf("Received valid proposal for height %d", proposal.Block.Height)

	// Move to prevote step
	c.step = Prevote

	// Send prevote
	vote := &Vote{
		Height:   proposal.Block.Height,
		Round:    proposal.Block.Round,
		Type:     Prevote,
		BlockID:  proposal.Block.Hash(),
		SenderID: c.config.NodeID,
	}

	c.broadcastVote(vote)
}

// validateProposal validates a proposal
func (c *Consensus) validateProposal(proposal *Proposal) bool {
	// Check height and round
	if proposal.Block.Height != c.height || proposal.Block.Round != c.round {
		return false
	}

	// Check proposer
	if !c.isProposer() {
		return false // Only proposer can send proposals in this simplified version
	}

	// Validate transactions (simplified)
	for _, txBytes := range proposal.Block.Txs {
		var tx Transaction
		if err := json.Unmarshal(txBytes, &tx); err != nil {
			return false
		}
		if !c.validateTransaction(&tx) {
			return false
		}
	}

	return true
}

// validateTransaction validates a transaction
func (c *Consensus) validateTransaction(tx *Transaction) bool {
	// Simplified validation - check signature, etc.
	return len(tx.ID) > 0 && len(tx.Sender) > 0
}

// broadcastVote broadcasts a vote to all peers
func (c *Consensus) broadcastVote(vote *Vote) {
	// Serialize vote
	voteBytes, err := json.Marshal(vote)
	if err != nil {
		log.Printf("Failed to serialize vote: %v", err)
		return
	}

	// Broadcast via P2P (simplified - in production, use proper message types)
	log.Printf("Broadcasting %s vote for height %d, round %d", vote.Type, vote.Height, vote.Round)

	// For now, just log - in production, send to all validators
	_ = voteBytes
}

// handleBlock handles a new block
func (c *Consensus) handleBlock(block *Block) {
	// Validate and apply block
	if c.validateBlock(block) {
		c.commitBlock(block)
	}
}

// validateBlock validates a block
func (c *Consensus) validateBlock(block *Block) bool {
	// Simplified validation
	return block.Height == c.height && len(block.Txs) >= 0
}

// commitBlock commits a block to the blockchain
func (c *Consensus) commitBlock(block *Block) {
	log.Printf("Committing block at height %d", block.Height)

	// Store block
	blockBytes, _ := json.Marshal(block)
	blockKey := []byte(fmt.Sprintf("block/%d", block.Height))
	c.store.Set(context.Background(), blockKey, blockBytes)

	// Store block hash
	hashKey := []byte(fmt.Sprintf("block-hash/%d", block.Height))
	c.store.Set(context.Background(), hashKey, block.Hash())

	// Clear mempool (transactions are now in block)
	c.votingMutex.Lock()
	c.mempool = nil
	c.votingMutex.Unlock()

	// Move to next height
	c.height++
}

// advanceToNextStep advances to the next consensus step
func (c *Consensus) advanceToNextStep() {
	switch c.step {
	case Propose:
		c.step = Prevote
		go c.startTimeout(Prevote, c.timeoutPrevote)
	case Prevote:
		c.step = Precommit
		go c.startTimeout(Precommit, c.timeoutPrecommit)
	case Precommit:
		c.step = Commit
		go c.startTimeout(Commit, c.timeoutCommit)
	case Commit:
		// Start new height
		c.startNewHeight()
	}
}

// Block represents a block in the blockchain
type Block struct {
	Height    uint64
	Round     int32
	Timestamp time.Time
	Txs       [][]byte
	LastHash  []byte
	StateHash []byte
}

// Hash returns the hash of the block
func (b *Block) Hash() []byte {
	h := sha256.New()
	binary.Write(h, binary.BigEndian, b.Height)
	binary.Write(h, binary.BigEndian, b.Round)
	h.Write(b.LastHash)
	h.Write(b.StateHash)
	for _, tx := range b.Txs {
		h.Write(tx)
	}
	return h.Sum(nil)
}

// Proposal represents a block proposal
type Proposal struct {
	Block *Block
	Round int32
}

// Vote represents a vote in the consensus process
type Vote struct {
	Height   uint64
	Round    int32
	Type     VoteType
	BlockID  []byte
	SenderID string
}

// VoteType represents the type of vote
type VoteType int

const (
	Prevote VoteType = iota
	Precommit
)

// String returns string representation of VoteType
func (vt VoteType) String() string {
	switch vt {
	case Prevote:
		return "Prevote"
	case Precommit:
		return "Precommit"
	default:
		return "Unknown"
	}
}
