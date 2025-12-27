# DeCube Glossary

Definitions of terms and concepts used in DeCube documentation.

## A

**Anti-Entropy**: A mechanism to detect and repair inconsistencies in distributed systems by comparing Merkle tree roots and synchronizing divergent state.

**Application Layer**: The top layer of DeCube architecture that provides high-level abstractions for computational workloads.

## B

**BFT (Byzantine Fault Tolerance)**: A consensus mechanism that can tolerate up to 1/3 of nodes being malicious or faulty.

**Block**: A collection of transactions that have been committed through consensus.

## C

**CAS (Content Addressable Storage)**: A storage system where data is addressed by its cryptographic hash rather than location.

**Catalog**: A CRDT-backed metadata store that tracks snapshots, images, and other resources.

**Chunk**: A fixed-size segment of data, typically 64MB, used for efficient storage and transfer.

**Cluster**: A group of nodes that work together using local RAFT consensus.

**Consensus Layer**: The layer responsible for achieving agreement across distributed nodes.

**CRDT (Conflict-Free Replicated Data Type)**: A data structure that can be replicated across multiple nodes and automatically resolve conflicts without coordination.

## D

**Delta**: A change or update to a CRDT that can be propagated to other nodes.

**Distributed Ledger**: An append-only transaction log maintained across multiple nodes.

## E

**ECDSA**: Elliptic Curve Digital Signature Algorithm, used for transaction authentication.

## G

**GCL (Global Consensus Layer)**: The BFT consensus mechanism that coordinates across multiple clusters.

**Gossip Protocol**: An epidemic protocol for efficient state dissemination across a network.

## H

**Hash**: A cryptographic digest of data, typically SHA-256, used for content addressing and integrity verification.

## L

**LWW Register (Last-Write-Wins Register)**: A CRDT type that resolves conflicts by timestamp, keeping the most recent write.

**Local Consensus**: RAFT consensus within a single cluster.

## M

**Merkle Tree**: A tree structure where each node is a hash of its children, used for efficient integrity verification.

**Metadata**: Information about data, such as size, creation time, and cluster location.

## N

**Node**: A single instance of DeCube running in a cluster.

## O

**OR-Set (Observed-Remove Set)**: A CRDT type that represents a set with add and remove operations that can be applied in any order.

## P

**PN-Counter (Positive-Negative Counter)**: A CRDT type that represents a counter that can be incremented and decremented.

**Proof**: Cryptographic evidence of a claim, such as data integrity or transaction finality.

## R

**RAFT**: A consensus algorithm that provides strong consistency within a cluster.

**Replication**: The process of copying data across multiple nodes for redundancy and availability.

## S

**Snapshot**: A point-in-time copy of system state, including data and metadata.

**Storage Layer**: The layer responsible for data persistence and retrieval.

## T

**Transaction**: An atomic operation that modifies system state.

**TLS (Transport Layer Security)**: A protocol for encrypted communication.

## V

**Vector Clock**: A mechanism for tracking causal relationships between events in a distributed system.

## Z

**Zero-Knowledge Proof (ZKP)**: A cryptographic proof that allows one party to prove knowledge of a value without revealing the value itself.

