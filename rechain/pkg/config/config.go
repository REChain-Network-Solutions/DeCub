package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the REChain node
type Config struct {
	Node     NodeConfig     `mapstructure:"node"`
	Network  NetworkConfig  `mapstructure:"network"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Consensus ConsensusConfig `mapstructure:"consensus"`
	CAS      CASConfig      `mapstructure:"cas"`
	Gossip   GossipConfig   `mapstructure:"gossip"`
	API      APIConfig      `mapstructure:"api"`
	Security SecurityConfig `mapstructure:"security"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
}

// NodeConfig holds node-specific configuration
type NodeConfig struct {
	ID       string `mapstructure:"id"`
	DataDir  string `mapstructure:"data_dir"`
	LogLevel string `mapstructure:"log_level"`
}

// NetworkConfig holds network configuration
type NetworkConfig struct {
	ListenAddress string   `mapstructure:"listen_address"`
	Bootstrap     []string `mapstructure:"bootstrap"`
	MaxPeers      int      `mapstructure:"max_peers"`
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	Engine     string `mapstructure:"engine"`
	Path       string `mapstructure:"path"`
	CacheSize  int64  `mapstructure:"cache_size"`
	Sync       bool   `mapstructure:"sync"`
}

// ConsensusConfig holds consensus configuration
type ConsensusConfig struct {
	Type        string        `mapstructure:"type"`
	BlockTime   time.Duration `mapstructure:"block_time"`
	TimeoutPropose time.Duration `mapstructure:"timeout_propose"`
	TimeoutPrevote time.Duration `mapstructure:"timeout_prevote"`
	TimeoutPrecommit time.Duration `mapstructure:"timeout_precommit"`
	TimeoutCommit time.Duration `mapstructure:"timeout_commit"`
}

// CASConfig holds CAS configuration
type CASConfig struct {
	Endpoint   string `mapstructure:"endpoint"`
	Bucket     string `mapstructure:"bucket"`
	AccessKey  string `mapstructure:"access_key"`
	SecretKey  string `mapstructure:"secret_key"`
	ChunkSize  int64  `mapstructure:"chunk_size"`
	UseSSL     bool   `mapstructure:"use_ssl"`
}

// GossipConfig holds gossip configuration
type GossipConfig struct {
	Port            int           `mapstructure:"port"`
	BootstrapPeers  []string      `mapstructure:"bootstrap_peers"`
	Fanout          int           `mapstructure:"fanout"`
	GossipInterval  time.Duration `mapstructure:"gossip_interval"`
	AntiEntropyInterval time.Duration `mapstructure:"anti_entropy_interval"`
}

// APIConfig holds API configuration
type APIConfig struct {
	REST RESTConfig `mapstructure:"rest"`
	GRPC GRPCConfig `mapstructure:"grpc"`
}

// RESTConfig holds REST API configuration
type RESTConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	Address string   `mapstructure:"address"`
	CORS    []string `mapstructure:"cors"`
}

