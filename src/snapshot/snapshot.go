package snapshot

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/etcd-io/etcd/client/v3"
	"github.com/etcd-io/etcd/client/v3/snapshot"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// SnapshotMetadata holds info about a snapshot
type SnapshotMetadata struct {
	ID         string
	MerkleRoot string
	Hashes     []string
	Cluster    string
	Timestamp  time.Time
}

// CreateSnapshot creates an etcd snapshot, chunks it, computes hashes, uploads to MinIO
func CreateSnapshot(etcdEndpoint, minioEndpoint, accessKey, secretKey, bucket, cluster string) (*SnapshotMetadata, error) {
	// Create etcd snapshot
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdEndpoint},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to etcd: %v", err)
	}
	defer cli.Close()

	snapshotFile := fmt.Sprintf("etcd-snapshot-%s-%d.db", cluster, time.Now().Unix())
	err = snapshot.Save(context.Background(), cli, snapshotFile)
	if err != nil {
		return nil, fmt.Errorf("failed to save snapshot: %v", err)
	}
	defer os.Remove(snapshotFile) // Clean up local file after upload

	// Chunk and upload
	file, err := os.Open(snapshotFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %v", err)
	}

	exists, err := minioClient.BucketExists(context.Background(), bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %v", err)
	}
	if !exists {
		err = minioClient.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %v", err)
		}
	}

	chunks, hashes, err := chunkAndUpload(file, minioClient, bucket, snapshotFile)
	if err != nil {
		return nil, err
	}

	merkleRoot := computeMerkleRoot(hashes)

	meta := &SnapshotMetadata{
		ID:         fmt.Sprintf("%s-%d", cluster, time.Now().Unix()),
		MerkleRoot: merkleRoot,
		Hashes:     hashes,
		Cluster:    cluster,
		Timestamp:  time.Now(),
	}

	return meta, nil
}

// chunkAndUpload splits file into chunks, uploads to MinIO, returns hashes
func chunkAndUpload(file *os.File, minioClient *minio.Client, bucket, snapshotFile string) ([][]byte, []string, error) {
	const chunkSize = 64 * 1024 * 1024 // 64MB
	var chunks [][]byte
	var hashes []string
	buf := make([]byte, chunkSize)
	chunkIndex := 0

	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		chunk := make([]byte, n)
		copy(chunk, buf[:n])
		chunks = append(chunks, chunk)

		hash := sha256.Sum256(chunk)
		hashStr := fmt.Sprintf("%x", hash)
		hashes = append(hashes, hashStr)

		// Upload chunk
		objectName := fmt.Sprintf("%s-chunk-%d", filepath.Base(snapshotFile), chunkIndex)
		_, err = minioClient.PutObject(context.Background(), bucket, objectName, io.NopCloser(io.Reader(chunk)), int64(n), minio.PutObjectOptions{})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to upload chunk %d: %v", chunkIndex, err)
		}
		log.Printf("Uploaded chunk %d: %s", chunkIndex, hashStr)
		chunkIndex++
	}
	return chunks, hashes, nil
}

// computeMerkleRoot builds a Merkle tree from hashes
func computeMerkleRoot(hashes []string) string {
	if len(hashes) == 0 {
		return ""
	}
	if len(hashes) == 1 {
		return hashes[0]
	}

	var level []string
	for _, h := range hashes {
		level = append(level, h)
	}

	for len(level) > 1 {
		var nextLevel []string
		for i := 0; i < len(level); i += 2 {
			if i+1 < len(level) {
				combined := level[i] + level[i+1]
				hash := sha256.Sum256([]byte(combined))
				nextLevel = append(nextLevel, fmt.Sprintf("%x", hash))
			} else {
				nextLevel = append(nextLevel, level[i])
			}
		}
		level = nextLevel
	}
	return level[0]
}

// RestoreSnapshot downloads chunks from MinIO, verifies Merkle root, restores to etcd
func RestoreSnapshot(meta *SnapshotMetadata, etcdEndpoint, minioEndpoint, accessKey, secretKey, bucket string) error {
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create MinIO client: %v", err)
	}

	// Download chunks
	var chunks [][]byte
	for i, hash := range meta.Hashes {
		objectName := fmt.Sprintf("%s-chunk-%d", meta.ID, i)
		obj, err := minioClient.GetObject(context.Background(), bucket, objectName, minio.GetObjectOptions{})
		if err != nil {
			return fmt.Errorf("failed to get object %s: %v", objectName, err)
		}
		chunk, err := io.ReadAll(obj)
		if err != nil {
			return fmt.Errorf("failed to read chunk %d: %v", i, err)
		}
		obj.Close()

		// Verify hash
		computedHash := sha256.Sum256(chunk)
		if fmt.Sprintf("%x", computedHash) != hash {
			return fmt.Errorf("hash mismatch for chunk %d", i)
		}
		chunks = append(chunks, chunk)
	}

	// Verify Merkle root
	var chunkHashes []string
	for _, chunk := range chunks {
		hash := sha256.Sum256(chunk)
		chunkHashes = append(chunkHashes, fmt.Sprintf("%x", hash))
	}
	if computeMerkleRoot(chunkHashes) != meta.MerkleRoot {
		return fmt.Errorf("Merkle root verification failed")
	}

	// Reassemble snapshot file
	snapshotFile := fmt.Sprintf("restored-%s.db", meta.ID)
	file, err := os.Create(snapshotFile)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, chunk := range chunks {
		_, err = file.Write(chunk)
		if err != nil {
			return err
		}
	}

	// Restore to etcd
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdEndpoint},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to etcd: %v", err)
	}
	defer cli.Close()

	err = snapshot.Restore(context.Background(), cli, snapshotFile)
	if err != nil {
		return fmt.Errorf("failed to restore snapshot: %v", err)
	}

	os.Remove(snapshotFile) // Clean up
	log.Printf("Snapshot %s restored successfully", meta.ID)
	return nil
}
