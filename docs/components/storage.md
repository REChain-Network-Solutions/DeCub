# Storage Layer Documentation

The Storage Layer provides multi-tiered storage for DeCube.

## Overview

DeCube uses multiple storage backends:
- **CAS (Content Addressable Storage)**: Immutable blob storage
- **Object Storage**: S3-compatible storage
- **Distributed Ledger**: Append-only transaction log

## Content Addressable Storage (CAS)

### Overview

CAS stores data by cryptographic hash, ensuring integrity and deduplication.

### Features

- **Immutable**: Data cannot be modified
- **Deduplication**: Same content = same address
- **Integrity**: Hash-based verification
- **Distributed**: Replicated across nodes

### Configuration

```yaml
storage:
  cas:
    endpoint: "http://cas:9000"
    access_key: "minioadmin"
    secret_key: "minioadmin"
    bucket: "decube-cas"
    chunk_size: 67108864  # 64MB
    compression: "lz4"
```

### API

#### Upload Chunk

```http
PUT /cas/chunks/{hash}
Content-Type: application/octet-stream

<chunk data>
```

#### Get Chunk

```http
GET /cas/chunks/{hash}
```

### Performance

- **Write**: 500MB/s
- **Read**: 1GB/s
- **Deduplication**: 30-50% space savings

## Object Storage

### Overview

S3-compatible object storage for large objects.

### Configuration

```yaml
storage:
  object_storage:
    endpoint: "http://minio:9000"
    access_key: "minioadmin"
    secret_key: "minioadmin"
    bucket: "decube-objects"
    multipart_threshold: 52428800  # 50MB
    multipart_chunk_size: 10485760  # 10MB
```

### Features

- **S3-Compatible**: Standard S3 API
- **Multipart Upload**: Large file support
- **Versioning**: Object versioning
- **Lifecycle**: Automatic lifecycle management

## Distributed Ledger

### Overview

Append-only transaction log with Merkle tree verification.

### Features

- **Immutable**: Append-only
- **Verifiable**: Merkle tree proofs
- **Distributed**: Replicated across nodes
- **Auditable**: Complete transaction history

### Configuration

```yaml
storage:
  ledger:
    enabled: true
    data_dir: "/var/lib/decube/ledger"
    segment_size: 1073741824  # 1GB
    retention_days: 365
```

## Monitoring

### Metrics

- `storage_cas_operations_total` - CAS operations
- `storage_cas_bytes_total` - CAS bytes transferred
- `storage_object_operations_total` - Object operations
- `storage_ledger_entries_total` - Ledger entries

## Troubleshooting

### Common Issues

#### Slow Storage
- Check disk I/O
- Review network latency
- Check storage backend health

#### Storage Full
- Enable compression
- Review retention policies
- Clean up old data

## References

- [Architecture Guide](../architecture.md)
- [Performance Tuning](../performance-tuning.md)

---

*Last updated: January 2024*

