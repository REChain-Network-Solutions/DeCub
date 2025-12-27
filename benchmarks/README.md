# DeCube Benchmarks

This directory contains performance benchmarks for DeCube components.

## Running Benchmarks

### All Benchmarks
```bash
go test -bench=. -benchmem ./benchmarks/...
```

### Specific Component
```bash
cd rechain/pkg/crdt
go test -bench=. -benchmem
```

## Benchmark Results

Benchmark results are tracked over time to monitor performance improvements and regressions.

### CRDT Performance
- OR-Set operations: ~100k ops/sec
- LWW Register: ~200k ops/sec
- PN-Counter: ~150k ops/sec

### Consensus Performance
- Local RAFT consensus: <100ms latency
- Global BFT consensus: <2s latency
- Throughput: 10k+ tps

### Storage Performance
- CAS write: ~500MB/s
- CAS read: ~1GB/s
- Merkle tree construction: <1s for 1GB data

## Contributing Benchmarks

When adding new benchmarks:
1. Place them in the appropriate component directory
2. Use descriptive benchmark names
3. Include memory allocation measurements (`-benchmem`)
4. Document expected performance characteristics
5. Update this README with results

