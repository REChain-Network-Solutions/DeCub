# DeCube Go SDK

Official Go SDK for interacting with DeCube services.

## Installation

```bash
go get github.com/REChain-Network-Solutions/DeCub/sdk/go
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    
    "github.com/REChain-Network-Solutions/DeCub/sdk/go/decube"
)

func main() {
    // Create client
    client, err := decube.NewClient("http://localhost:8080")
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Create snapshot
    snapshot, err := client.Snapshots.Create(ctx, &decube.SnapshotRequest{
        ID: "snapshot-001",
        Metadata: map[string]interface{}{
            "cluster": "cluster-a",
            "size":    1073741824,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Created snapshot: %s", snapshot.ID)
}
```

## API Reference

### Client

```go
// Create client with default options
client, err := decube.NewClient("http://localhost:8080")

// Create client with custom options
client, err := decube.NewClientWithOptions("http://localhost:8080", &decube.Options{
    Timeout: 30 * time.Second,
    Token:   "your-token",
})
```

### Snapshots

```go
// Create snapshot
snapshot, err := client.Snapshots.Create(ctx, &decube.SnapshotRequest{
    ID: "snapshot-001",
    Metadata: map[string]interface{}{
        "cluster": "cluster-a",
    },
})

// Get snapshot
snapshot, err := client.Snapshots.Get(ctx, "snapshot-001")

// List snapshots
snapshots, err := client.Snapshots.List(ctx)

// Delete snapshot
err := client.Snapshots.Delete(ctx, "snapshot-001")
```

### Catalog

```go
// List all snapshots
snapshots, err := client.Catalog.ListSnapshots(ctx)

// Query with filters
snapshots, err := client.Catalog.Query(ctx, &decube.QueryRequest{
    Type: "snapshots",
    Filters: map[string]string{
        "cluster": "cluster-a",
    },
})
```

### Gossip

```go
// Get gossip status
status, err := client.Gossip.Status(ctx)

// Trigger sync
err := client.Gossip.Sync(ctx)
```

## Error Handling

```go
snapshot, err := client.Snapshots.Create(ctx, req)
if err != nil {
    if decube.IsNotFound(err) {
        log.Println("Snapshot not found")
    } else if decube.IsUnauthorized(err) {
        log.Println("Authentication failed")
    } else {
        log.Printf("Error: %v", err)
    }
    return
}
```

## Examples

See [examples/](../examples/) directory for more examples.

## Documentation

- [API Reference](https://pkg.go.dev/github.com/REChain-Network-Solutions/DeCub/sdk/go)
- [Integration Guide](../../docs/integration.md)

