package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type Config struct {
	ControlPlaneURL string `yaml:"control_plane_url" mapstructure:"control_plane_url"`
	GCLURL          string `yaml:"gcl_url" mapstructure:"gcl_url"`
	CatalogURL      string `yaml:"catalog_url" mapstructure:"catalog_url"`
	GossipURL       string `yaml:"gossip_url" mapstructure:"gossip_url"`
	StorageURL      string `yaml:"storage_url" mapstructure:"storage_url"`
	ClusterID       string `yaml:"cluster_id" mapstructure:"cluster_id"`
	Timeout         int    `yaml:"timeout" mapstructure:"timeout"`
}

type SnapshotMetadata struct {
	ID       string                 `json:"id"`
	Size     int64                  `json:"size"`
	Chunks   int                    `json:"chunks"`
	Hashes   []string               `json:"hashes"`
	Created  time.Time              `json:"created"`
	Metadata map[string]interface{} `json:"metadata"`
}

type Transaction struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Signature string                 `json:"signature"`
}

type CommitProof struct {
	TxHash     string   `json:"tx_hash"`
	BlockHash  string   `json:"block_hash"`
	Height     int64    `json:"height"`
	Signatures []string `json:"signatures"`
}

var config Config
var cfgFile string

func main() {
	cobra.OnInitialize(initConfig)

	rootCmd := &cobra.Command{
		Use:   "decubectl",
		Short: "DeCube CLI tool",
		Long:  `A command-line tool for managing DeCube clusters`,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.decube/config.yaml)")

	// Snapshot commands
	snapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Manage snapshots",
	}
	snapshotCreateCmd := &cobra.Command{
		Use:   "create <id> <etcd-dir> <volume-dir>",
		Short: "Create a snapshot",
		Args:  cobra.ExactArgs(3),
		Run:   snapshotCreate,
	}
	snapshotRestoreCmd := &cobra.Command{
		Use:   "restore <id> <restore-dir>",
		Short: "Restore a snapshot",
		Args:  cobra.ExactArgs(2),
		Run:   snapshotRestore,
	}
	snapshotCmd.AddCommand(snapshotCreateCmd, snapshotRestoreCmd)

	// GCL commands
	gclCmd := &cobra.Command{
		Use:   "gcl",
		Short: "Global Consensus Layer operations",
	}
	gclTxCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transaction operations",
	}
	gclTxPublishCmd := &cobra.Command{
		Use:   "publish <type> <payload-json>",
		Short: "Publish a transaction",
		Args:  cobra.ExactArgs(2),
		Run:   gclTxPublish,
	}
	gclTxProofCmd := &cobra.Command{
		Use:   "proof <tx-hash>",
		Short: "Get transaction proof",
		Args:  cobra.ExactArgs(1),
		Run:   gclTxProof,
	}
	gclTxCmd.AddCommand(gclTxPublishCmd, gclTxProofCmd)
	gclCmd.AddCommand(gclTxCmd)

	// CRDT commands
	crdtCmd := &cobra.Command{
		Use:   "crdt",
		Short: "CRDT operations",
	}
	crdtMergeCmd := &cobra.Command{
		Use:   "merge <type> <key> <value>",
		Short: "Merge CRDT value",
		Args:  cobra.ExactArgs(3),
		Run:   crdtMerge,
	}
	crdtCmd.AddCommand(crdtMergeCmd)

	// Gossip commands
	gossipCmd := &cobra.Command{
		Use:   "gossip",
		Short: "Gossip protocol operations",
	}
	gossipSyncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Trigger gossip synchronization",
		Run:   gossipSync,
	}
	gossipCmd.AddCommand(gossipSyncCmd)

	// Status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show cluster status",
		Run:   showStatus,
	}

	rootCmd.AddCommand(snapshotCmd, gclCmd, crdtCmd, gossipCmd, statusCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(filepath.Join(home, ".decube"))
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Unable to decode config: %v", err)
	}

	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 30
	}
}

func httpClient() *http.Client {
	return &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}
}

func makeRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := httpClient()
	return client.Do(req)
}

