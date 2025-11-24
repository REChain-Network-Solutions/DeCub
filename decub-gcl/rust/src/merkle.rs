use crate::types::{MerkleNode, MerkleProof, Transaction, hash_transaction};
use sha2::{Digest, Sha256};

pub fn build_merkle_tree(txs: &[Transaction]) -> Option<(MerkleNode, String)> {
    if txs.is_empty() {
        return None;
    }

    let mut nodes: Vec<MerkleNode> = txs
        .iter()
        .map(|tx| MerkleNode {
            hash: hash_transaction(tx),
            left: None,
            right: None,
        })
        .collect();

    while nodes.len() > 1 {
        let mut new_nodes = Vec::new();
        for chunk in nodes.chunks(2) {
            let left = &chunk[0];
            let right = if chunk.len() == 2 { &chunk[1] } else { &chunk[0] };
            let combined = format!("{}{}", left.hash, right.hash);
            let mut hasher = Sha256::new();
            hasher.update(combined);
            let hash = format!("{:x}", hasher.finalize());
            new_nodes.push(MerkleNode {
                hash,
                left: Some(Box::new(left.clone())),
                right: Some(Box::new(right.clone())),
            });
        }
        nodes = new_nodes;
    }

    let root = nodes.into_iter().next().unwrap();
    let root_hash = root.hash.clone();
    Some((root, root_hash))
}

pub fn generate_merkle_proof(root: &MerkleNode, index: usize) -> MerkleProof {
    let mut proof = MerkleProof {
        hashes: Vec::new(),
        index,
    };
    let mut current = root;
    let mut idx = index;
    while current.left.is_some() || current.right.is_some() {
        if idx % 2 == 0 {
            if let Some(right) = &current.right {
                proof.hashes.push(right.hash.clone());
            }
        } else {
            if let Some(left) = &current.left {
                proof.hashes.push(left.hash.clone());
            }
        }
        if idx % 2 == 0 {
            if let Some(left) = &current.left {
                current = left;
            } else {
                break;
            }
        } else {
            if let Some(right) = &current.right {
                current = right;
            } else {
                break;
            }
        }
        idx /= 2;
    }
    proof
}
