# Changelog

All notable changes to DeCube will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial implementation of gossip synchronization layer with CRDT integration
- Merkle tree-based anti-entropy for catalog consistency
- Enhanced CLI tools (decubectl, rechainctl) with comprehensive commands
- Docker Compose setup for local development
- Kubernetes manifests for production deployment
- Comprehensive test suite including chaos testing
- API documentation and setup guides

### Changed
- Improved consensus mechanisms with hybrid RAFT/BFT approach
- Enhanced cryptographic security with key rotation and ZKP support
- Updated storage layer with multi-tier architecture

### Fixed
- Resolved issues with delta serialization in gossip protocol
- Fixed Merkle proof verification edge cases
- Corrected transaction finality guarantees

## [0.1.0] - 2024-01-15

### Added
- Core CRDT catalog implementation with OR-Set and LWW-Register
- Basic snapshot lifecycle management (create, upload, restore)
- Content Addressable Storage (CAS) with MinIO integration
- Mock Global Consensus Layer (GCL) for testing
- REST API server for catalog operations
- Unit tests for Merkle trees and CRDT operations
- Integration tests for snapshot workflows

### Infrastructure
- Go module structure for all components
- Docker containers for each service
- Makefile for build automation
- Basic CI/CD pipeline setup

## [0.0.1] - 2023-12-01

### Added
- Initial project structure and architecture design
- Proof-of-concept implementations for core components
- Basic documentation and whitepaper
- Repository setup with GitHub workflows

---

## Types of Changes
- `Added` for new features
- `Changed` for changes in existing functionality
- `Deprecated` for soon-to-be removed features
- `Removed` for now removed features
- `Fixed` for any bug fixes
- `Security` in case of vulnerabilities

## Versioning
This project uses [Semantic Versioning](https://semver.org/). For the versions available, see the [tags on this repository](https://github.com/REChain-Network-Solutions/DeCub/tags).
