package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/rechain/rechain/internal/api"
	"github.com/rechain/rechain/internal/cas"
	"github.com/rechain/rechain/internal/consensus"
	"github.com/rechain/rechain/internal/gossip"
	"github.com/rechain/rechain/internal/security"
	"github.com/rechain/rechain/internal/storage"
	"github.com/rechain/rechain/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullSystemIntegration(t *testing.T) {
	// Load test configuration
	cfg := config.DefaultConfig()
	cfg.Node.DataDir = t.TempDir()
	cfg.Storage.Path = cfg.Node.DataDir + "/storage"
	cfg.API.REST.Address = "localhost:0" // Use random port
	cfg.API.GRPC.Address = "localhost:0" // Use random port

	// Initialize components
	store, err := storage.NewBadgerStore(cfg.Storage.Path)
	require.NoError(t, err)
	defer store.Close()

	consensusEngine, err := consensus.NewConsensus(store, cfg.Consensus)
	require.NoError(t, err)

	gossipProtocol, err := gossip.NewGossipProtocol(cfg.Gossip)
	require.NoError(t, err)

	casStore, err := cas.NewCASStore(cfg.CAS)
	require.NoError(t, err)

	securityManager, err := security.NewSecurityManager(cfg.Security)
	require.NoError(t, err)

	// Create API server
	apiServer := api.NewServer(consensusEngine, gossipProtocol, casStore, securityManager, cfg.API)

	// Start API server
	go func() {
		if err := apiServer.Start(); err != nil {
			t.Logf("API server error: %v", err)
		}
	}()
	defer apiServer.Stop()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Get the actual server address
	restAddr := apiServer.RESTAddr()
	require.NotEmpty(t, restAddr)

	baseURL := fmt.Sprintf("http://%s", restAddr)

	t.Run("Health Check", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var health map[string]bool
		err = json.NewDecoder(resp.Body).Decode(&health)
		require.NoError(t, err)
		assert.True(t, health["ok"])
	})

	t.Run("Store and Retrieve Object", func(t *testing.T) {
		// Test data
		testData := []byte("Hello, REChain Integration Test!")
		testMetadata := map[string]string{
			"test":     "true",
			"filename": "test.txt",
		}

		// Store object
		storeReq := map[string]interface{}{
			"data":     testData,
			"metadata": testMetadata,
		}
		storeJSON, _ := json.Marshal(storeReq)

		resp, err := http.Post(baseURL+"/cas/objects", "application/json", bytes.NewReader(storeJSON))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var storeResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&storeResp)
		require.NoError(t, err)

		cid, ok := storeResp["cid"].(string)
		require.True(t, ok)
		assert.NotEmpty(t, cid)

		// Retrieve object
		resp2, err := http.Get(baseURL + "/cas/objects/" + cid)
		require.NoError(t, err)
		defer resp2.Body.Close()

		assert.Equal(t, http.StatusOK, resp2.StatusCode)

		var getResp map[string]interface{}
		err = json.NewDecoder(resp2.Body).Decode(&getResp)
		require.NoError(t, err)

		retrievedData, ok := getResp["data"].(string)
		require.True(t, ok)
		assert.Equal(t, string(testData), retrievedData)

		retrievedMetadata, ok := getResp["metadata"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, testMetadata["test"], retrievedMetadata["test"])
		assert.Equal(t, testMetadata["filename"], retrievedMetadata["filename"])
	})

	t.Run("Submit and Query Transaction", func(t *testing.T) {
		// Submit transaction
		txReq := map[string]interface{}{
			"type":    "test",
			"payload": map[string]string{"message": "integration test"},
		}
		txJSON, _ := json.Marshal(txReq)

		resp, err := http.Post(baseURL+"/txs", "application/json", bytes.NewReader(txJSON))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var txResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&txResp)
		require.NoError(t, err)

		txID, ok := txResp["tx_id"].(string)
		require.True(t, ok)
		assert.NotEmpty(t, txID)

		// Query transaction
		resp2, err := http.Get(baseURL + "/txs/" + txID)
		require.NoError(t, err)
		defer resp2.Body.Close()

		assert.Equal(t, http.StatusOK, resp2.StatusCode)

		var queryResp map[string]interface{}
		err = json.NewDecoder(resp2.Body).Decode(&queryResp)
		require.NoError(t, err)

		found, ok := queryResp["found"].(bool)
		require.True(t, ok)
		assert.True(t, found)
	})

	t.Run("Gossip State", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/gossip/state")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var gossipResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&gossipResp)
		require.NoError(t, err)

		// Should have some basic state information
		assert.Contains(t, gossipResp, "peer_count")
		assert.Contains(t, gossipResp, "last_sync")
	})

	t.Run("Node Info", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/node/info")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var nodeResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&nodeResp)
		require.NoError(t, err)

		assert.Contains(t, nodeResp, "node_id")
		assert.Contains(t, nodeResp, "version")
		assert.Contains(t, nodeResp, "network")
		assert.Contains(t, nodeResp, "block_height")
		assert.Contains(t, nodeResp, "consensus")
	})

	t.Run("Consensus State", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/consensus/state")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var consensusResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&consensusResp)
		require.NoError(t, err)

		assert.Contains(t, consensusResp, "height")
		assert.Contains(t, consensusResp, "round")
		assert.Contains(t, consensusResp, "step")
		assert.Contains(t, consensusResp, "proposer")
		assert.Contains(t, consensusResp, "validators")
	})
}

