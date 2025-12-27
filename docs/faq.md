# Frequently Asked Questions (FAQ)

Common questions and answers about DeCube.

## General

### What is DeCube?

DeCube is a decentralized compute platform that implements Byzantine Fault Tolerant (BFT) consensus for secure, distributed execution of computational workloads across geographically distributed clusters.

### What problem does DeCube solve?

DeCube addresses the need for secure, scalable, and fault-tolerant distributed computing infrastructure that can operate across multiple clusters while maintaining data integrity and consistency.

### Is DeCube production-ready?

DeCube is currently in active development. Version 0.1.x is available for testing and evaluation. Production readiness is planned for version 0.2.0.

## Architecture

### How does DeCube handle consensus?

DeCube uses a hybrid consensus approach:
- **Local Consensus**: RAFT for strong consistency within clusters
- **Global Consensus**: BFT for Byzantine fault tolerance across clusters

### What is the difference between local and global consensus?

- **Local Consensus (RAFT)**: Fast, strong consistency within a single cluster
- **Global Consensus (BFT)**: Slower but more secure, tolerates Byzantine failures across the entire network

### How does gossip protocol work?

Gossip protocol uses an epidemic model to efficiently disseminate state changes across the network. Nodes periodically exchange deltas with random peers, ensuring eventual consistency.

## Storage

### What is Content Addressable Storage (CAS)?

CAS stores data by its cryptographic hash rather than location. This ensures data integrity and enables deduplication.

### How are snapshots stored?

Snapshots are chunked into 64MB segments, each with a SHA-256 hash. Chunks are stored in CAS, and metadata is tracked in the CRDT catalog.

### Can I use my own storage backend?

Currently, DeCube uses S3-compatible storage. Support for additional backends is planned for future versions.

## CRDTs

### What CRDT types are supported?

Currently supported:
- OR-Set (Observed-Remove Set)
- LWW Register (Last-Write-Wins Register)
- PN-Counter (Positive-Negative Counter)

More types are planned for future releases.

### How do CRDTs resolve conflicts?

CRDTs are mathematically designed to resolve conflicts automatically without coordination. Different CRDT types use different conflict resolution strategies (e.g., LWW uses timestamps, OR-Set uses unique tags).

## Security

### How secure is DeCube?

DeCube implements multiple security layers:
- ECDSA digital signatures for authentication
- TLS 1.3 for encrypted communication
- Merkle trees for data integrity verification
- BFT consensus for Byzantine fault tolerance

### How are keys managed?

Keys are managed hierarchically with automatic rotation every 90 days. Support for Hardware Security Modules (HSM) is planned.

### Does DeCube support encryption at rest?

Yes, DeCube supports AES-256-GCM encryption for data at rest.

## Performance

### What is the throughput?

- Local operations: 10,000+ transactions per second
- Global consensus: Depends on network latency, typically <2 seconds
- Storage: 500MB/s write, 1GB/s read

### How many nodes can a cluster support?

Currently tested up to 100 nodes per cluster. Support for 1000+ nodes is planned.

### What is the latency?

- Local RAFT consensus: <100ms
- Global BFT consensus: <2 seconds
- Cross-cluster sync: <1 second for deltas

## Deployment

### How do I deploy DeCube?

DeCube can be deployed using:
- Docker Compose (for development/testing)
- Kubernetes (for production)
- Standalone binaries

See the [Deployment Guide](deployment.md) for details.

### What are the system requirements?

- CPU: 2+ cores
- RAM: 4GB minimum, 8GB+ recommended
- Disk: 10GB+ free space
- Network: Stable internet connection

### Can I run DeCube on cloud providers?

Yes, DeCube can run on any cloud provider that supports Docker or Kubernetes, including AWS, GCP, Azure, and others.

## Development

### How do I contribute?

See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines. We welcome contributions!

### What programming languages are used?

Primarily Go, with some components in Rust. See the [Development Guide](development.md) for details.

### How do I report bugs?

Open an issue on GitHub with:
- Steps to reproduce
- Expected vs actual behavior
- Environment details
- Logs and error messages

### How do I request features?

Open a feature request issue with:
- Use case description
- Proposed implementation
- Benefits and impact

## Licensing

### What license does DeCube use?

DeCube is licensed under the BSD 3-Clause License. See [LICENSE](../LICENSE) for details.

### Can I use DeCube commercially?

Yes, the BSD 3-Clause License allows commercial use.

## Support

### Where can I get help?

- **Documentation**: Check the [docs](.) directory
- **Issues**: [GitHub Issues](https://github.com/REChain-Network-Solutions/DeCub/issues)
- **Discussions**: [GitHub Discussions](https://github.com/REChain-Network-Solutions/DeCub/discussions)
- **Security**: Email security@decube.io

### Is there commercial support available?

Commercial support options are being developed. Contact us for more information.

## Roadmap

### What's coming next?

See the [Roadmap](roadmap.md) for planned features and improvements.

### When will version X be released?

Release dates are estimates and subject to change. Check the roadmap and GitHub releases for updates.

---

Have a question not answered here? Open an issue or start a discussion on GitHub!

