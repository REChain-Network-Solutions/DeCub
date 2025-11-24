# Decub Snapshot Lifecycle Manager

A comprehensive Go-based tool for managing the full lifecycle of Decub snapshots, including creation, chunking, uploading, registration, verification, and restoration.

## Features

- **Snapshot Creation**: Create snapshots from etcd and volume data
- **Data Chunking**: Automatically chunk large files into 64MB pieces
- **Object Store Upload**: Upload chunks with SHA256 verification
- **Metadata Registration**: Register snapshot metadata via GCL transactions
- **Verification & Restoration**: Verify integrity and restore snapshots

## Installation

```bash
go mod tidy
go build -o decub-snapshot
```

## Usage

### Create a Snapshot

```bash
./decub-snapshot create my-snapshot /var/lib/etcd /var/lib/volumes \
  --etcd http://localhost:2379 \
  --object-store http://localhost:9000 \
  --gcl http://localhost:8080
```

This will:
1. Create etcd snapshot using `etcdctl`
2. Create volume snapshot using `tar`
3. Combine the snapshots
4. Chunk the data into 64MB files
5. Upload chunks to object store with SHA256 verification
6. Register metadata via GCL transaction

### Restore a Snapshot

```bash
./decub-snapshot restore my-snapshot /tmp/restore \
  --etcd http://localhost:2379 \
  --object-store http://localhost:9000 \
  --gcl http://localhost:8080
```

This will:
1. Retrieve snapshot metadata from GCL
2. Download and verify all chunks
3. Reconstruct the original snapshot
4. Extract etcd and volume data to restore path

## Workflow Details

### 1. Snapshot Creation
- Uses `etcdctl snapshot save` to create etcd backup
- Uses `tar -czf` to compress volume data
- Combines etcd and volume snapshots into a single file

### 2. Data Chunking
- Splits large files into 64MB chunks
- Each chunk is stored as a separate file

### 3. Upload with Verification
- Calculates SHA256 hash for each chunk
- Uploads to object store (S3-compatible)
- Stores hash for later verification

### 4. Metadata Registration
- Creates metadata including:
  - Snapshot ID
  - Timestamp
  - Chunk count
  - SHA256 hashes for each chunk
  - Total size
- Registers via GCL transaction

### 5. Verification and Restoration
- Retrieves metadata from GCL
- Downloads each chunk
- Verifies SHA256 hash against stored value
- Reconstructs original file from chunks
- Extracts etcd and volume data

## Dependencies

- etcdctl (for etcd snapshots)
- tar (for volume compression)
- awscli (for S3-compatible object store)
- gcl-cli (for GCL transactions)

## Configuration

All endpoints can be configured via command-line flags:
- `--etcd`: Etcd endpoint (default: http://localhost:2379)
- `--object-store`: Object store endpoint (default: http://localhost:9000)
- `--gcl`: GCL endpoint (default: http://localhost:8080)

## Logs

The tool provides detailed logging for each step, including:
- Commands executed
- File paths and sizes
- Hash calculations
- Upload/download progress
- Verification results