// GRPCConfig holds gRPC API configuration
type GRPCConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Address string `mapstructure:"address"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	TLSEnabled    bool   `mapstructure:"tls_enabled"`
	CertFile      string `mapstructure:"cert_file"`
	KeyFile       string `mapstructure:"key_file"`
	CAFile        string `mapstructure:"ca_file"`
	EncryptData   bool   `mapstructure:"encrypt_data"`
	SignTxs       bool   `mapstructure:"sign_txs"`
	HSMEnabled    bool   `mapstructure:"hsm_enabled"`
	AuditLogPath  string `mapstructure:"audit_log_path"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Address string `mapstructure:"address"`
	Path    string `mapstructure:"path"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Node: NodeConfig{
			ID:       "",
			DataDir:  "./data",
			LogLevel: "info",
		},
		Network: NetworkConfig{
			ListenAddress: "tcp://0.0.0.0:26656",
			Bootstrap:     []string{},
			MaxPeers:      50,
		},
		Storage: StorageConfig{
			Engine:    "badger",
			Path:      "",
			CacheSize: 100 * 1024 * 1024, // 100MB
			Sync:      true,
		},
		Consensus: ConsensusConfig{
			Type:             "bft",
			BlockTime:        1 * time.Second,
			TimeoutPropose:   3 * time.Second,
			TimeoutPrevote:   1 * time.Second,
			TimeoutPrecommit: 1 * time.Second,
			TimeoutCommit:    1 * time.Second,
		},
		CAS: CASConfig{
			Endpoint:  "localhost:9000",
			Bucket:    "rechain-objects",
			AccessKey: "rechain",
			SecretKey: "rechain123",
			ChunkSize: 64 * 1024 * 1024, // 64MB
			UseSSL:    false,
		},
		Gossip: GossipConfig{
			Port:               26656,
			BootstrapPeers:     []string{},
			Fanout:             3,
			GossipInterval:     100 * time.Millisecond,
			AntiEntropyInterval: 10 * time.Second,
		},
		API: APIConfig{
			REST: RESTConfig{
				Enabled: true,
				Address: "0.0.0.0:1317",
				CORS:    []string{"*"},
			},
			GRPC: GRPCConfig{
				Enabled: true,
				Address: "0.0.0.0:9090",
			},
		},
		Security: SecurityConfig{
			TLSEnabled:   false,
			CertFile:     "",
			KeyFile:      "",
			CAFile:       "",
			EncryptData:  true,
			SignTxs:      true,
			HSMEnabled:   false,
			AuditLogPath: "./logs/audit.log",
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Address: "0.0.0.0:9091",
			Path:    "/metrics",
		},
	}
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	viper := viper.New()

	// Set defaults
	viper.SetDefault("node.data_dir", cfg.Node.DataDir)
	viper.SetDefault("node.log_level", cfg.Node.LogLevel)
	viper.SetDefault("network.listen_address", cfg.Network.ListenAddress)
	viper.SetDefault("network.max_peers", cfg.Network.MaxPeers)
	viper.SetDefault("storage.engine", cfg.Storage.Engine)
	viper.SetDefault("storage.cache_size", cfg.Storage.CacheSize)
	viper.SetDefault("storage.sync", cfg.Storage.Sync)
	viper.SetDefault("consensus.type", cfg.Consensus.Type)
	viper.SetDefault("consensus.block_time", cfg.Consensus.BlockTime)
	viper.SetDefault("consensus.timeout_propose", cfg.Consensus.TimeoutPropose)
	viper.SetDefault("consensus.timeout_prevote", cfg.Consensus.TimeoutPrevote)
	viper.SetDefault("consensus.timeout_precommit", cfg.Consensus.TimeoutPrecommit)
	viper.SetDefault("consensus.timeout_commit", cfg.Consensus.TimeoutCommit)
	viper.SetDefault("cas.endpoint", cfg.CAS.Endpoint)
	viper.SetDefault("cas.bucket", cfg.CAS.Bucket)
	viper.SetDefault("cas.access_key", cfg.CAS.AccessKey)
	viper.SetDefault("cas.secret_key", cfg.CAS.SecretKey)
	viper.SetDefault("cas.chunk_size", cfg.CAS.ChunkSize)
	viper.SetDefault("cas.use_ssl", cfg.CAS.UseSSL)
	viper.SetDefault("gossip.port", cfg.Gossip.Port)
	viper.SetDefault("gossip.fanout", cfg.Gossip.Fanout)
	viper.SetDefault("gossip.gossip_interval", cfg.Gossip.GossipInterval)
	viper.SetDefault("gossip.anti_entropy_interval", cfg.Gossip.AntiEntropyInterval)
	viper.SetDefault("api.rest.enabled", cfg.API.REST.Enabled)
	viper.SetDefault("api.rest.address", cfg.API.REST.Address)
	viper.SetDefault("api.rest.cors", cfg.API.REST.CORS)
	viper.SetDefault("api.grpc.enabled", cfg.API.GRPC.Enabled)
	viper.SetDefault("api.grpc.address", cfg.API.GRPC.Address)
	viper.SetDefault("security.tls_enabled", cfg.Security.TLSEnabled)
	viper.SetDefault("security.encrypt_data", cfg.Security.EncryptData)
	viper.SetDefault("security.sign_txs", cfg.Security.SignTxs)
	viper.SetDefault("security.hsm_enabled", cfg.Security.HSMEnabled)
	viper.SetDefault("security.audit_log_path", cfg.Security.AuditLogPath)
	viper.SetDefault("logging.level", cfg.Logging.Level)
	viper.SetDefault("logging.format", cfg.Logging.Format)
	viper.SetDefault("logging.output", cfg.Logging.Output)
	viper.SetDefault("logging.max_size", cfg.Logging.MaxSize)
	viper.SetDefault("logging.max_backups", cfg.Logging.MaxBackups)
	viper.SetDefault("logging.max_age", cfg.Logging.MaxAge)
	viper.SetDefault("metrics.enabled", cfg.Metrics.Enabled)
	viper.SetDefault("metrics.address", cfg.Metrics.Address)
	viper.SetDefault("metrics.path", cfg.Metrics.Path)

	// Environment variable bindings
	viper.SetEnvPrefix("RECHAIN")
	viper.AutomaticEnv()

	// Set config file
	if configPath != "" {
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal into config struct
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}
