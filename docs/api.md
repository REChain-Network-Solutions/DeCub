# DeCube API Reference

This document provides comprehensive API documentation for all DeCube services and components.

## Service Endpoints

| Service | Default Port | Protocol | Description |
|---------|-------------|----------|-------------|
| Catalog | 8080 | HTTP/REST | CRDT-based metadata management |
| Gossip | 8082 | HTTP/REST | Gossip synchronization control |
| GCL | 8081 | HTTP/REST | Global Consensus Layer |
| Control Plane | 8083 | HTTP/REST | Orchestration and management |
| Storage | 9000 | S3 | Object storage API |
| CAS | 9001 | HTTP/REST | Content Addressable Storage |

## Catalog Service API

Base URL: `http://localhost:8080`

### Snapshot Operations

#### Create Snapshot
```http
POST /api/v1/snapshots
Content-Type: application/json

{
  "id": "snapshot-001",
  "metadata": {
    "size": 1073741824,
    "created": "2024-01-15T10:30:00Z",
    "cluster": "cluster-a",
    "description": "Production database backup"
  }
}
```

**Response:**
```json
{
  "id": "snapshot-001",
  "status": "created",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Get Snapshot
```http
GET /api/v1/snapshots/{id}
```

**Response:**
```json
{
  "id": "snapshot-001",
  "metadata": {
    "size": 1073741824,
    "created": "2024-01-15T10:30:00Z",
    "cluster": "cluster-a",
    "chunks": 17,
    "hashes": ["sha256:..."],
    "status": "completed"
  }
}
```

#### Update Snapshot Metadata
```http
PUT /api/v1/snapshots/{id}/metadata
Content-Type: application/json

{
  "tags": ["production", "database"],
  "retention": "30d"
}
```

#### Delete Snapshot
```http
DELETE /api/v1/snapshots/{id}
```

#### List Snapshots
```http
GET /api/v1/snapshots?cluster=cluster-a&limit=10&offset=0
```

**Query Parameters:**
- `cluster`: Filter by cluster ID
- `status`: Filter by status (created, uploading, completed, failed)
- `limit`: Maximum number of results (default: 50)
- `offset`: Pagination offset (default: 0)

### Image Operations

#### Create Image
```http
POST /api/v1/images
Content-Type: application/json

{
  "id": "ubuntu-20.04",
  "metadata": {
    "os": "ubuntu",
    "version": "20.04",
    "size": 2147483648,
    "architecture": "amd64"
  }
}
```

#### Get Image
```http
GET /api/v1/images/{id}
```

#### List Images
```http
GET /api/v1/images?os=ubuntu&limit=20
```

### Query Operations

#### Advanced Query
```http
GET /api/v1/catalog/query?q=cluster=cluster-a AND size>1GB&type=snapshots
```

**Query Syntax:**
- Field operators: `=`, `!=`, `>`, `<`, `>=`, `<=`
- Logical operators: `AND`, `OR`, `NOT`
- Grouping: parentheses `( )`
- Wildcards: `*` for pattern matching

#### Search by Tags
```http
GET /api/v1/catalog/search?tags=production,database
```

### CRDT Operations

#### Get Pending Deltas
```http
GET /api/v1/crdt/deltas
```

**Response:**
```json
[
  {
    "node_id": "node-001",
    "vector_clock": {"node-001": 5, "node-002": 3},
    "type": "lww",
    "key": "snapshots:snapshot-001",
    "data": {
      "metadata": {"size": 1024}
    },
    "timestamp": 1705312200000000000
  }
]
```

#### Apply Delta
```http
POST /api/v1/crdt/deltas
Content-Type: application/json

{
  "node_id": "node-002",
  "vector_clock": {"node-001": 5, "node-002": 4},
  "type": "orset",
  "key": "images:ubuntu-20.04:tags",
  "data": {
    "add": ["web-server"],
    "remove": []
  },
  "timestamp": 1705312300000000000
}
```

#### Clear Processed Deltas
```http
POST /api/v1/crdt/deltas/clear
```

## Gossip Service API

Base URL: `http://localhost:8082`

### Status Operations

#### Get Gossip Status
```http
GET /api/v1/status
```

**Response:**
```json
{
  "node_id": "node-001",
  "merkle_root": "a1b2c3...",
  "peers": 5,
  "snapshots": 150,
  "images": 25,
  "pending_deltas": 3,
  "uptime": "2h30m",
  "version": "0.1.0"
}
```

### Synchronization Operations

#### Trigger Synchronization
```http
POST /api/v1/sync
```

