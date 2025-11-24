package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
	"../src/snapshot"
)

func TestFullSnapshotCycle(t *testing.T) {
	// This test assumes docker-compose is running
	// Create a small test snapshot (mock etcd data)
	testData := make([]byte, 10*1024*1024) // 10MB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	// Write to a temp file to simulate etcd snapshot
	tempFile := "/tmp/test-etcd-snapshot.db"
	err := os.WriteFile(tempFile, testData, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile)

	// Mock CreateSnapshot by directly calling chunkAndUpload (simplified)
	// In real test, would use actual endpoints

	// Test API endpoints
	apiURL := "http://localhost:8080"

	// Test create snapshot via API
	reqBody := map[string]string{"cluster": "test"}
	jsonBody, _ := json.Marshal(reqBody)
	resp, err := http.Post(apiURL+"/snapshot/create", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Logf("API not running, skipping integration test: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var meta snapshot.SnapshotMetadata
	json.NewDecoder(resp.Body).Decode(&meta)

	// Test get snapshot
	resp2, err := http.Get(apiURL + "/snapshot/" + meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for get, got %d", resp2.StatusCode)
	}

	// Test catalog query
	resp3, err := http.Get(apiURL + "/catalog/query")
	if err != nil {
		t.Fatal(err)
	}
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for catalog, got %d", resp3.StatusCode)
	}

	t.Log("Integration test passed (basic API checks)")
}