func snapshotCreate(cmd *cobra.Command, args []string) {
	id := args[0]
	etcdDir := args[1]
	volumeDir := args[2]

	fmt.Printf("Creating snapshot %s from %s and %s...\n", id, etcdDir, volumeDir)

	// Call control plane to create snapshot
	payload := map[string]interface{}{
		"id":        id,
		"etcd_dir":  etcdDir,
		"volume_dir": volumeDir,
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := makeRequest("POST", config.ControlPlaneURL+"/api/v1/snapshots", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to create snapshot: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Snapshot creation failed: %s", string(body))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Printf("Snapshot created successfully: %v\n", result)
}

func snapshotRestore(cmd *cobra.Command, args []string) {
	id := args[0]
	restoreDir := args[1]

	fmt.Printf("Restoring snapshot %s to %s...\n", id, restoreDir)

	// Call control plane to restore snapshot
	payload := map[string]interface{}{
		"id":          id,
		"restore_dir": restoreDir,
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := makeRequest("POST", config.ControlPlaneURL+"/api/v1/snapshots/restore", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to restore snapshot: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Snapshot restore failed: %s", string(body))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Printf("Snapshot restored successfully: %v\n", result)
}

func gclTxPublish(cmd *cobra.Command, args []string) {
	txType := args[0]
	payloadJSON := args[1]

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		log.Fatalf("Invalid JSON payload: %v", err)
	}

	fmt.Printf("Publishing %s transaction...\n", txType)

	tx := Transaction{
		Type:    txType,
		Payload: payload,
		// In real implementation, sign the transaction
		Signature: "dummy-signature",
	}

	jsonData, _ := json.Marshal(tx)
	resp, err := makeRequest("POST", config.GCLURL+"/api/v1/transactions", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to publish transaction: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Transaction publish failed: %s", string(body))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Printf("Transaction published: %v\n", result)
}

func gclTxProof(cmd *cobra.Command, args []string) {
	txHash := args[0]

	fmt.Printf("Getting proof for transaction %s...\n", txHash)

	resp, err := makeRequest("GET", config.GCLURL+"/api/v1/transactions/"+txHash+"/proof", nil)
	if err != nil {
		log.Fatalf("Failed to get proof: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Proof retrieval failed: %s", string(body))
	}

	var proof CommitProof
	json.NewDecoder(resp.Body).Decode(&proof)

	fmt.Printf("Transaction Proof:\n")
	fmt.Printf("  Tx Hash: %s\n", proof.TxHash)
	fmt.Printf("  Block Hash: %s\n", proof.BlockHash)
	fmt.Printf("  Height: %d\n", proof.Height)
	fmt.Printf("  Signatures: %d\n", len(proof.Signatures))
}

func crdtMerge(cmd *cobra.Command, args []string) {
	crdtType := args[0]
	key := args[1]
	value := args[2]

	fmt.Printf("Merging %s CRDT: %s = %s\n", crdtType, key, value)

	payload := map[string]interface{}{
		"type":  crdtType,
		"key":   key,
		"value": value,
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := makeRequest("POST", config.CatalogURL+"/api/v1/crdt/merge", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to merge CRDT: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("CRDT merge failed: %s", string(body))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Printf("CRDT merged successfully: %v\n", result)
}

func gossipSync(cmd *cobra.Command, args []string) {
	fmt.Println("Triggering gossip synchronization...")

	resp, err := makeRequest("POST", config.GossipURL+"/api/v1/sync", nil)
	if err != nil {
		log.Fatalf("Failed to trigger sync: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Gossip sync failed: %s", string(body))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Printf("Gossip sync completed: %v\n", result)
}

func showStatus(cmd *cobra.Command, args []string) {
	fmt.Println("DeCube Cluster Status")
	fmt.Println("====================")

	// Get control plane status
	fmt.Println("\nControl Plane:")
	resp, err := makeRequest("GET", config.ControlPlaneURL+"/api/v1/status", nil)
	if err == nil && resp.StatusCode == http.StatusOK {
		var status map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&status)
		for k, v := range status {
			fmt.Printf("  %s: %v\n", k, v)
		}
		resp.Body.Close()
	} else {
		fmt.Println("  Status: Unavailable")
	}

	// Get GCL status
	fmt.Println("\nGlobal Consensus Layer:")
	resp, err = makeRequest("GET", config.GCLURL+"/api/v1/status", nil)
	if err == nil && resp.StatusCode == http.StatusOK {
		var status map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&status)
		for k, v := range status {
			fmt.Printf("  %s: %v\n", k, v)
		}
		resp.Body.Close()
	} else {
		fmt.Println("  Status: Unavailable")
	}

	// Get catalog status
	fmt.Println("\nCatalog Service:")
	resp, err = makeRequest("GET", config.CatalogURL+"/api/v1/status", nil)
	if err == nil && resp.StatusCode == http.StatusOK {
		var status map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&status)
		for k, v := range status {
			fmt.Printf("  %s: %v\n", k, v)
		}
		resp.Body.Close()
	} else {
		fmt.Println("  Status: Unavailable")
	}

	// Get gossip status
	fmt.Println("\nGossip Service:")
	resp, err = makeRequest("GET", config.GossipURL+"/api/v1/status", nil)
	if err == nil && resp.StatusCode == http.StatusOK {
		var status map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&status)
		for k, v := range status {
			fmt.Printf("  %s: %v\n", k, v)
		}
		resp.Body.Close()
	} else {
		fmt.Println("  Status: Unavailable")
	}

	// Get storage status
	fmt.Println("\nStorage Service:")
	resp, err = makeRequest("GET", config.StorageURL+"/api/v1/status", nil)
	if err == nil && resp.StatusCode == http.StatusOK {
		var status map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&status)
		for k, v := range status {
			fmt.Printf("  %s: %v\n", k, v)
		}
		resp.Body.Close()
	} else {
		fmt.Println("  Status: Unavailable")
	}
}
