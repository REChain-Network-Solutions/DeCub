# Catalog Service Documentation

The Catalog Service provides CRDT-backed metadata management for DeCube.

## Overview

The Catalog Service maintains distributed metadata using Conflict-Free Replicated Data Types (CRDTs) to ensure eventual consistency across clusters.

## Architecture

### Components

- **CRDT Engine**: Handles CRDT operations (OR-Set, LWW Register, PN-Counter)
- **Vector Clocks**: Tracks causal relationships
- **Merkle Trees**: Ensures data integrity
- **Storage Backend**: Persistent storage for catalog data

### Data Model

```go
type CatalogEntry struct {
    ID        string
    Type      string  // "snapshot", "image", etc.
    Metadata  map[string]interface{}
    VectorClock map[string]uint64
    Timestamp   time.Time
}
```

## API

### REST API

#### List Entries

```http
GET /catalog/{type}
```

#### Get Entry

```http
GET /catalog/{type}/{id}
```

#### Create Entry

```http
POST /catalog/{type}
Content-Type: application/json

{
  "id": "entry-001",
  "metadata": {...}
}
```

#### Update Entry

```http
PUT /catalog/{type}/{id}
```

#### Delete Entry

```http
DELETE /catalog/{type}/{id}
```

### gRPC API

```protobuf
service Catalog {
  rpc Create(CreateRequest) returns (CreateResponse);
  rpc Get(GetRequest) returns (GetResponse);
  rpc List(ListRequest) returns (ListResponse);
  rpc Update(UpdateRequest) returns (UpdateResponse);
  rpc Delete(DeleteRequest) returns (DeleteResponse);
}
```

## Configuration

```yaml
catalog:
  enabled: true
  bind_addr: "0.0.0.0:8080"
  data_dir: "/var/lib/decube/catalog"
  crdt_type: "orset"  # orset, lww, pncounter
  batch_size: 100
  compaction_interval: "1h"
```

## CRDT Types

### OR-Set (Observed-Remove Set)

- **Use Case**: Snapshot lists, tag sets
- **Properties**: Add/remove operations are commutative
- **Conflict Resolution**: Unique tags per add operation

### LWW Register (Last-Write-Wins)

- **Use Case**: Single-value metadata
- **Properties**: Timestamp-based conflict resolution
- **Conflict Resolution**: Most recent write wins

### PN-Counter (Positive-Negative Counter)

- **Use Case**: Counters, statistics
- **Properties**: Increment/decrement operations
- **Conflict Resolution**: Sum of all increments minus sum of all decrements

## Performance

- **Throughput**: 10,000+ operations/second
- **Latency**: <10ms for local operations
- **Consistency**: Eventual consistency with anti-entropy

## Monitoring

### Metrics

- `catalog_operations_total` - Total operations
- `catalog_operations_duration_seconds` - Operation duration
- `catalog_entries_total` - Total entries
- `catalog_crdt_merges_total` - CRDT merge operations

## Troubleshooting

### Common Issues

#### High Latency
- Check storage I/O performance
- Review batch sizes
- Check network connectivity

#### Inconsistent Data
- Trigger anti-entropy sync
- Check Merkle tree roots
- Review vector clocks

## References

- [CRDT Documentation](../../rechain/pkg/crdt/README.md)
- [API Reference](../api-reference.md)

---

*Last updated: January 2024*

