# Network Protocol Documentation

This document describes the network protocols used by DeCube.

## Table of Contents

1. [Protocol Overview](#protocol-overview)
2. [Gossip Protocol](#gossip-protocol)
3. [Consensus Protocol](#consensus-protocol)
4. [Storage Protocol](#storage-protocol)
5. [Message Formats](#message-formats)

## Protocol Overview

DeCube uses multiple protocols for different purposes:

- **HTTP/gRPC**: Client-server communication
- **Gossip Protocol**: Peer-to-peer state synchronization
- **RAFT**: Local cluster consensus
- **BFT**: Global consensus
- **libp2p**: Peer discovery and networking

## Gossip Protocol

### Overview

The gossip protocol is used for efficient state dissemination across the network.

### Message Types

#### Delta Message

```go
type DeltaMessage struct {
    NodeID      string
    VectorClock map[string]uint64
    Type        string  // "orset", "lww", "pncounter"
    Key         string
    Data        interface{}
    Timestamp   int64
    Signature   []byte
}
```

#### Sync Request

```go
type SyncRequest struct {
    NodeID      string
    VectorClock map[string]uint64
    MerkleRoot  []byte
}
```

#### Sync Response

```go
type SyncResponse struct {
    Deltas      []DeltaMessage
    MerkleRoot  []byte
    FullSync    bool
}
```

### Protocol Flow

1. **Periodic Sync**
   - Nodes periodically exchange sync requests
   - Compare Merkle roots
   - Exchange deltas if roots differ

2. **Delta Propagation**
   - Local changes generate deltas
   - Deltas are gossiped to random peers
   - Peers apply deltas and propagate further

3. **Anti-Entropy**
   - Periodic full state comparison
   - Merkle tree root comparison
   - Full sync if roots mismatch

### Message Format

```
[Header: 8 bytes][Payload: variable]
```

Header:
- Version: 1 byte
- Type: 1 byte
- Length: 4 bytes
- Checksum: 2 bytes

## Consensus Protocol

### RAFT Protocol

#### Message Types

- **AppendEntries**: Leader sends log entries to followers
- **AppendEntriesResponse**: Follower response
- **RequestVote**: Candidate requests votes
- **RequestVoteResponse**: Voter response
- **InstallSnapshot**: Leader sends snapshot

#### Message Format

```go
type AppendEntries struct {
    Term         uint64
    LeaderID     string
    PrevLogIndex uint64
    PrevLogTerm  uint64
    Entries      []LogEntry
    LeaderCommit uint64
}
```

### BFT Protocol

#### Message Types

- **Propose**: Proposer sends proposal
- **PreVote**: Validator pre-votes
- **Vote**: Validator votes
- **Commit**: Validator commits

#### Message Format

```go
type Proposal struct {
    Block       Block
    Round       uint64
    ProposerID  string
    Signature   []byte
}
```

## Storage Protocol

### CAS Protocol

#### Upload Chunk

```
PUT /cas/chunks/{hash}
Content-Type: application/octet-stream
X-Chunk-Size: <size>
X-Chunk-Index: <index>

<chunk data>
```

#### Get Chunk

```
GET /cas/chunks/{hash}
```

Response:
```
Content-Type: application/octet-stream
X-Chunk-Size: <size>
X-Chunk-Hash: <hash>

<chunk data>
```

### Object Storage Protocol

S3-compatible protocol for object storage operations.

## Message Formats

### Binary Protocol

#### Header

```
+------------------+
| Version (1 byte) |
+------------------+
| Type (1 byte)    |
+------------------+
| Length (4 bytes) |
+------------------+
| Checksum (2)     |
+------------------+
```

#### Message Types

- `0x01`: Delta message
- `0x02`: Sync request
- `0x03`: Sync response
- `0x04`: Consensus message
- `0x05`: Storage message

### JSON Protocol

For HTTP/REST APIs, JSON is used:

```json
{
  "type": "delta",
  "node_id": "node-001",
  "vector_clock": {
    "node-001": 5,
    "node-002": 3
  },
  "data": {
    "key": "snapshots:snapshot-001",
    "value": {...}
  }
}
```

## Protocol Versions

### Version 1.0

- Initial protocol version
- Gossip protocol
- RAFT consensus
- Basic BFT

### Version 1.1 (Planned)

- Enhanced gossip with compression
- Improved BFT performance
- Protocol compression

## Security

### Message Authentication

All messages are authenticated using ECDSA signatures:

```go
signature = ECDSA_Sign(private_key, message_hash)
```

### Message Encryption

Sensitive messages are encrypted using AES-256-GCM:

```go
encrypted = AES_GCM_Encrypt(key, nonce, plaintext)
```

### TLS

All client-server communication uses TLS 1.3.

## Performance Considerations

### Message Size

- Maximum message size: 1MB
- Recommended delta size: <64KB
- Chunk size: 64MB

### Batching

Multiple operations can be batched:

```go
type BatchMessage struct {
    Operations []Operation
    Signature  []byte
}
```

## References

- [Architecture Guide](architecture.md)
- [API Reference](api-reference.md)
- [Security Guide](../SECURITY.md)

---

*Last updated: January 2024*