func TestSecurityIntegration(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Node.DataDir = t.TempDir()
	cfg.Security.EncryptData = true
	cfg.Security.SignTxs = true

	securityManager, err := security.NewSecurityManager(cfg.Security)
	require.NoError(t, err)

	t.Run("Encrypt/Decrypt Data", func(t *testing.T) {
		originalData := []byte("Sensitive data to encrypt")

		// Encrypt
		encrypted, err := securityManager.EncryptData(context.Background(), originalData)
		require.NoError(t, err)
		assert.NotEqual(t, originalData, encrypted)

		// Decrypt
		decrypted, err := securityManager.DecryptData(context.Background(), encrypted)
		require.NoError(t, err)
		assert.Equal(t, originalData, decrypted)
	})

	t.Run("Sign/Verify Transaction", func(t *testing.T) {
		txData := []byte("Transaction data to sign")

		// Sign
		signature, err := securityManager.SignTransaction(context.Background(), txData)
		require.NoError(t, err)
		assert.NotEmpty(t, signature)

		// Verify
		valid, err := securityManager.VerifyTransaction(context.Background(), txData, signature)
		require.NoError(t, err)
		assert.True(t, valid)

		// Verify with wrong data should fail
		valid, err = securityManager.VerifyTransaction(context.Background(), []byte("wrong data"), signature)
		require.NoError(t, err)
		assert.False(t, valid)
	})
}

func TestCASIntegration(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Node.DataDir = t.TempDir()

	// Create a temporary MinIO setup for testing
	// In real tests, you'd use testcontainers or a mock

	casStore, err := cas.NewCASStore(cfg.CAS)
	if err != nil {
		t.Skip("CAS store not available, skipping test")
	}

	t.Run("Store and Retrieve Large Object", func(t *testing.T) {
		// Create a 10MB test file
		largeData := make([]byte, 10*1024*1024)
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}

		metadata := map[string]string{
			"size": "10MB",
			"type": "test",
		}

		// Store object
		cid, err := casStore.StoreObject(context.Background(), largeData, metadata)
		require.NoError(t, err)
		assert.NotEmpty(t, cid)

		// Retrieve object
		retrievedData, retrievedMetadata, err := casStore.GetObject(context.Background(), cid)
		require.NoError(t, err)
		assert.Equal(t, largeData, retrievedData)
		assert.Equal(t, metadata, retrievedMetadata)
	})
}
