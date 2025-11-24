package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// Initialize consensus with mock validators
	validators := []Validator{
		{ID: "val1", PubKey: "pub1"},
		{ID: "val2", PubKey: "pub2"},
		{ID: "val3", PubKey: "pub3"},
	}
	cons = NewConsensus(validators)

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

	http.HandleFunc("/gcl/tx", SubmitTx)
	http.HandleFunc("/gcl/block/", GetBlock)
	http.HandleFunc("/gcl/proof/", GetProof)

	fmt.Println("Starting GCL server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
