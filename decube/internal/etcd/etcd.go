package etcd

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/embed"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/decube/decube/pkg/config"
)

// EtcdManager manages the embedded etcd instance
type EtcdManager struct {
	config     *config.Config
	etcd       *embed.Etcd
	client     *clientv3.Client
	isLeader   bool
	leaderAddr string
}

// NewEtcdManager creates a new etcd manager
func NewEtcdManager(cfg *config.Config) *EtcdManager {
	return &EtcdManager{
		config: cfg,
	}
}

// Start starts the embedded etcd server
func (e *EtcdManager) Start() error {
	embedCfg := embed.NewConfig()

	// Basic configuration
	embedCfg.Name = e.config.Etcd.Name
	embedCfg.Dir = e.config.Etcd.DataDir
	embedCfg.WalDir = e.config.Etcd.WalDir

	// Network configuration
	embedCfg.LCUrls = []url.URL{{Scheme: "http", Host: e.config.Node.ListenAddress}}
	embedCfg.ACUrls = []url.URL{{Scheme: "http", Host: e.config.Node.ListenAddress}}

	// Peer configuration
	peerURLs := make([]url.URL, len(e.config.Node.PeerAddresses))
	for i, addr := range e.config.Node.PeerAddresses {
		peerURLs[i] = url.URL{Scheme: "http", Host: addr}
	}
	embedCfg.LPUrls = peerURLs
	embedCfg.APUrls = peerURLs

	// Cluster configuration
	embedCfg.InitialCluster = e.buildInitialCluster()
	embedCfg.ClusterState = embed.ClusterStateFlagNew

	// Performance tuning
	embedCfg.SnapshotCount = uint64(e.config.Etcd.SnapshotCount)
	embedCfg.HeartbeatInterval = time.Duration(e.config.Etcd.HeartbeatInterval) * time.Millisecond
	embedCfg.ElectionTimeout = time.Duration(e.config.Etcd.ElectionTimeout) * time.Millisecond
	embedCfg.MaxSnapFiles = uint(e.config.Etcd.MaxSnapshots)
	embedCfg.MaxWalFiles = uint(e.config.Etcd.MaxWals)
	embedCfg.AutoCompactionRetention = e.config.Etcd.AutoCompactionRetention
	embedCfg.QuotaBackendBytes = int64(e.config.Etcd.QuotaBackendBytes)

	// Security configuration
	if e.config.Security.TLSEnabled {
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{},
			ClientCAs:    nil,
		}
		if e.config.Security.CertFile != "" && e.config.Security.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(e.config.Security.CertFile, e.config.Security.KeyFile)
			if err != nil {
				return fmt.Errorf("failed to load TLS cert: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		embedCfg.ClientTLSInfo = embed.TLSInfo{
			CertFile:      e.config.Security.CertFile,
			KeyFile:       e.config.Security.KeyFile,
			TrustedCAFile: e.config.Security.CAFile,
		}
		embedCfg.PeerTLSInfo = embedCfg.ClientTLSInfo
	}

	// Start etcd
	etcd, err := embed.StartEtcd(embedCfg)
	if err != nil {
		return fmt.Errorf("failed to start etcd: %w", err)
	}

	e.etcd = etcd

	// Wait for etcd to be ready
	select {
	case <-etcd.Server.ReadyNotify():
		log.Printf("etcd server is ready")
	case <-time.After(60 * time.Second):
		etcd.Server.Stop()
		return fmt.Errorf("etcd took too long to start")
	}

	// Create client
	clientCfg := clientv3.Config{
		Endpoints:   []string{"http://" + e.config.Node.ListenAddress},
		DialTimeout: 5 * time.Second,
	}

	if e.config.Security.TLSEnabled {
		tlsConfig := &tls.Config{}
		if e.config.Security.CertFile != "" && e.config.Security.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(e.config.Security.CertFile, e.config.Security.KeyFile)
			if err != nil {
				return fmt.Errorf("failed to load client TLS cert: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		if e.config.Security.CAFile != "" {
			// Load CA cert
		}
		clientCfg.TLS = tlsConfig
	}

	client, err := clientv3.New(clientCfg)
	if err != nil {
		return fmt.Errorf("failed to create etcd client: %w", err)
	}

	e.client = client

	// Start leadership monitoring
	go e.monitorLeadership()

	log.Printf("etcd manager started successfully")
	return nil
}

// Stop stops the etcd server
func (e *EtcdManager) Stop() error {
	if e.client != nil {
		e.client.Close()
	}
	if e.etcd != nil {
		e.etcd.Close()
	}
	return nil
}

// GetClient returns the etcd client
func (e *EtcdManager) GetClient() *clientv3.Client {
	return e.client
}

// IsLeader returns whether this node is the leader
func (e *EtcdManager) IsLeader() bool {
	return e.isLeader
}

// GetLeaderAddr returns the leader address
func (e *EtcdManager) GetLeaderAddr() string {
	return e.leaderAddr
}

// Put stores a key-value pair with strong consistency
func (e *EtcdManager) Put(ctx context.Context, key, value string) error {
	_, err := e.client.Put(ctx, key, value)
	return err
}

// Get retrieves a value by key
func (e *EtcdManager) Get(ctx context.Context, key string) (string, error) {
	resp, err := e.client.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if len(resp.Kvs) == 0 {
		return "", rpctypes.ErrKeyNotFound
	}

	return string(resp.Kvs[0].Value), nil
}

// Delete removes a key
func (e *EtcdManager) Delete(ctx context.Context, key string) error {
	_, err := e.client.Delete(ctx, key)
	return err
}

// GetWithPrefix retrieves all keys with a given prefix
func (e *EtcdManager) GetWithPrefix(ctx context.Context, prefix string) (map[string]string, error) {
	resp, err := e.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, kv := range resp.Kvs {
		result[string(kv.Key)] = string(kv.Value)
	}

	return result, nil
}

// Watch watches for changes to keys with a given prefix
func (e *EtcdManager) Watch(ctx context.Context, prefix string) clientv3.WatchChan {
	return e.client.Watch(ctx, prefix, clientv3.WithPrefix())
}

// CreateSnapshot creates a snapshot of the current etcd state
func (e *EtcdManager) CreateSnapshot(ctx context.Context) ([]byte, error) {
	return e.etcd.Server.Snapshot(ctx)
}

// RestoreFromSnapshot restores etcd from a snapshot
func (e *EtcdManager) RestoreFromSnapshot(snapshotData []byte, skipHashCheck bool) error {
	// This is a simplified implementation
	// In production, you'd need to stop etcd, restore from snapshot, and restart
	return fmt.Errorf("snapshot restore not implemented")
}

// buildInitialCluster builds the initial cluster configuration string
func (e *EtcdManager) buildInitialCluster() string {
	var peers []string
	for i, addr := range e.config.Node.PeerAddresses {
		name := fmt.Sprintf("node-%d", i+1)
		peers = append(peers, fmt.Sprintf("%s=http://%s", name, addr))
	}
	return strings.Join(peers, ",")
}

// monitorLeadership monitors leadership changes
func (e *EtcdManager) monitorLeadership() {
	watchCh := e.etcd.Server.WatchLeadership()

	for {
		select {
		case <-e.etcd.Server.StopNotify():
			return
		case <-watchCh:
			leaderID := e.etcd.Server.Leader()
			e.isLeader = leaderID == e.etcd.Server.ID()

			if e.isLeader {
				e.leaderAddr = e.config.Node.ListenAddress
			} else {
				// Find leader address from cluster members
				members := e.etcd.Server.Cluster().Members()
				for _, member := range members {
					if member.ID == leaderID {
						if len(member.PeerURLs) > 0 {
							u, err := url.Parse(member.PeerURLs[0])
							if err == nil {
								host, _, err := net.SplitHostPort(u.Host)
								if err == nil {
									e.leaderAddr = host + ":2379" // Assume client port
								}
							}
						}
						break
					}
				}
			}

			log.Printf("Leadership changed. Is leader: %v, Leader addr: %s", e.isLeader, e.leaderAddr)
		}
	}
}
