# DeCube Python SDK

Official Python SDK for interacting with DeCube services.

## Installation

```bash
pip install decube-sdk
```

## Quick Start

```python
from decube import Client

# Create client
client = Client("http://localhost:8080", token="your-token")

# Create snapshot
snapshot = client.snapshots.create(
    id="snapshot-001",
    metadata={
        "cluster": "cluster-a",
        "size": 1073741824
    }
)

print(f"Created snapshot: {snapshot.id}")
```

## API Reference

### Client

```python
from decube import Client

# Create client
client = Client("http://localhost:8080")

# Create client with options
client = Client(
    endpoint="http://localhost:8080",
    token="your-token",
    timeout=30
)
```

### Snapshots

```python
# Create snapshot
snapshot = client.snapshots.create(
    id="snapshot-001",
    metadata={"cluster": "cluster-a"}
)

# Get snapshot
snapshot = client.snapshots.get("snapshot-001")

# List snapshots
snapshots = client.snapshots.list()

# Delete snapshot
client.snapshots.delete("snapshot-001")
```

### Catalog

```python
# List all snapshots
snapshots = client.catalog.list_snapshots()

# Query with filters
snapshots = client.catalog.query(
    type="snapshots",
    filters={"cluster": "cluster-a"}
)
```

### Gossip

```python
# Get gossip status
status = client.gossip.status()

# Trigger sync
client.gossip.sync()
```

## Error Handling

```python
from decube import Client, NotFoundError, UnauthorizedError

try:
    snapshot = client.snapshots.create(id="snapshot-001")
except NotFoundError:
    print("Snapshot not found")
except UnauthorizedError:
    print("Authentication failed")
except Exception as e:
    print(f"Error: {e}")
```

## Examples

See [examples/](../examples/) directory for more examples.

## Documentation

- [Integration Guide](../../docs/integration.md)
- [API Documentation](../../docs/api.md)

