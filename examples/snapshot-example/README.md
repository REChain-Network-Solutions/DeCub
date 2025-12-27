# Snapshot Example

This example demonstrates how to interact with DeCube's snapshot service.

## Prerequisites

- DeCube services running (start with `docker-compose up -d`)
- Go 1.19 or higher

## Running the Example

```bash
cd examples/snapshot-example
go run main.go
```

## What This Example Shows

1. **Service Health Check**: Waits for services to be ready
2. **Create Snapshot**: Creates a new snapshot with metadata
3. **Query Snapshot**: Retrieves snapshot information
4. **List Snapshots**: Lists all available snapshots
5. **Delete Snapshot**: Removes a snapshot

## API Endpoints Used

- `GET /catalog/health` - Health check
- `POST /snapshots` - Create snapshot
- `GET /snapshots/{id}` - Get snapshot
- `GET /catalog/snapshots` - List snapshots
- `DELETE /snapshots/{id}` - Delete snapshot

