use crate::consensus::Consensus;
use crate::merkle::generate_merkle_proof;
use crate::types::{Block, Transaction, hash_block};
use serde_json;
use std::collections::HashMap;
use std::sync::{Arc, RwLock};
use warp::Filter;

pub type Ledger = Arc<RwLock<Vec<Block>>>;

pub fn submit_tx(
    ledger: Ledger,
    cons: Arc<Consensus>,
) -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    warp::path!("gcl" / "tx")
        .and(warp::post())
        .and(warp::body::json())
        .and(with_ledger(ledger))
        .and(with_consensus(cons))
        .and_then(handle_submit_tx)
}

pub fn get_block(
    ledger: Ledger,
) -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    warp::path!("gcl" / "block" / u64)
        .and(warp::get())
        .and(with_ledger(ledger))
        .and_then(handle_get_block)
}

pub fn get_proof(
    ledger: Ledger,
) -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    warp::path!("gcl" / "proof" / String)
        .and(warp::get())
        .and(with_ledger(ledger))
        .and_then(handle_get_proof)
}

fn with_ledger(
    ledger: Ledger,
) -> impl Filter<Extract = (Ledger,), Error = std::convert::Infallible> + Clone {
    warp::any().map(move || ledger.clone())
}

fn with_consensus(
    cons: Arc<Consensus>,
) -> impl Filter<Extract = (Arc<Consensus>,), Error = std::convert::Infallible> + Clone {
    warp::any().map(move || cons.clone())
}

async fn handle_submit_tx(
    tx: Transaction,
    ledger: Ledger,
    cons: Arc<Consensus>,
) -> Result<impl warp::Reply, warp::Rejection> {
    let mut ledger_guard = ledger.write().unwrap();
    let height = ledger_guard.len() as u64 + 1;
    let prev_hash = if height > 1 {
        hash_block(&ledger_guard[height as usize - 2])
    } else {
        String::new()
    };
    let block = cons.propose_block(height, prev_hash, vec![tx], "validator1".to_string());
    let sigs = cons.sign_block(&block);
    if cons.verify_quorum(&sigs) {
        ledger_guard.push(block);
        Ok(warp::reply::with_status(
            format!("Transaction submitted, block {} created", height),
            warp::http::StatusCode::OK,
        ))
    } else {
        Ok(warp::reply::with_status(
            "Consensus failed".to_string(),
            warp::http::StatusCode::INTERNAL_SERVER_ERROR,
        ))
    }
}

async fn handle_get_block(height: u64, ledger: Ledger) -> Result<impl warp::Reply, warp::Rejection> {
    let ledger_guard = ledger.read().unwrap();
    if height < 1 || height > ledger_guard.len() as u64 {
        return Ok(warp::reply::with_status(
            "Block not found".to_string(),
            warp::http::StatusCode::NOT_FOUND,
        ));
    }
    let block = &ledger_guard[height as usize - 1];
    Ok(warp::reply::json(block))
}

async fn handle_get_proof(tx_id: String, ledger: Ledger) -> Result<impl warp::Reply, warp::Rejection> {
    let ledger_guard = ledger.read().unwrap();
    for block in ledger_guard.iter() {
        for (i, tx) in block.txs.iter().enumerate() {
            if tx.tx_id == tx_id {
                if let Some((root, _)) = crate::merkle::build_merkle_tree(&block.txs) {
                    let proof = generate_merkle_proof(&root, i);
                    return Ok(warp::reply::json(&proof));
                }
            }
        }
    }
    Ok(warp::reply::with_status(
        "Transaction not found".to_string(),
        warp::http::StatusCode::NOT_FOUND,
    ))
}
