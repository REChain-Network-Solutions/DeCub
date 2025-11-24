package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rechain/rechain/internal/api"
	"github.com/rechain/rechain/internal/cas"
	"github.com/rechain/rechain/internal/consensus"
	"github.com/rechain/rechain/internal/gcl"
	"github.com/rechain/rechain/internal/gossip"
	"github.com/rechain/rechain/internal/security"
	"github.com/rechain/rechain/internal/storage"
	"github.com/spf13/viper"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "./config/config.yaml", "Path to configuration file")
	flag.Parse()

	// Initialize configuration
	if err := initConfig(*configFile); err != nil {
		log.Fatalf("Error initializing config: %v", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize storage
	store, err := storage.NewBadgerStore(viper.GetString("storage.path"))
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize security
	keyManager, err := security.NewKeyManager()
	if err != nil {
		log.Fatalf("Failed to initialize security: %v", err)
	}

	// Initialize CAS
	casStore, err := cas.NewCAS(
		viper.GetString("cas.endpoint"),
		viper.GetString("cas.access_key"),
		viper.GetString("cas.secret_key"),
		viper.GetString("cas.bucket"),
		viper.GetBool("cas.use_ssl"),
	)
	if err != nil {
		log.Fatalf("Failed to initialize CAS: %v", err)
	}

	// Initialize gossip protocol
	gossipProto, err := gossip.NewGossipProtocol(viper.GetString("network.listen_address"))
	if err != nil {
		log.Fatalf("Failed to initialize gossip: %v", err)
	}
	defer gossipProto.Stop()

	// Add bootstrap peers
	for _, peerAddr := range viper.GetStringSlice("network.bootstrap") {
		if err := gossipProto.AddPeer(peerAddr); err != nil {
			log.Printf("Failed to add bootstrap peer %s: %v", peerAddr, err)
		}
	}

	// Initialize consensus
	consensusEngine, err := consensus.NewConsensus(store, nil) // P2P will be integrated later
	if err != nil {
		log.Fatalf("Failed to initialize consensus: %v", err)
	}
	defer consensusEngine.Stop()

	// Initialize GCL node (legacy, will be replaced by gossip)
	gclNode, err := gcl.NewNode(store)
	if err != nil {
		log.Fatalf("Failed to initialize GCL node: %v", err)
	}

	// Start GCL node
	if err := gclNode.Start(ctx); err != nil {
		log.Fatalf("Failed to start GCL node: %v", err)
	}
	defer gclNode.Stop()

	// Initialize API servers
	restServer := api.NewServer(consensusEngine, store, casStore, gossipProto, keyManager)
	grpcServer, err := api.NewGRPCServer(restServer)
	if err != nil {
		log.Fatalf("Failed to create gRPC server: %v", err)
	}

	// Start API servers
	go func() {
		restAddr := viper.GetString("api.rest_address")
		log.Printf("Starting REST API server on %s", restAddr)
		if err := restServer.Start(restAddr); err != nil {
			log.Printf("REST API server error: %v", err)
		}
	}()

	go func() {
		grpcAddr := viper.GetString("api.grpc_address")
		log.Printf("Starting gRPC API server on %s", grpcAddr)
		if err := grpcServer.Start(grpcAddr); err != nil {
			log.Printf("gRPC API server error: %v", err)
		}
	}()

	// Start gossip protocol
	if err := gossipProto.Start(); err != nil {
		log.Fatalf("Failed to start gossip protocol: %v", err)
	}

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Shutdown sequence
	log.Println("Shutting down...")

	if err := grpcServer.Stop(); err != nil {
		log.Printf("Error stopping gRPC server: %v", err)
	}

	if err := restServer.Stop(); err != nil {
		log.Printf("Error stopping REST server: %v", err)
	}

	if err := consensusEngine.Stop(); err != nil {
		log.Printf("Error stopping consensus: %v", err)
	}

	if err := gclNode.Stop(); err != nil {
		log.Printf("Error stopping GCL node: %v", err)
	}
}

func initConfig(configFile string) error {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	// Set default values
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
		log.Printf("Config file not found at %s, using defaults", configFile)
	}

	// Override with environment variables
	viper.SetEnvPrefix("RECHAIN")
	viper.AutomaticEnv()

	return nil
}