**Response:**
```json
{
  "status": "sync_initiated",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Get Synchronization Status
```http
GET /api/v1/sync/status
```

**Response:**
```json
{
  "active": true,
  "last_sync": "2024-01-15T10:25:00Z",
  "next_sync": "2024-01-15T10:35:00Z",
  "sync_duration": "5.2s",
  "deltas_processed": 12
}
```

### Peer Management

#### List Peers
```http
GET /api/v1/peers
```

**Response:**
```json
[
  {
    "id": "QmPeer1...",
    "address": "/ip4/192.168.1.100/tcp/7000",
    "connected": true,
    "last_seen": "2024-01-15T10:29:00Z"
  }
]
```

#### Add Peer
```http
POST /api/v1/peers
Content-Type: application/json

{
  "address": "/ip4/192.168.1.101/tcp/7000/p2p/QmPeer2..."
}
```

#### Remove Peer
```http
DELETE /api/v1/peers/{peer_id}
```

### Merkle Operations

#### Get Merkle Root
```http
GET /api/v1/merkle/root
```

**Response:**
```json
{
  "root": "a1b2c3d4...",
  "height": 12,
  "leaf_count": 256
}
```

#### Get Merkle Proof
```http
GET /api/v1/merkle/proof?type=snapshot&id=snapshot-001
```

**Response:**
```json
{
  "root_hash": "a1b2c3d4...",
  "leaf_hash": "e5f6g7h8...",
  "proof": ["hash1", "hash2", "hash3"],
  "index": 42,
  "num_leaves": 256
}
```

#### Verify Merkle Proof
```http
POST /api/v1/merkle/verify
Content-Type: application/json

{
  "proof": {
    "root_hash": "a1b2c3d4...",
    "leaf_hash": "e5f6g7h8...",
    "proof": ["hash1", "hash2", "hash3"],
    "index": 42,
    "num_leaves": 256
  }
}
```

## Global Consensus Layer (GCL) API

Base URL: `http://localhost:8081`

### Transaction Operations

#### Submit Transaction
```http
POST /api/v1/transactions
Content-Type: application/json

{
  "type": "snapshot_create",
  "payload": {
    "snapshot_id": "snapshot-001",
    "cluster_id": "cluster-a",
    "size": 1073741824
  },
  "signature": "0x3045...",
  "public_key": "0x04..."
}
```

