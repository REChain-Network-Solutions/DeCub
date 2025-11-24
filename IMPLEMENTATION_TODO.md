# REChain PoC Phase 0 Implementation Steps

## Detailed Steps to Complete All In Progress Items

### 1. Implement etcd snapshot tool (Go/Python script for snapshot, chunking, upload)
- [x] Update src/snapshot/snapshot.go: Add etcd client integration for snapshot creation
- [x] Integrate MinIO upload for chunks
- [x] Implement proper Merkle tree computation

### 2. Set up MinIO for CAS storage
- [x] Ensure MinIO integration in snapshot.go works with docker-compose

### 3. Compute Merkle tree for chunks
- [x] Enhance computeMerkleRoot function in snapshot.go for full Merkle tree

### 4. Mock GCL (single-node Tendermint or simple REST mock)
- [x] Enhance src/gcl/mock_gcl.go to run as REST server
- [x] Add to docker-compose.yml

### 5. Publish snapshot tx to GCL
- [x] Integrate GCL publishing in snapshot creation

### 6. Basic restore with Merkle verification
- [x] Add RestoreSnapshot function in snapshot.go

### 7. Simple CRDT catalog (in-memory OR-Set)
- [x] Create src/catalog/catalog.go with OR-Set implementation

### 8. REST API server for snapshot ops
- [x] Update src/api/server.go to integrate snapshot, catalog, GCL

### 9. CLI tool (rechainctl) for commands
- [x] Update cmd/rechainctl/main.go to call actual functions

### 10. Unit tests for Merkle, chunking
- [x] Enhance tests/snapshot_test.go
- [x] Add tests for catalog and Merkle

### 11. Integration test: snapshot -> upload -> restore
- [x] Create tests/integration_test.go

### 12. Docker Compose for local setup (etcd, MinIO, GCL mock)
- [x] Update config/docker-compose.yml with GCL mock service
