use crate::types::{Block, Header, Transaction, hash_block};
use chrono::Utc;
use sha2::{Digest, Sha256};

#[derive(Clone, Debug)]
pub struct Validator {
    pub id: String,
    pub pub_key: String,
}

#[derive(Clone, Debug)]
pub struct Consensus {
    pub validators: Vec<Validator>,
    pub threshold: usize, // >=2/3
}

impl Consensus {
    pub fn new(validators: Vec<Validator>) -> Self {
        let threshold = (2 * validators.len()) / 3;
        Consensus {
            validators,
            threshold,
        }
    }

    pub fn sign_block(&self, block: &Block) -> Vec<String> {
        self.validators
            .iter()
            .map(|v| {
                let data = format!("{}{}", v.id, hash_block(block));
                let mut hasher = Sha256::new();
                hasher.update(data);
                format!("{:x}", hasher.finalize())
            })
            .collect()
    }

    pub fn verify_quorum(&self, signatures: &[String]) -> bool {
        signatures.len() >= self.threshold
    }

    pub fn propose_block(
        &self,
        height: u64,
        prev_hash: String,
        txs: Vec<Transaction>,
        proposer: String,
    ) -> Block {
        let merkle_root = if let Some((_, root_hash)) = crate::merkle::build_merkle_tree(&txs) {
            root_hash
        } else {
            String::new()
        };
        let header = Header {
            height,
            prev_hash,
            merkle_root,
            proposer,
            timestamp: Utc::now(),
        };
        Block { header, txs }
    }
}
