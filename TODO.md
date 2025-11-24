# Gossip Synchronization Layer Implementation TODO

## Phase 1: Enhance Gossip Layer
- [ ] Modify `decub-gossip/main.go` to integrate with CRDT catalog deltas
- [ ] Replace simple key-value CRDT with catalog delta exchange
- [ ] Add delta serialization/deserialization for gossip messages

## Phase 2: Merkle Root Maintenance
- [ ] Create `decub-gossip/merkle.go` for catalog Merkle tree
- [ ] Implement Merkle root calculation from catalog state
- [ ] Add periodic Merkle root broadcasting

## Phase 3: Anti-Entropy Sync
- [ ] Add Merkle root comparison logic
- [ ] Implement full state sync when roots mismatch
- [ ] Add sync request/response handling

## Phase 4: Configuration
- [ ] Create `decub-gossip/config.go` with peer list and intervals
- [ ] Add environment variable support
- [ ] Implement config validation

## Phase 5: CLI Tool
- [ ] Create `cmd/decubectl/main.go` with gossip commands
- [ ] Implement `decubectl gossip status` command
- [ ] Implement `decubectl gossip sync` command

## Phase 6: Integration
- [ ] Update `decub-catalog/crdt_catalog.go` to include gossip client
- [ ] Add gossip initialization to catalog service
- [ ] Test end-to-end synchronization

## Testing
- [ ] Test gossip synchronization between two nodes
- [ ] Verify anti-entropy sync on divergence
- [ ] Test CLI commands functionality
