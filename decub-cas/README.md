# DeCube Content-Addressed Storage (CAS)

This module implements a Content-Addressed Storage system using MinIO (S3-compatible) with Merkle proofs for integrity.

## Features

- Content-addressed storage (data stored by its hash)
- Chunking for large files
- Merkle tree construction and proof generation
- S3-compatible API via MinIO
- LevelDB for local caching and metadata
- REST API for store/retrieve operations

## Architecture

- **Storage Backend**: MinIO (S3-compatible object store)
- **Addressing**: SHA-256 content hashing
- **Integrity**: Merkle trees for chunk verification
- **Caching**: LevelDB for fast local access
- **Chunking**: Configurable chunk sizes (default 1MB)

## API Endpoints

- `POST /store`: Store data, returns content hash
- `GET /retrieve/{hash}`: Retrieve data by hash
- `POST /chunk/store`: Chunk and store large data, returns hashes and Merkle root
- `GET /chunk/retrieve/{hashes}`: Retrieve and reassemble chunks

## Running

```bash
go mod tidy
go run main.go localhost:9000 minioadmin minioadmin [bucket-name]
```

Assumes MinIO is running locally on port 9000.

## Example Usage

Store data:
```bash
curl -X POST -d "Hello, World!" http://localhost:8080/store
# Returns: a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e38
```

Retrieve data:
```bash
curl http://localhost:8080/retrieve/a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e38
# Returns: Hello, World!
```

Chunk store:
```bash
curl -X POST -d "$(cat largefile.bin)" http://localhost:8080/chunk/store
# Returns: {"hashes": ["hash1", "hash2"], "merkle_root": "root_hash"}
