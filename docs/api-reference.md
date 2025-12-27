# API Reference

Complete API reference for DeCube services.

## Table of Contents

1. [REST API](#rest-api)
2. [gRPC API](#grpc-api)
3. [Authentication](#authentication)
4. [Error Handling](#error-handling)
5. [Rate Limiting](#rate-limiting)

## REST API

### Base URL

```
http://localhost:8080
```

### Authentication

All API requests require authentication via Bearer token:

```http
Authorization: Bearer <token>
```

### Catalog Service

#### List Snapshots

```http
GET /catalog/snapshots
```

**Response:**
```json
{
  "snapshots": [
    {
      "id": "snapshot-001",
      "metadata": {
        "size": 1073741824,
        "created": "2024-01-15T10:30:00Z",
        "cluster": "cluster-a"
      }
    }
  ]
}
```

#### Get Snapshot

```http
GET /catalog/snapshots/{id}
```

**Response:**
```json
{
  "id": "snapshot-001",
  "metadata": {
    "size": 1073741824,
    "created": "2024-01-15T10:30:00Z",
    "cluster": "cluster-a"
  }
}
```

#### Query Catalog

```http
GET /catalog/query?type={type}&{filters}
```

**Parameters:**
- `type`: Resource type (snapshots, images, etc.)
- `filters`: Key-value pairs for filtering

**Example:**
```http
GET /catalog/query?type=snapshots&cluster=cluster-a&limit=10
```

### Snapshot Service

#### Create Snapshot

```http
POST /snapshots
Content-Type: application/json

{
  "id": "snapshot-001",
  "metadata": {
    "size": 1073741824,
    "cluster": "cluster-a"
  }
}
```

**Response:**
```json
{
  "id": "snapshot-001",
  "status": "created",
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### Get Snapshot Status

```http
GET /snapshots/{id}/status
```

#### Delete Snapshot

```http
DELETE /snapshots/{id}
```

### Gossip Service

#### Get Status

```http
GET /gossip/status
```

**Response:**
```json
{
  "node_id": "node-001",
  "peers": 3,
  "sync_status": "synced",
  "last_sync": "2024-01-15T10:30:00Z"
}
```

#### Trigger Sync

```http
POST /gossip/sync
```

#### Get Deltas

```http
GET /gossip/deltas?since={timestamp}
```

### Consensus Service

#### Submit Transaction

```http
POST /consensus/transactions
Content-Type: application/json

{
  "type": "snapshot_create",
  "payload": {
    "snapshot_id": "snapshot-001",
    "cluster_id": "cluster-a"
  },
  "signature": "0x..."
}
```

#### Get Block

```http
GET /consensus/blocks/{height}
```

#### Get Transaction

```http
GET /consensus/transactions/{tx_hash}
```

### Storage Service

#### Upload Object

```http
PUT /storage/objects/{object_id}
Content-Type: application/octet-stream

<binary data>
```

#### Get Object

```http
GET /storage/objects/{object_id}
```

#### Upload Chunk

```http
POST /storage/chunks
Content-Type: application/json

{
  "snapshot_id": "snapshot-001",
  "chunk_index": 0,
  "data": "<base64-encoded>",
  "hash": "sha256-hash"
}
```

### Health and Status

#### Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Readiness Check

```http
GET /ready
```

#### Metrics

```http
GET /metrics
```

## gRPC API

### Service Definitions

#### DeCube Service

```protobuf
service DeCube {
  rpc CreateSnapshot(CreateSnapshotRequest) returns (CreateSnapshotResponse);
  rpc GetSnapshot(GetSnapshotRequest) returns (GetSnapshotResponse);
  rpc ListSnapshots(ListSnapshotsRequest) returns (ListSnapshotsResponse);
  rpc DeleteSnapshot(DeleteSnapshotRequest) returns (DeleteSnapshotResponse);
}
```

#### Catalog Service

```protobuf
service Catalog {
  rpc ListSnapshots(ListSnapshotsRequest) returns (ListSnapshotsResponse);
  rpc GetSnapshot(GetSnapshotRequest) returns (GetSnapshotResponse);
  rpc Query(QueryRequest) returns (QueryResponse);
}
```

### Message Types

#### CreateSnapshotRequest

```protobuf
message CreateSnapshotRequest {
  string id = 1;
  map<string, string> metadata = 2;
}
```

#### CreateSnapshotResponse

```protobuf
message CreateSnapshotResponse {
  string id = 1;
  string status = 2;
  google.protobuf.Timestamp created_at = 3;
}
```

## Authentication

### JWT Authentication

#### Get Token

```http
POST /auth/login
Content-Type: application/json

{
  "username": "user",
  "password": "pass"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600
}
```

#### Refresh Token

```http
POST /auth/refresh
Authorization: Bearer <token>
```

### mTLS Authentication

For gRPC, use mutual TLS:

```go
creds, err := credentials.NewClientTLSFromFile("ca.pem", "")
conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(creds))
```

## Error Handling

### Error Response Format

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Snapshot not found",
    "details": {
      "snapshot_id": "snapshot-001"
    }
  }
}
```

### HTTP Status Codes

- `200 OK` - Success
- `201 Created` - Resource created
- `400 Bad Request` - Invalid request
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - Service unavailable

### Error Codes

- `INVALID_REQUEST` - Invalid request format
- `UNAUTHORIZED` - Authentication failed
- `FORBIDDEN` - Insufficient permissions
- `NOT_FOUND` - Resource not found
- `CONFLICT` - Resource conflict
- `RATE_LIMIT_EXCEEDED` - Rate limit exceeded
- `INTERNAL_ERROR` - Internal server error
- `SERVICE_UNAVAILABLE` - Service unavailable

## Rate Limiting

### Limits

- **Default**: 1000 requests per minute per IP
- **Authenticated**: 5000 requests per minute per user
- **Burst**: 100 requests per second

### Headers

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1642248000
```

### Rate Limit Exceeded

```http
HTTP/1.1 429 Too Many Requests
Retry-After: 60

{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded",
    "retry_after": 60
  }
}
```

## Pagination

### Request

```http
GET /catalog/snapshots?page=1&limit=10
```

### Response

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 100,
    "pages": 10
  }
}
```

## Filtering and Sorting

### Filtering

```http
GET /catalog/snapshots?cluster=cluster-a&size_min=1000000
```

### Sorting

```http
GET /catalog/snapshots?sort=created_at&order=desc
```

## Webhooks

### Register Webhook

```http
POST /webhooks
Content-Type: application/json

{
  "url": "https://example.com/webhook",
  "events": ["snapshot.created", "snapshot.deleted"],
  "secret": "webhook-secret"
}
```

### Webhook Payload

```json
{
  "event": "snapshot.created",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "id": "snapshot-001",
    "metadata": {...}
  },
  "signature": "sha256-hmac-signature"
}
```

## SDK Examples

### Go

```go
import "github.com/REChain-Network-Solutions/DeCub/sdk/go/decube"

client, _ := decube.NewClient("http://localhost:8080")
snapshot, _ := client.Snapshots.Create(ctx, &decube.SnapshotRequest{
    ID: "snapshot-001",
})
```

### Python

```python
from decube import Client

client = Client("http://localhost:8080")
snapshot = client.snapshots.create(id="snapshot-001")
```

### JavaScript

```javascript
const decube = require('decube-sdk');

const client = new decube.Client('http://localhost:8080');
const snapshot = await client.snapshots.create({
    id: 'snapshot-001'
});
```

## OpenAPI Specification

See `api/openapi.yaml` for complete OpenAPI 3.0 specification.

## Versioning

API versioning is done via URL path:

```
/api/v1/snapshots
```

Current version: `v1`

## Changelog

See [API Changelog](api-changelog.md) for API changes.

---

*Last updated: January 2024*

