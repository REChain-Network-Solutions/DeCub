mod types;
mod merkle;
mod consensus;
mod api;

use api::{submit_tx, get_block, get_proof, Ledger};
use consensus::{Consensus, Validator};
use std::sync::Arc;
use warp::Filter;

#[tokio::main]
async fn main() {
    // Initialize consensus with mock validators
    let validators = vec![
        Validator {
            id: "val1".to_string(),
            pub_key: "pub1".to_string(),
        },
        Validator {
            id: "val2".to_string(),
            pub_key: "pub2".to_string(),
        },
        Validator {
            id: "val3".to_string(),
            pub_key: "pub3".to_string(),
        },
    ];
    let cons = Arc::new(Consensus::new(validators));
    let ledger: Ledger = Arc::new(std::sync::RwLock::new(Vec::new()));

    // Sample block JSON (as comment)
    // {
    //   "header": {
    //     "height": 1,
    //     "prev_hash": "",
    //     "merkle_root": "hash...",
    //     "proposer": "validator1",
    //     "timestamp": "2023-01-01T00:00:00Z"
    //   },
    //   "txs": [
    //     {
    //       "tx_id": "tx1",
    //       "type": "transfer",
    //       "origin": "user1",
    //       "payload": "data",
    //       "sig": "sig1"
    //     }
    //   ]
    // }

    let routes = submit_tx(ledger.clone(), cons.clone())
        .or(get_block(ledger.clone()))
        .or(get_proof(ledger.clone()));

    println!("Starting GCL server on :8080");
    warp::serve(routes).run(([127, 0, 0, 1], 8080)).await;
}
