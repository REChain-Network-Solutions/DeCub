# DeCube System Test Plan

## Overview

This test plan covers comprehensive testing of the DeCube distributed system, including CRDT-based catalog synchronization, gossip protocols, snapshot lifecycle management, and consensus mechanisms. The plan includes unit, integration, and chaos testing scenarios.

## 1. Unit Tests

### Merkle Proofs
- **Test Merkle Tree Construction**
  - Verify correct root hash calculation for various data sizes
  - Test with empty data, single leaf, and large datasets
  - Validate tree structure and hash propagation

- **Test Merkle Proof Generation**
  - Generate proofs for different leaf indices
  - Verify proof validation against root hash
  - Test proof size and efficiency

- **Test Merkle Proof Verification**
  - Valid proofs should verify successfully
  - Invalid proofs (tampered data, wrong indices) should fail
  - Test edge cases: first leaf, last leaf, middle leaves

### CRDT Merge Operations
- **OR-Set Operations**
  - Test add operations with unique tags
  - Test remove operations and conflict resolution
  - Verify concurrent add/remove scenarios

- **LWW-Register Operations**
  - Test timestamp-based conflict resolution
  - Verify later writes win semantics
  - Test merge operations with different node IDs

- **Vector Clock Operations**
  - Test increment operations
  - Test merge operations for concurrent events
  - Verify causality comparison logic

### Encryption and Security
- **Key Generation and Management**
  - Test cryptographic key pair generation
  - Verify key rotation procedures
  - Test key persistence and recovery

- **Data Encryption/Decryption**
  - Test symmetric encryption of data chunks
  - Verify decryption with correct/incorrect keys
  - Test encryption performance with large data

- **Digital Signatures**
  - Test transaction signing and verification
  - Verify signature validation for tampered data
  - Test multi-signature scenarios

## 2. Integration Tests

### Snapshot Lifecycle (snapshot→upload→register→restore)

#### End-to-End Snapshot Creation
- **Create Test Environment**
  - Set up mock etcd instance with test data
  - Create test volume data structure
  - Configure object store and GCL endpoints

- **Snapshot Creation Phase**
  - Execute `decub-snapshot create` command
  - Verify etcd snapshot creation
  - Verify volume data compression
  - Validate combined snapshot file integrity

- **Upload Phase**
  - Test chunking into 64MB pieces
  - Verify SHA256 hash calculation for each chunk
  - Test upload to object store with progress tracking
  - Validate uploaded chunk integrity

- **Registration Phase**
  - Test metadata extraction from snapshot
  - Verify GCL transaction creation
  - Test transaction submission and confirmation
  - Validate metadata storage in catalog

- **Restore Phase**
  - Test metadata retrieval from GCL
  - Verify chunk download with hash verification
  - Test snapshot reconstruction from chunks
  - Validate restore to target directories

#### Cross-Component Integration
- **Catalog-Gossip Integration**
  - Test delta generation from catalog changes
  - Verify gossip message broadcasting
  - Test delta application on receiving nodes
  - Validate eventual consistency across nodes

- **Consensus Integration**
  - Test hybrid 2PC with RAFT and BFT
  - Verify prepare/commit phases
  - Test failure scenarios and recovery

## 3. Chaos Tests

### Network Partition Simulation
- **Partition Scenarios**
  - Split network into isolated segments
  - Test gossip message delivery during partition
  - Verify anti-entropy sync after partition healing
  - Test Merkle root comparison and full sync

- **Expected Behavior**
  - Nodes in different partitions continue independent operation
  - Upon reconnection, Merkle root mismatch triggers sync
  - Full state transfer for diverged partitions
  - Eventual consistency restoration

### Validator Loss Simulation
- **Single Validator Failure**
  - Simulate validator node crash/failure
  - Test consensus continuation with remaining validators
  - Verify transaction processing during failure
  - Test recovery when validator returns

- **Multiple Validator Failures**
  - Test BFT tolerance up to f faulty nodes (f = (n-1)/3)
  - Verify system halts when too many validators fail
  - Test recovery procedures for failed validators

- **Expected Behavior**
  - Consensus continues if < 1/3 validators fail
  - Failed validators are detected and replaced
  - Transaction finality maintained
  - State recovery from surviving validators

### Corrupted Chunk Simulation
- **Upload Corruption**
  - Simulate chunk corruption during upload
  - Test hash verification failure detection
  - Verify automatic retry mechanisms
  - Test corruption handling in object store

- **Download Corruption**
  - Simulate chunk corruption in object store
  - Test hash verification during download
  - Verify chunk redownload on corruption detection
  - Test snapshot reconstruction with corrupted chunks

- **Expected Behavior**
  - Corruption detected via SHA256 mismatch
  - Automatic retry with exponential backoff
  - Failed operations logged and reported
  - Snapshot integrity maintained through verification

## 4. Recovery and Re-sync Behavior

### Node Recovery
- **Cold Start Recovery**
  - New node joins existing cluster
  - Downloads current catalog state via gossip
  - Verifies Merkle root consistency
  - Applies missing deltas in causal order

- **Crash Recovery**
  - Node restarts after failure
  - Replays RAFT log for incomplete transactions
  - Checks BFT state for global transaction status
  - Completes or rolls back pending operations

### State Synchronization
- **Merkle-Based Anti-Entropy**
  - Periodic Merkle root exchange between peers
  - Root mismatch triggers full state comparison
  - Efficient delta transfer for small differences
  - Full sync for major divergence

- **CRDT State Reconciliation**
  - Vector clock comparison for causality
  - Application of newer operations only
  - Conflict resolution via CRDT semantics
  - Eventual consistency guarantee

### Transaction Recovery
- **Incomplete Transactions**
  - Identify prepared but uncommitted transactions
  - Query BFT network for commit status
  - Apply commit or trigger rollback accordingly
  - Ensure atomicity across domains

- **Network Healing**
  - Detect partition resolution
  - Exchange state summaries
  - Synchronize diverged operations
  - Validate consistency across all nodes

## Test Execution Guidelines

### Environment Setup
- Multi-node test cluster (minimum 4 nodes for BFT)
- Isolated network segments for partition testing
- Mock services for external dependencies (etcd, object store)
- Monitoring and logging infrastructure

### Success Criteria
- All unit tests pass with >95% coverage
- Integration tests complete end-to-end workflows
- Chaos tests demonstrate system resilience
- Recovery completes within specified time bounds
- No data loss or inconsistency after failures

### Performance Benchmarks
- Snapshot creation: <5 minutes for 100GB data
- Gossip propagation: <1 second for delta broadcast
- Recovery time: <30 seconds for node restart
- Consensus latency: <2 seconds for transaction commit