func setDefaults() {
	// Node defaults
	viper.SetDefault("node.id", "")
	viper.SetDefault("node.data_dir", "./data")
	viper.SetDefault("node.log_level", "info")
	viper.SetDefault("node.enable_metrics", true)

	// Network defaults
	viper.SetDefault("network.listen_address", "/ip4/0.0.0.0/tcp/26656")
	viper.SetDefault("network.bootstrap", []string{})
	viper.SetDefault("network.max_peers", 50)

	// Storage defaults
	viper.SetDefault("storage.engine", "badger")
	viper.SetDefault("storage.path", "./data/chain")
	viper.SetDefault("storage.cache_size", 100*1024*1024)
	viper.SetDefault("storage.sync", true)

	// Consensus defaults
	viper.SetDefault("consensus.type", "bft")
	viper.SetDefault("consensus.block_time", "1s")
	viper.SetDefault("consensus.timeout_propose", "3s")
	viper.SetDefault("consensus.timeout_prevote", "1s")
	viper.SetDefault("consensus.timeout_precommit", "1s")
	viper.SetDefault("consensus.timeout_commit", "1s")

	// CAS defaults
	viper.SetDefault("cas.endpoint", "http://localhost:9000")
	viper.SetDefault("cas.access_key", "rechain")
	viper.SetDefault("cas.secret_key", "rechain123")
	viper.SetDefault("cas.bucket", "rechain-cas")
	viper.SetDefault("cas.use_ssl", false)
	viper.SetDefault("cas.chunk_size", 64*1024*1024)
	viper.SetDefault("cas.max_retries", 3)

	// Gossip defaults
	viper.SetDefault("gossip.enabled", true)
	viper.SetDefault("gossip.fanout", 3)
	viper.SetDefault("gossip.interval", "1s")
	viper.SetDefault("gossip.anti_entropy_interval", "30s")
	viper.SetDefault("gossip.message_ttl", 10)

	// API defaults
	viper.SetDefault("api.enabled", true)
	viper.SetDefault("api.rest_address", "0.0.0.0:1317")
	viper.SetDefault("api.grpc_address", "0.0.0.0:9090")
	viper.SetDefault("api.enable_cors", true)
	viper.SetDefault("api.cors_allowed_origins", []string{"*"})
	viper.SetDefault("api.rate_limiting_enabled", true)
	viper.SetDefault("api.rate_limit_rps", 100)

	// Security defaults
	viper.SetDefault("security.tls_enabled", true)
	viper.SetDefault("security.cert_file", "./certs/server.crt")
	viper.SetDefault("security.key_file", "./certs/server.key")
	viper.SetDefault("security.ca_file", "./certs/ca.crt")
	viper.SetDefault("security.client_cert_required", false)
	viper.SetDefault("security.hsm_enabled", false)
	viper.SetDefault("security.hsm_address", "tcp://localhost:12345")
	viper.SetDefault("security.audit_enabled", true)

	// Monitoring defaults
	viper.SetDefault("monitoring.prometheus_enabled", true)
	viper.SetDefault("monitoring.prometheus_address", "0.0.0.0:9091")
	viper.SetDefault("monitoring.metrics_prefix", "rechain")
	viper.SetDefault("monitoring.health_check_enabled", true)

	// Logging defaults
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")
	viper.SetDefault("logging.max_size", 100)
	viper.SetDefault("logging.max_age", 30)
	viper.SetDefault("logging.max_backups", 5)
	viper.SetDefault("logging.compress", true)

	// Database defaults
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.connection_string", "./data/metadata.db")
	viper.SetDefault("database.max_open_conns", 10)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "1h")

	// Backup defaults
	viper.SetDefault("backup.enabled", true)
	viper.SetDefault("backup.interval", "24h")
	viper.SetDefault("backup.retention", "168h")
	viper.SetDefault("backup.directory", "./backups")
	viper.SetDefault("backup.remote_enabled", false)

	// Development defaults
	viper.SetDefault("development.debug", false)
	viper.SetDefault("development.pprof_enabled", false)
	viper.SetDefault("development.mock_services", false)
}