**Response:**
```json
{
  "tx_hash": "0x1234567890abcdef...",
  "status": "submitted",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Get Transaction
```http
GET /api/v1/transactions/{tx_hash}
```

**Response:**
```json
{
  "hash": "0x1234567890abcdef...",
  "type": "snapshot_create",
  "payload": {...},
  "status": "confirmed",
  "block_height": 12345,
  "block_hash": "0xabcdef...",
  "confirmations": 15,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Get Transaction Proof
```http
GET /api/v1/transactions/{tx_hash}/proof
```

**Response:**
```json
{
  "tx_hash": "0x1234567890abcdef...",
  "block_hash": "0xabcdef...",
  "block_height": 12345,
  "merkle_proof": ["hash1", "hash2", "hash3"],
  "signatures": [
    {
      "validator": "validator-001",
      "signature": "0x3045..."
    }
  ]
}
```

### Block Operations

#### Get Latest Block
```http
GET /api/v1/blocks/latest
```

#### Get Block by Height
```http
GET /api/v1/blocks/{height}
```

**Response:**
```json
{
  "height": 12345,
  "hash": "0xabcdef...",
  "previous_hash": "0x987654...",
  "timestamp": "2024-01-15T10:30:00Z",
  "transactions": ["0x123...", "0x456..."],
  "merkle_root": "0x789...",
  "validator": "validator-001",
  "signature": "0x3045..."
}
```

#### Get Block Transactions
```http
GET /api/v1/blocks/{height}/transactions
```

### Validator Operations

#### List Validators
```http
GET /api/v1/validators
```

**Response:**
```json
[
  {
    "id": "validator-001",
    "public_key": "0x04...",
    "voting_power": 100,
    "status": "active"
  }
]
```

#### Get Validator Status
```http
GET /api/v1/validators/{validator_id}
```

## Control Plane API

Base URL: `http://localhost:8083`

### Snapshot Management

#### Create Snapshot
```http
POST /api/v1/snapshots
Content-Type: application/json

{
  "id": "snapshot-001",
  "etcd_dir": "/var/lib/etcd",
  "volume_dirs": ["/data/volumes"],
  "cluster_id": "cluster-a"
}
```

#### Restore Snapshot
```http
POST /api/v1/snapshots/{id}/restore
Content-Type: application/json

{
  "target_dir": "/restore/path",
  "cluster_id": "cluster-a"
}
```

#### Get Snapshot Status
```http
GET /api/v1/snapshots/{id}/status
```

**Response:**
```json
{
  "id": "snapshot-001",
  "phase": "uploading",
  "progress": 65,
  "chunks_total": 17,
  "chunks_uploaded": 11,
  "estimated_time_remaining": "2m30s",
  "errors": []
}
```

### Cluster Operations

#### Get Cluster Status
```http
GET /api/v1/cluster/status
```

**Response:**
```json
{
  "cluster_id": "cluster-a",
  "nodes": [
    {
      "id": "node-001",
      "address": "192.168.1.100:8080",
      "status": "healthy",
      "role": "leader"
    }
  ],
  "services": {
    "catalog": "healthy",
    "gossip": "healthy",
    "gcl": "healthy"
  }
}
```

#### Scale Cluster
```http
POST /api/v1/cluster/scale
Content-Type: application/json

{
  "service": "catalog",
  "replicas": 3
}
```

## Storage Service API (S3-Compatible)

Base URL: `http://localhost:9000`

### Bucket Operations

#### List Buckets
```http
GET /
```

#### Create Bucket
```http
PUT /{bucket}
```

#### Delete Bucket
```http
DELETE /{bucket}
```

### Object Operations

#### Put Object
```http
PUT /{bucket}/{object}
Content-Type: application/octet-stream

<binary data>
```

#### Get Object
```http
GET /{bucket}/{object}
```

#### Delete Object
```http
DELETE /{bucket}/{object}
```

#### List Objects
```http
GET /{bucket}?prefix={prefix}&delimiter={delimiter}
```

### Multipart Upload

#### Initiate Multipart Upload
```http
POST /{bucket}/{object}?uploads
```

**Response:**
```xml
<InitiateMultipartUploadResult>
  <Bucket>bucket</Bucket>
  <Key>object</Key>
  <UploadId>upload-id</UploadId>
</InitiateMultipartUploadResult>
```

#### Upload Part
```http
PUT /{bucket}/{object}?partNumber={part}&uploadId={uploadId}
Content-Type: application/octet-stream

<binary data>
```

#### Complete Multipart Upload
```http
POST /{bucket}/{object}?uploadId={uploadId}
Content-Type: application/xml

<CompleteMultipartUpload>
  <Part>
    <ETag>"etag1"</ETag>
    <PartNumber>1</PartNumber>
  </Part>
  <Part>
    <ETag>"etag2"</ETag>
    <PartNumber>2</PartNumber>
  </Part>
</CompleteMultipartUpload>
```

## CAS (Content Addressable Storage) API

Base URL: `http://localhost:9001`

### Chunk Operations

#### Store Chunk
```http
PUT /api/v1/chunks/{hash}
Content-Type: application/octet-stream

<binary chunk data>
```

#### Retrieve Chunk
```http
GET /api/v1/chunks/{hash}
```

#### Verify Chunk
```http
HEAD /api/v1/chunks/{hash}
```

**Response Headers:**
- `X-Content-Length`: Size of the chunk
- `X-Content-Hash`: SHA-256 hash of the chunk
- `X-Upload-Time`: When the chunk was uploaded

### Batch Operations

#### Store Multiple Chunks
```http
POST /api/v1/chunks/batch
Content-Type: application/json

{
  "chunks": [
    {
      "hash": "sha256:...",
      "data": "base64-encoded-data"
    }
  ]
}
```

#### Retrieve Multiple Chunks
```http
POST /api/v1/chunks/batch/retrieve
Content-Type: application/json

{
  "hashes": ["sha256:...", "sha256:..."]
}
```

## Error Responses

All APIs return standard HTTP status codes and JSON error responses:

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "The request parameters are invalid",
    "details": {
      "field": "id",
      "reason": "must be alphanumeric"
    }
  }
}
```

### Common Error Codes
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource already exists or conflict
- `500 Internal Server Error`: Server-side error

## Authentication

APIs requiring authentication use Bearer token authentication:

```http
Authorization: Bearer <token>
```

Tokens are obtained through the authentication service (future implementation).

## Rate Limiting

APIs implement rate limiting to prevent abuse:

- **Catalog API**: 1000 requests per minute per IP
- **Gossip API**: 500 requests per minute per IP
- **GCL API**: 100 requests per minute per IP
- **Storage API**: 10000 requests per minute per IP

Rate limit headers are included in responses:
- `X-RateLimit-Limit`: Maximum requests per time window
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Time when the limit resets (Unix timestamp)

## Versioning

APIs are versioned using URL prefixes:
- `v1`: Current stable version
- `v2`: Next major version (under development)

Breaking changes will be introduced in new major versions only.

## SDKs and Client Libraries

Official client libraries are available for:
- **Go**: `go get github.com/decube/decube-go`
- **Python**: `pip install decube-python`
- **JavaScript**: `npm install @decube/client`

Community-maintained libraries:
- **Java**: Available on Maven Central
- **Rust**: Available on crates.io
- **C#**: Available on NuGet

## Webhooks

Services support webhooks for event-driven integrations:

```json
{
  "event": "snapshot.completed",
  "data": {
    "snapshot_id": "snapshot-001",
    "cluster_id": "cluster-a",
    "size": 1073741824
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

Configure webhooks through the control plane API.
