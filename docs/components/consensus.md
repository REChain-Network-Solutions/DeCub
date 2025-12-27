# Consensus Layer Documentation

The Consensus Layer provides both local (RAFT) and global (BFT) consensus for DeCube.

## Overview

DeCube uses a hybrid consensus approach:
- **Local Consensus (RAFT)**: Fast, strong consistency within clusters
- **Global Consensus (BFT)**: Byzantine fault tolerance across network

## RAFT Consensus

### Overview

RAFT provides strong consistency within a single cluster.

### States

- **Leader**: Handles all client requests
- **Follower**: Replicates log entries
- **Candidate**: Seeks votes for leadership

### Configuration

```yaml
raft:
  bind_addr: "0.0.0.0:7000"
  data_dir: "/var/lib/decube/raft"
  heartbeat_timeout: "1s"
  election_timeout: "2s"
  snapshot_interval: "30s"
  snapshot_threshold: 1000
```

### Performance

- **Latency**: <100ms for local operations
- **Throughput**: 10,000+ operations/second
- **Fault Tolerance**: Tolerates (N-1)/2 failures

## BFT Consensus

### Overview

Byzantine Fault Tolerant consensus for global coordination.

### Validators

- **Minimum**: 4 validators (tolerates 1 Byzantine)
- **Recommended**: 7+ validators
- **Fault Tolerance**: Tolerates up to 1/3 Byzantine nodes

### Configuration

```yaml
gcl:
  enabled: true
  endpoints:
    - "gcl-node-1:8080"
    - "gcl-node-2:8080"
    - "gcl-node-3:8080"
  timeout: "5s"
  retry_attempts: 3
  batch_size: 100
  batch_timeout: "100ms"
```

### Consensus Process

1. **Propose**: Validator proposes block
2. **Pre-Vote**: Validators pre-vote on proposal
3. **Vote**: Validators vote on proposal
4. **Commit**: Validators commit block

### Performance

- **Latency**: <2 seconds for global consensus
- **Throughput**: 1,000+ transactions/second
- **Finality**: Immediate after commit

## Hybrid 2PC

### Two-Phase Commit

Coordinates between local and global consensus:

1. **Prepare Phase**: Local RAFT prepare, GCL propose
2. **Commit Phase**: Global BFT commit, local RAFT apply

## Monitoring

### Metrics

- `raft_leader_elections_total` - Leader elections
- `raft_log_entries_total` - Log entries
- `bft_proposals_total` - BFT proposals
- `bft_commits_total` - BFT commits
- `consensus_latency_seconds` - Consensus latency

## Troubleshooting

### Common Issues

#### Leader Election Issues
- Check network connectivity
- Review election timeout
- Verify quorum

#### Slow Consensus
- Check network latency
- Review batch sizes
- Check validator health

## References

- [Architecture Guide](../architecture.md)
- [Performance Tuning](../performance-tuning.md)

---

*Last updated: January 2024*

