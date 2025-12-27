package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	catalogEndpoint = "http://localhost:8080/catalog"
	snapshotEndpoint = "http://localhost:8080/snapshots"
)

type Snapshot struct {
	ID       string                 `json:"id"`
	Metadata map[string]interface{} `json:"metadata"`
}

func main() {
	fmt.Println("DeCube Snapshot Example")
	fmt.Println("=======================")

	// Wait for services to be ready
	fmt.Println("\nWaiting for services to be ready...")
	waitForService(catalogEndpoint, 30*time.Second)

	// Create a snapshot
	fmt.Println("\n1. Creating snapshot...")
	snapshot := createSnapshot("example-snapshot-001", map[string]interface{}{
		"size":    1073741824,
		"created": time.Now().Format(time.RFC3339),
		"cluster": "cluster-a",
	})
	fmt.Printf("   Created snapshot: %s\n", snapshot.ID)

	// Query snapshot
	fmt.Println("\n2. Querying snapshot...")
	queried := querySnapshot(snapshot.ID)
	if queried != nil {
		fmt.Printf("   Found snapshot: %s\n", queried.ID)
		fmt.Printf("   Metadata: %+v\n", queried.Metadata)
	}

	// List all snapshots
	fmt.Println("\n3. Listing all snapshots...")
	snapshots := listSnapshots()
	fmt.Printf("   Found %d snapshots\n", len(snapshots))
	for _, s := range snapshots {
		fmt.Printf("   - %s\n", s.ID)
	}

	// Clean up
	fmt.Println("\n4. Cleaning up...")
	deleteSnapshot(snapshot.ID)
	fmt.Println("   Snapshot deleted")
}

func waitForService(endpoint string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(endpoint + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			fmt.Println("   ✓ Services are ready")
			return
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Println("   ⚠ Services may not be ready")
}

func createSnapshot(id string, metadata map[string]interface{}) *Snapshot {
	snapshot := &Snapshot{
		ID:       id,
		Metadata: metadata,
	}

	jsonData, _ := json.Marshal(snapshot)
	resp, err := http.Post(snapshotEndpoint, "application/json", 
		bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("   Error: %s\n", string(body))
		return nil
	}

	return snapshot
}

func querySnapshot(id string) *Snapshot {
	resp, err := http.Get(fmt.Sprintf("%s/%s", snapshotEndpoint, id))
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var snapshot Snapshot
	json.NewDecoder(resp.Body).Decode(&snapshot)
	return &snapshot
}

func listSnapshots() []Snapshot {
	resp, err := http.Get(catalogEndpoint + "/snapshots")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	var snapshots []Snapshot
	json.NewDecoder(resp.Body).Decode(&snapshots)
	return snapshots
}

func deleteSnapshot(id string) {
	req, _ := http.NewRequest("DELETE", 
		fmt.Sprintf("%s/%s", snapshotEndpoint, id), nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
}

