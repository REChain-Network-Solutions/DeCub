use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Transaction {
    pub tx_id: String,
    pub tx_type: String,
    pub origin: String,
    pub payload: String,
    pub sig: String,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Header {
    pub height: u64,
    pub prev_hash: String,
    pub merkle_root: String,
    pub proposer: String,
    pub timestamp: DateTime<Utc>,
}

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct Block {
    pub header: Header,
    pub txs: Vec<Transaction>,
}

#[derive(Clone, Debug)]
pub struct MerkleNode {
    pub hash: String,
    pub left: Option<Box<MerkleNode>>,
    pub right: Option<Box<MerkleNode>>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct MerkleProof {
    pub hashes: Vec<String>,
    pub index: usize,
}

pub fn hash_transaction(tx: &Transaction) -> String {
    let data = format!("{}{}{}{}{}", tx.tx_id, tx.tx_type, tx.origin, tx.payload, tx.sig);
    let mut hasher = Sha256::new();
    hasher.update(data);
    format!("{:x}", hasher.finalize())
}

pub fn hash_block(block: &Block) -> String {
    let data = format!(
        "{}{}{}{}",
        block.header.prev_hash,
        block.header.merkle_root,
        block.header.proposer,
        block.header.timestamp.to_rfc3339()
    );
    let mut hasher = Sha256::new();
    hasher.update(data);
    format!("{:x}", hasher.finalize())
}
