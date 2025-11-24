# DeCube Local Control Plane

This module implements the local control plane using etcd for strong consistency within a cluster.

## Features

- etcd client integration for key-value operations
- Snapshot creation and restoration
- REST API for control plane operations
- Configuration via Viper

## API Endpoints

- `POST /snapshot/create`: Create an etcd snapshot
- `POST /snapshot/restore`: Restore from snapshot
- `PUT /kv/{key}`: Put a key-value pair
- `GET /kv/{key}`: Get a value by key

## Running

```bash
go mod tidy
go run main.go
```

Assumes etcd is running on localhost:2379.

## Configuration

Create a `config.yaml`:

```yaml
etcd:
  endpoints:
    - "localhost:2379"
```

## Example Usage

Put a key:
```bash
curl -X PUT -H "Content-Type: application/json" -d '{"value": "hello"}' http://localhost:8080/kv/mykey
```

Get a key:
```bash
curl http://localhost:8080/kv/mykey
# Returns: {"key": "mykey", "value": "hello"}
```

Create snapshot:
```bash
curl -X POST http://localhost:8080/snapshot/create
