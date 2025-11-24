package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

// GossipConfig holds configuration for the gossip synchronization layer
type GossipConfig struct {
	// Node identification
	NodeID string `json:"node_id"`

	// Network configuration
	ListenAddr    string   `json:"listen_addr"`
	InitialPeers  []string `json:"initial_peers"`
	AdvertiseAddr string   `json:"advertise_addr"`

	// Gossip intervals
	GossipInterval       time.Duration `json:"gossip_interval"`
	AntiEntropyInterval  time.Duration `json:"anti_entropy_interval"`
	SyncInterval         time.Duration `json:"sync_interval"`

	// Merkle tree configuration
	MerkleTreeDepth int `json:"merkle_tree_depth"`

	// Catalog service configuration
	CatalogAddr string `json:"catalog_addr"`

	// TLS configuration
	EnableTLS     bool   `json:"enable_tls"`
	CertFile      string `json:"cert_file"`
	KeyFile       string `json:"key_file"`
	CACertFile    string `json:"ca_cert_file"`

	// Logging
	LogLevel string `json:"log_level"`
}

// DefaultConfig returns a default gossip configuration
func DefaultConfig() *GossipConfig {
	nodeID := os.Getenv("DECUB_NODE_ID")
	if nodeID == "" {
		nodeID = fmt.Sprintf("node-%d", time.Now().Unix())
	}

	return &GossipConfig{
		NodeID:               nodeID,
		ListenAddr:           "/ip4/0.0.0.0/tcp/0",
		InitialPeers:         []string{},
		AdvertiseAddr:        "",
		GossipInterval:       5 * time.Second,
		AntiEntropyInterval:  30 * time.Second,
		SyncInterval:         60 * time.Second,
		MerkleTreeDepth:      16,
		CatalogAddr:          "http://localhost:8080",
		EnableTLS:            false,
		CertFile:             "",
		KeyFile:             "",
		CACertFile:           "",
		LogLevel:             "info",
	}
}

// LoadConfigFromFile loads configuration from a JSON file
func LoadConfigFromFile(filename string) (*GossipConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	config := DefaultConfig()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	// Override with environment variables
	config.overrideFromEnv()

	return config, nil
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() *GossipConfig {
	config := DefaultConfig()
	config.overrideFromEnv()
	return config
}

// overrideFromEnv overrides configuration with environment variables
func (c *GossipConfig) overrideFromEnv() {
	if nodeID := os.Getenv("DECUB_NODE_ID"); nodeID != "" {
		c.NodeID = nodeID
	}
	if listenAddr := os.Getenv("DECUB_LISTEN_ADDR"); listenAddr != "" {
		c.ListenAddr = listenAddr
	}
	if advertiseAddr := os.Getenv("DECUB_ADVERTISE_ADDR"); advertiseAddr != "" {
		c.AdvertiseAddr = advertiseAddr
	}
	if initialPeers := os.Getenv("DECUB_INITIAL_PEERS"); initialPeers != "" {
		// Parse comma-separated list
		c.InitialPeers = parseCommaSeparatedList(initialPeers)
	}
	if gossipInterval := os.Getenv("DECUB_GOSSIP_INTERVAL"); gossipInterval != "" {
		if d, err := time.ParseDuration(gossipInterval); err == nil {
			c.GossipInterval = d
		}
	}
	if antiEntropyInterval := os.Getenv("DECUB_ANTI_ENTROPY_INTERVAL"); antiEntropyInterval != "" {
		if d, err := time.ParseDuration(antiEntropyInterval); err == nil {
			c.AntiEntropyInterval = d
		}
	}
	if syncInterval := os.Getenv("DECUB_SYNC_INTERVAL"); syncInterval != "" {
		if d, err := time.ParseDuration(syncInterval); err == nil {
			c.SyncInterval = d
		}
	}
	if merkleDepth := os.Getenv("DECUB_MERKLE_DEPTH"); merkleDepth != "" {
		if depth, err := strconv.Atoi(merkleDepth); err == nil {
			c.MerkleTreeDepth = depth
		}
	}
	if catalogAddr := os.Getenv("DECUB_CATALOG_ADDR"); catalogAddr != "" {
		c.CatalogAddr = catalogAddr
	}
	if enableTLS := os.Getenv("DECUB_ENABLE_TLS"); enableTLS != "" {
		if enable, err := strconv.ParseBool(enableTLS); err == nil {
			c.EnableTLS = enable
		}
	}
	if certFile := os.Getenv("DECUB_CERT_FILE"); certFile != "" {
		c.CertFile = certFile
	}
	if keyFile := os.Getenv("DECUB_KEY_FILE"); keyFile != "" {
		c.KeyFile = keyFile
	}
	if caCertFile := os.Getenv("DECUB_CA_CERT_FILE"); caCertFile != "" {
		c.CACertFile = caCertFile
	}
	if logLevel := os.Getenv("DECUB_LOG_LEVEL"); logLevel != "" {
		c.LogLevel = logLevel
	}
}

// Validate checks if the configuration is valid
func (c *GossipConfig) Validate() error {
	if c.NodeID == "" {
		return fmt.Errorf("node_id cannot be empty")
	}
	if c.ListenAddr == "" {
		return fmt.Errorf("listen_addr cannot be empty")
	}
	if c.GossipInterval <= 0 {
		return fmt.Errorf("gossip_interval must be positive")
	}
	if c.AntiEntropyInterval <= 0 {
		return fmt.Errorf("anti_entropy_interval must be positive")
	}
	if c.SyncInterval <= 0 {
		return fmt.Errorf("sync_interval must be positive")
	}
	if c.MerkleTreeDepth <= 0 {
		return fmt.Errorf("merkle_tree_depth must be positive")
	}
	if c.CatalogAddr == "" {
		return fmt.Errorf("catalog_addr cannot be empty")
	}
	if c.EnableTLS {
		if c.CertFile == "" || c.KeyFile == "" {
			return fmt.Errorf("cert_file and key_file are required when TLS is enabled")
		}
	}
	return nil
}

// parseCommaSeparatedList parses a comma-separated string into a slice
func parseCommaSeparatedList(s string) []string {
	var result []string
	for _, item := range []string{s} {
		// Simple split by comma (could be enhanced with proper CSV parsing)
		parts := []string{}
		current := ""
		for _, r := range item {
			if r == ',' {
				if current != "" {
					parts = append(parts, current)
					current = ""
				}
			} else {
				current += string(r)
			}
		}
		if current != "" {
			parts = append(parts, current)
		}
		result = append(result, parts...)
	}
	return result
}

// String returns a string representation of the config (without sensitive data)
func (c *GossipConfig) String() string {
	return fmt.Sprintf("GossipConfig{NodeID: %s, ListenAddr: %s, Peers: %v, GossipInterval: %v, AntiEntropyInterval: %v}",
		c.NodeID, c.ListenAddr, c.InitialPeers, c.GossipInterval, c.AntiEntropyInterval)
}
