# DeCube Gossip and CRDT Metadata Synchronization

This module implements gossip protocols and CRDT-based metadata synchronization for DeCube.

## Features

- Libp2p-based peer-to-peer networking
- GossipSub for efficient message dissemination
- CRDT (Conflict-free Replicated Data Type) for metadata
- Merkle tree-based anti-entropy for consistency
- LevelDB for persistent storage

## Architecture

- **Gossip Protocol**: Uses GossipSub for broadcasting metadata updates
- **CRDT**: Last-writer-wins map for key-value metadata
- **Anti-Entropy**: Periodic Merkle root exchange and full sync when needed
- **Persistence**: LevelDB stores CRDT state

## Running

```bash
go mod tidy
go run main.go /ip4/0.0.0.0/tcp/0
```

To connect to another node:

```bash
go run main.go /ip4/0.0.0.0/tcp/0 /ip4/192.168.1.100/tcp/4001/p2p/<peer-id>
```

## Topics

- `decub/metadata`: For CRDT updates
- `decub/anti-entropy`: For Merkle root and sync messages

## API

The gossip node doesn't expose an HTTP API; it operates purely via P2P messages.
