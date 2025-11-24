# DeCube Global Consensus Layer (GCL)

This project implements the Global Consensus Layer (GCL) for DeCube in both Go and Rust.

## Features

- Append-only block ledger with Tendermint-like BFT consensus
- Block structure with header and transactions
- Merkle proof generation for transactions
- REST API endpoints:
  - POST /gcl/tx: Submit a transaction
  - GET /gcl/block/{height}: Get a block by height
  - GET /gcl/proof/{tx_id}: Get Merkle proof for a transaction
- Simulated quorum signatures (>=2/3 validators)

## Block Structure

```json
{
  "header": {
    "height": 1,
    "prev_hash": "",
    "merkle_root": "hash...",
    "proposer": "validator1",
    "timestamp": "2023-01-01T00:00:00Z"
  },
  "txs": [
    {
      "tx_id": "tx1",
      "type": "transfer",
      "origin": "user1",
      "payload": "data",
      "sig": "sig1"
    }
  ]
}
```

## Running

### Go Version

```bash
cd go
go mod tidy
go run .
```

### Rust Version

```bash
cd rust
cargo run
```

Both versions run on port 8080.

## API Usage

- Submit TX: `curl -X POST -H "Content-Type: application/json" -d '{"tx_id":"tx1","type":"transfer","origin":"user1","payload":"data","sig":"sig1"}' http://localhost:8080/gcl/tx`
- Get Block: `curl http://localhost:8080/gcl/block/1`
- Get Proof: `curl http://localhost:8080/gcl/proof/tx1`
