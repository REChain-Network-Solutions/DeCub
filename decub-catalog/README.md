# CRDT-Backed Catalog Service

A distributed metadata catalog service using Conflict-Free Replicated Data Types (CRDTs) for eventual consistency across multiple nodes.

## Features

- **OR-Set CRDT**: For snapshot and image lists with add/remove operations
- **LWW-Register CRDT**: For metadata fields with last-write-wins semantics
- **Vector Clocks**: For causal ordering and conflict detection
- **Delta Exchange**: Efficient gossip-based synchronization
- **REST API**: Standard HTTP endpoints for operations
- **Persistence**: LevelDB for durable state storage

## CRDT Types

### OR-Set (Observed-Remove Set)
- Used for snapshot and image collections
- Supports concurrent add/remove operations
- Resolves conflicts using unique tags per operation

### LWW-Register (Last-Write-Wins Register)
- Used for metadata fields
- Resolves conflicts by timestamp (later wins)
- Ensures metadata consistency across nodes

### Vector Clock
- Tracks causality between operations
- Prevents application of stale deltas
- Enables efficient conflict resolution

## API Endpoints

### Snapshot Operations
- `POST /snapshots/add/{id}` - Add snapshot with metadata
- `DELETE /snapshots/remove/{id}` - Remove snapshot
- `PUT /snapshots/metadata/{id}` - Update snapshot metadata

### Image Operations
- `POST /images/add/{id}` - Add image with metadata

### Query Operations
- `GET /catalog/query?type=snapshots&q=...` - Query catalog

### CRDT Operations
- `GET /crdt/delta` - Get pending deltas for gossip
- `POST /crdt/delta` - Apply received delta
- `POST /crdt/delta/clear` - Clear processed deltas

## Usage Examples

### Start Service
```bash
go run crdt_catalog.go
```

### Run Example
```bash
go run crdt_catalog.go example
```

### Add Snapshot
```bash
curl -X POST http://localhost:8080/snapshots/add/snap1 \
  -H "Content-Type: application/json" \
  -d '{"size": 1024, "created": "2023-01-01T00:00:00Z", "cluster": "cluster-a"}'
```

### Query Snapshots
```bash
curl "http://localhost:8080/catalog/query?type=snapshots"
```

### Get Deltas for Gossip
```bash
curl http://localhost:8080/crdt/delta
```

### Apply Delta
```bash
curl -X POST http://localhost:8080/crdt/delta \
  -H "Content-Type: application/json" \
  -d @delta.json
```

## Delta Exchange Example

```go
// Node1 generates deltas
deltas := node1.GetDeltas()

// Send deltas to Node2
for _, delta := range deltas {
    node2.ApplyDelta(delta)
}

// Node2 generates its deltas
deltas2 := node2.GetDeltas()

// Send back to Node1
for _, delta := range deltas2 {
    node1.ApplyDelta(delta)
}
```

## Conflict Resolution

### OR-Set Conflicts
- Add operations are always merged
- Remove operations affect all add tags
- Concurrent add/remove resolved by tag presence

### LWW Conflicts
- Later timestamp always wins
- Ties broken by node ID (lexicographic)

### Causal Ordering
- Vector clocks prevent applying stale operations
- Only newer operations are applied
- Maintains consistency across the cluster

## Architecture

```
┌─────────────┐    ┌─────────────┐
│   Node 1    │◄──►│   Node 2    │
│             │    │             │
│ OR-Set      │    │ OR-Set      │
│ LWW-Regs    │    │ LWW-Regs    │
│ Vector Clock│    │ Vector Clock│
└─────────────┘    └─────────────┘
       ▲                  ▲
       │                  │
       └──── Gossip ──────┘
```

## Integration with Gossip

The service integrates with gossip protocols (Serf/libp2p) for delta exchange:

1. **Generate Deltas**: `GET /crdt/delta` returns pending changes
2. **Apply Deltas**: `POST /crdt/delta` applies received changes
3. **Clear Deltas**: `POST /crdt/delta/clear` removes processed deltas

## Persistence

- **LevelDB**: Durable storage for CRDT state
- **Vector Clocks**: Serialized for crash recovery
- **OR-Sets**: Serialized with add/remove tags
- **Metadata**: LWW registers with timestamps

## Example Output

```
=== CRDT Catalog Example ===

--- Initial State ---
Node1 snapshots: []
Node2 snapshots: []

--- Node1 adds snapshot ---
Node1 snapshots: [{id:snap1 metadata:map[cluster:cluster-a created:2023-12-07T10:30:45Z size:1024]}]

--- Concurrent metadata updates (conflict) ---
Node1 metadata: map[size:1024 status:completed updated:2023-12-07T10:30:45Z]
Node2 metadata: map[size:1024 status:failed updated:2023-12-07T10:30:45Z]

--- Delta Exchange ---
Node1 sending 3 deltas to Node2
Applying delta: snapshots:snap1 (orset)
Applying delta: snapshot_metadata:snap1 (lww)
Applying delta: snapshots:snap2 (orset)
Node2 sending 1 deltas to Node1
Applying delta: snapshot_metadata:snap1 (lww)

--- After Sync ---
Node1 snapshots: [{id:snap1 metadata:map[size:1024 status:failed updated:2023-12-07T10:30:45Z]} {id:snap2 metadata:map[cluster:cluster-b created:2023-12-07T10:30:45Z size:2048]}]
Node2 snapshots: [{id:snap1 metadata:map[size:1024 status:failed updated:2023-12-07T10:30:45Z]} {id:snap2 metadata:map[cluster:cluster-b created:2023-12-07T10:30:45Z size:2048]}]

--- LWW Conflict Resolution ---
Both nodes have metadata for snap1, but with different timestamps.
The node with the later timestamp wins (LWW semantics).
Winner: map[size:1024 status:failed updated:2023-12-07T10:30:45Z]
```

This demonstrates eventual consistency, conflict resolution, and delta synchronization in a distributed CRDT system.
