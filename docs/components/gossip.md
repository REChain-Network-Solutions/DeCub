# Gossip Protocol Documentation

The Gossip Protocol provides efficient state synchronization across the DeCube network.

## Overview

The Gossip Protocol uses an epidemic model to disseminate state changes across the network, ensuring eventual consistency.

## Architecture

### Components

- **Gossip Engine**: Core gossip protocol implementation
- **Delta Generator**: Creates deltas from local changes
- **Anti-Entropy**: Detects and repairs inconsistencies
- **Peer Manager**: Manages peer connections

### Protocol Flow

1. **Local Change**: Node makes local change
2. **Delta Generation**: Change is converted to delta
3. **Gossip Dissemination**: Delta is sent to random peers
4. **Delta Application**: Peers apply delta and propagate
5. **Anti-Entropy**: Periodic consistency checks

## Configuration

```yaml
gossip:
  enabled: true
  bind_addr: "0.0.0.0:8000"
  advertise_addr: ""
  peers:
    - "node-1:8000"
    - "node-2:8000"
    - "node-3:8000"
  sync_interval: "10s"
  anti_entropy_interval: "60s"
  fanout: 3
  max_message_size: 1048576  # 1MB
```

## Message Types

### Delta Message

```go
type DeltaMessage struct {
    NodeID      string
    VectorClock map[string]uint64
    Type        string
    Key         string
    Data        interface{}
    Timestamp   int64
    Signature   []byte
}
```

### Sync Request

```go
type SyncRequest struct {
    NodeID      string
    VectorClock map[string]uint64
    MerkleRoot  []byte
}
```

### Sync Response

```go
type SyncResponse struct {
    Deltas     []DeltaMessage
    MerkleRoot []byte
    FullSync   bool
}
```

## Anti-Entropy

### Process

1. **Merkle Root Comparison**: Compare Merkle tree roots
2. **Divergence Detection**: Identify differences
3. **Delta Exchange**: Exchange missing deltas
4. **Full Sync**: If needed, perform full state sync

### Configuration

```yaml
anti_entropy:
  enabled: true
  interval: "60s"
  merkle_tree_depth: 16
  full_sync_threshold: 0.1  # 10% divergence triggers full sync
```

## Performance

- **Delta Propagation**: <1 second across network
- **Sync Latency**: <5 seconds for full sync
- **Network Overhead**: <1% of bandwidth

## Monitoring

### Metrics

- `gossip_deltas_sent_total` - Deltas sent
- `gossip_deltas_received_total` - Deltas received
- `gossip_sync_duration_seconds` - Sync duration
- `gossip_peers_connected` - Connected peers

## Troubleshooting

### Common Issues

#### Slow Propagation
- Check network latency
- Review fanout configuration
- Check message size limits

#### Inconsistent State
- Trigger manual sync
- Check anti-entropy logs
- Review Merkle tree roots

## References

- [Network Protocol](../network-protocol.md)
- [Architecture Guide](../architecture.md)

---

*Last updated: January 2024*

