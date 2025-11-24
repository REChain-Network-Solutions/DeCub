package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const chunkSize = 64 * 1024 * 1024 // 64MB

type SnapshotManager struct {
	etcdEndpoint string
	objectStore  string
	gclEndpoint  string
}

func NewSnapshotManager(etcd, objStore, gcl string) *SnapshotManager {
	return &SnapshotManager{
		etcdEndpoint: etcd,
		objectStore:  objStore,
		gclEndpoint:  gcl,
	}
}

func (sm *SnapshotManager) CreateSnapshot(snapshotID, etcdPath, volumePath string) error {
	log.Printf("Step 1: Creating snapshot %s", snapshotID)

	// Create etcd snapshot
	etcdSnapPath := fmt.Sprintf("/tmp/etcd-%s.snap", snapshotID)
	cmd := fmt.Sprintf("etcdctl snapshot save %s --endpoints=%s", etcdSnapPath, sm.etcdEndpoint)
	log.Printf("Running: %s", cmd)
	// Execute command (simulated)
	log.Printf("Etcd snapshot created at %s", etcdSnapPath)

	// Create volume snapshot (simulated)
	volumeSnapPath := fmt.Sprintf("/tmp/volume-%s.tar.gz", snapshotID)
	cmd = fmt.Sprintf("tar -czf %s %s", volumeSnapPath, volumePath)
	log.Printf("Running: %s", cmd)
	log.Printf("Volume snapshot created at %s", volumeSnapPath)

	// Combine snapshots
	combinedPath := fmt.Sprintf("/tmp/combined-%s.snap", snapshotID)
	cmd = fmt.Sprintf("cat %s %s > %s", etcdSnapPath, volumeSnapPath, combinedPath)
	log.Printf("Running: %s", cmd)
	log.Printf("Combined snapshot created at %s", combinedPath)

	return sm.processAndUpload(snapshotID, combinedPath)
}

func (sm *SnapshotManager) processAndUpload(snapshotID, combinedPath string) error {
	log.Printf("Step 2: Chunking data into 64MB files")

	chunks, err := sm.chunkFile(combinedPath, snapshotID)
	if err != nil {
		return err
	}

	log.Printf("Created %d chunks", len(chunks))

	log.Printf("Step 3: Uploading to object store with sha256 verification")

	hashes := make([]string, len(chunks))
	for i, chunkPath := range chunks {
		hash, err := sm.uploadChunk(chunkPath, snapshotID, i)
		if err != nil {
			return err
		}
		hashes[i] = hash
		log.Printf("Uploaded chunk %d with hash %s", i, hash)
	}

	log.Printf("Step 4: Registering snapshot metadata via GCL tx")

	metadata := map[string]interface{}{
		"id":          snapshotID,
		"timestamp":   time.Now().Unix(),
		"chunk_count": len(chunks),
		"hashes":      hashes,
		"total_size":  sm.getFileSize(combinedPath),
	}

	return sm.registerMetadata(metadata)
}

func (sm *SnapshotManager) chunkFile(filePath, snapshotID string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var chunks []string
	buffer := make([]byte, chunkSize)
	chunkIndex := 0

	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}

		chunkPath := fmt.Sprintf("/tmp/%s-chunk-%d", snapshotID, chunkIndex)
		chunkFile, err := os.Create(chunkPath)
		if err != nil {
			return nil, err
		}

		_, err = chunkFile.Write(buffer[:n])
		chunkFile.Close()
		if err != nil {
			return nil, err
		}

		chunks = append(chunks, chunkPath)
		chunkIndex++
	}

	return chunks, nil
}

func (sm *SnapshotManager) uploadChunk(chunkPath, snapshotID string, index int) (string, error) {
	file, err := os.Open(chunkPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", err
	}
	hash := hex.EncodeToString(hasher.Sum(nil))

	// Simulate upload to object store
	objectKey := fmt.Sprintf("snapshots/%s/chunk-%d", snapshotID, index)
	cmd := fmt.Sprintf("aws s3 cp %s s3://%s/%s --endpoint-url=%s", chunkPath, sm.objectStore, objectKey, sm.objectStore)
	log.Printf("Running: %s", cmd)
	log.Printf("Uploaded chunk to %s", objectKey)

	return hash, nil
}

func (sm *SnapshotManager) registerMetadata(metadata map[string]interface{}) error {
	// Simulate GCL transaction
	cmd := fmt.Sprintf("gcl-cli tx register-snapshot --metadata='%v' --endpoint=%s", metadata, sm.gclEndpoint)
	log.Printf("Running: %s", cmd)
	log.Printf("Snapshot metadata registered with ID: %s", metadata["id"])

	return nil
}

func (sm *SnapshotManager) getFileSize(filePath string) int64 {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	return info.Size()
}

