package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the DeCube node
type Config struct {
	Node        NodeConfig        `mapstructure:"node"`
	Etcd        EtcdConfig        `mapstructure:"etcd"`
	API         APIConfig         `mapstructure:"api"`
	Replication ReplicationConfig `mapstructure:"replication"`
	Snapshot    SnapshotConfig    `mapstructure:"snapshot"`
	Logging     LoggingConfig     `mapstructure:"logging"`
	Security    SecurityConfig    `mapstructure:"security"`
}

// NodeConfig holds node-specific configuration
type NodeConfig struct {
	ID           string   `mapstructure:"id"`
	DataDir      string   `mapstructure:"data_dir"`
	ListenAddress string  `mapstructure:"listen_address"`
	PeerAddresses []string `mapstructure:"peer_addresses"`
}

// EtcdConfig holds etcd configuration
type EtcdConfig struct {
	Name               string `mapstructure:"name"`
	DataDir            string `mapstructure:"data_dir"`
	WalDir             string `mapstructure:"wal_dir"`
	SnapshotCount      uint64 `mapstructure:"snapshot_count"`
	HeartbeatInterval  int    `mapstructure:"heartbeat_interval"`
	ElectionTimeout    int    `mapstructure:"election_timeout"`
	MaxSnapshots       uint   `mapstructure:"max_snapshots"`
	MaxWals            uint   `mapstructure:"max_wals"`
	AutoCompactionRetention string `mapstructure:"auto_compaction_retention"`
	QuotaBackendBytes  int64  `mapstructure:"quota_backend_bytes"`
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
	CORS    []string `mapstructure:"cors_origins"`
}

// GRPCConfig holds gRPC API configuration
type GRPCConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Address string `mapstructure:"address"`
}

// ReplicationConfig holds replication configuration
type ReplicationConfig struct {
	Enabled      bool          `mapstructure:"enabled"`
	PeerTimeout  time.Duration `mapstructure:"peer_timeout"`
	RetryInterval time.Duration `mapstructure:"retry_interval"`
	MaxRetries   int           `mapstructure:"max_retries"`
}

// SnapshotConfig holds snapshot configuration
type SnapshotConfig struct {
	Enabled       bool          `mapstructure:"enabled"`
	Interval      time.Duration `mapstructure:"interval"`
	RetentionCount int          `mapstructure:"retention_count"`
	Compression   bool          `mapstructure:"compression"`
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

// SecurityConfig holds security configuration
type SecurityConfig struct {
	TLSEnabled bool   `mapstructure:"tls_enabled"`
	CertFile   string `mapstructure:"cert_file"`
	KeyFile    string `mapstructure:"key_file"`
	CAFile     string `mapstructure:"ca_file"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Node: NodeConfig{
			ID:            "node-1",
			DataDir:       "/var/lib/decube",
			ListenAddress: "0.0.0.0:2379",
			PeerAddresses: []string{"node-1:2380", "node-2:2380", "node-3:2380"},
		},
		Etcd: EtcdConfig{
			Name:                   "node-1",
			DataDir:                "/var/lib/decube/etcd",
			WalDir:                 "/var/lib/decube/etcd/wal",
			SnapshotCount:          10000,
			HeartbeatInterval:      100,
			ElectionTimeout:        1000,
			MaxSnapshots:           5,
			MaxWals:                5,
			AutoCompactionRetention: "1h",
			QuotaBackendBytes:      4294967296, // 4GB
		},
		API: APIConfig{
			REST: RESTConfig{
				Enabled: true,
				Address: "0.0.0.0:8080",
				CORS:    []string{"*"},
			},
			GRPC: GRPCConfig{
				Enabled: true,
				Address: "0.0.0.0:9090",
			},
		},
		Replication: ReplicationConfig{
			Enabled:      true,
			PeerTimeout:  5 * time.Second,
			RetryInterval: 1 * time.Second,
			MaxRetries:   3,
		},
		Snapshot: SnapshotConfig{
			Enabled:       true,
			Interval:      1 * time.Hour,
			RetentionCount: 10,
			Compression:   true,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
		},
		Security: SecurityConfig{
			TLSEnabled: false,
			CertFile:   "",
			KeyFile:    "",
			CAFile:     "",
		},
	}
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	viper := viper.New()

	// Set defaults
	viper.SetDefault("node.id", cfg.Node.ID)
	viper.SetDefault("node.data_dir", cfg.Node.DataDir)
	viper.SetDefault("node.listen_address", cfg.Node.ListenAddress)
	viper.SetDefault("node.peer_addresses", cfg.Node.PeerAddresses)
	viper.SetDefault("etcd.name", cfg.Etcd.Name)
	viper.SetDefault("etcd.data_dir", cfg.Etcd.DataDir)
	viper.SetDefault("etcd.wal_dir", cfg.Etcd.WalDir)
	viper.SetDefault("etcd.snapshot_count", cfg.Etcd.SnapshotCount)
	viper.SetDefault("etcd.heartbeat_interval", cfg.Etcd.HeartbeatInterval)
	viper.SetDefault("etcd.election_timeout", cfg.Etcd.ElectionTimeout)
	viper.SetDefault("etcd.max_snapshots", cfg.Etcd.MaxSnapshots)
	viper.SetDefault("etcd.max_wals", cfg.Etcd.MaxWals)
	viper.SetDefault("etcd.auto_compaction_retention", cfg.Etcd.AutoCompactionRetention)
	viper.SetDefault("etcd.quota_backend_bytes", cfg.Etcd.QuotaBackendBytes)
	viper.SetDefault("api.rest.enabled", cfg.API.REST.Enabled)
	viper.SetDefault("api.rest.address", cfg.API.REST.Address)
	viper.SetDefault("api.rest.cors_origins", cfg.API.REST.CORS)
	viper.SetDefault("api.grpc.enabled", cfg.API.GRPC.Enabled)
	viper.SetDefault("api.grpc.address", cfg.API.GRPC.Address)
	viper.SetDefault("replication.enabled", cfg.Replication.Enabled)
	viper.SetDefault("replication.peer_timeout", cfg.Replication.PeerTimeout)
	viper.SetDefault("replication.retry_interval", cfg.Replication.RetryInterval)
	viper.SetDefault("replication.max_retries", cfg.Replication.MaxRetries)
	viper.SetDefault("snapshot.enabled", cfg.Snapshot.Enabled)
	viper.SetDefault("snapshot.interval", cfg.Snapshot.Interval)
	viper.SetDefault("snapshot.retention_count", cfg.Snapshot.RetentionCount)
	viper.SetDefault("snapshot.compression", cfg.Snapshot.Compression)
	viper.SetDefault("logging.level", cfg.Logging.Level)
	viper.SetDefault("logging.format", cfg.Logging.Format)
	viper.SetDefault("logging.output", cfg.Logging.Output)
	viper.SetDefault("logging.max_size", cfg.Logging.MaxSize)
	viper.SetDefault("logging.max_backups", cfg.Logging.MaxBackups)
	viper.SetDefault("logging.max_age", cfg.Logging.MaxAge)
	viper.SetDefault("security.tls_enabled", cfg.Security.TLSEnabled)
	viper.SetDefault("security.cert_file", cfg.Security.CertFile)
	viper.SetDefault("security.key_file", cfg.Security.KeyFile)
	viper.SetDefault("security.ca_file", cfg.Security.CAFile)

	// Environment variable bindings
	viper.SetEnvPrefix("DECUBE")
	viper.AutomaticEnv()

	// Set config file
	if configPath != "" {
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			return nil, err
		}
	}

	// Unmarshal into config struct
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