func (sm *SnapshotManager) VerifyAndRestore(snapshotID, restorePath string) error {
	log.Printf("Step 5: Verifying proof and restoring snapshot %s", snapshotID)

	// Get metadata from GCL
	metadata, err := sm.getMetadata(snapshotID)
	if err != nil {
		return err
	}

	log.Printf("Retrieved metadata for snapshot %s", snapshotID)

	// Download and verify chunks
	var combinedData []byte
	hashes := metadata["hashes"].([]string)
	chunkCount := int(metadata["chunk_count"].(float64))

	for i := 0; i < chunkCount; i++ {
		chunkData, err := sm.downloadAndVerifyChunk(snapshotID, i, hashes[i])
		if err != nil {
			return err
		}
		combinedData = append(combinedData, chunkData...)
		log.Printf("Verified and downloaded chunk %d", i)
	}

	// Restore combined snapshot
	combinedPath := fmt.Sprintf("/tmp/restore-%s.snap", snapshotID)
	err = os.WriteFile(combinedPath, combinedData, 0644)
	if err != nil {
		return err
	}

	log.Printf("Combined snapshot restored to %s", combinedPath)

	// Extract etcd and volume data
	return sm.extractSnapshots(combinedPath, restorePath)
}

func (sm *SnapshotManager) getMetadata(snapshotID string) (map[string]interface{}, error) {
	// Simulate getting metadata from GCL
	cmd := fmt.Sprintf("gcl-cli query snapshot %s --endpoint=%s", snapshotID, sm.gclEndpoint)
	log.Printf("Running: %s", cmd)

	// Mock metadata
	return map[string]interface{}{
		"id":          snapshotID,
		"timestamp":   time.Now().Unix(),
		"chunk_count": 2,
		"hashes":      []string{"mockhash1", "mockhash2"},
		"total_size":  128 * 1024 * 1024,
	}, nil
}

func (sm *SnapshotManager) downloadAndVerifyChunk(snapshotID string, index int, expectedHash string) ([]byte, error) {
	objectKey := fmt.Sprintf("snapshots/%s/chunk-%d", snapshotID, index)
	localPath := fmt.Sprintf("/tmp/download-%s-%d", snapshotID, index)

	// Simulate download
	cmd := fmt.Sprintf("aws s3 cp s3://%s/%s %s --endpoint-url=%s", sm.objectStore, objectKey, localPath, sm.objectStore)
	log.Printf("Running: %s", cmd)

	// Read file and verify hash
	data, err := os.ReadFile(localPath)
	if err != nil {
		return nil, err
	}

	hasher := sha256.New()
	hasher.Write(data)
	actualHash := hex.EncodeToString(hasher.Sum(nil))

	if actualHash != expectedHash {
		return nil, fmt.Errorf("hash mismatch for chunk %d: expected %s, got %s", index, expectedHash, actualHash)
	}

	return data, nil
}

func (sm *SnapshotManager) extractSnapshots(combinedPath, restorePath string) error {
	// Simulate extraction
	etcdRestore := filepath.Join(restorePath, "etcd")
	volumeRestore := filepath.Join(restorePath, "volumes")

	os.MkdirAll(etcdRestore, 0755)
	os.MkdirAll(volumeRestore, 0755)

	// Extract etcd snapshot
	cmd := fmt.Sprintf("head -c 64M %s > %s/etcd.snap", combinedPath, etcdRestore)
	log.Printf("Running: %s", cmd)

	// Extract volume snapshot
	cmd = fmt.Sprintf("tail -c +64M %s | tar -xzf - -C %s", combinedPath, volumeRestore)
	log.Printf("Running: %s", cmd)

	log.Printf("Snapshot restored to %s", restorePath)
	return nil
}

func main() {
	var etcdEndpoint, objectStore, gclEndpoint string

	rootCmd := &cobra.Command{
		Use:   "decub-snapshot",
		Short: "Decub snapshot lifecycle manager",
	}

	createCmd := &cobra.Command{
		Use:   "create [snapshot-id] [etcd-path] [volume-path]",
		Short: "Create a new snapshot",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			sm := NewSnapshotManager(etcdEndpoint, objectStore, gclEndpoint)
			err := sm.CreateSnapshot(args[0], args[1], args[2])
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Snapshot %s created successfully", args[0])
		},
	}

	restoreCmd := &cobra.Command{
		Use:   "restore [snapshot-id] [restore-path]",
		Short: "Restore a snapshot",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			sm := NewSnapshotManager(etcdEndpoint, objectStore, gclEndpoint)
			err := sm.VerifyAndRestore(args[0], args[1])
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Snapshot %s restored successfully to %s", args[0], args[1])
		},
	}

	rootCmd.PersistentFlags().StringVar(&etcdEndpoint, "etcd", "http://localhost:2379", "Etcd endpoint")
	rootCmd.PersistentFlags().StringVar(&objectStore, "object-store", "http://localhost:9000", "Object store endpoint")
	rootCmd.PersistentFlags().StringVar(&gclEndpoint, "gcl", "http://localhost:8080", "GCL endpoint")

	rootCmd.AddCommand(createCmd, restoreCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
