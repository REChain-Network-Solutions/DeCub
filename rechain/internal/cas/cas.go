package cas

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// CAS implements Content-Addressed Storage with S3 compatibility
type CAS struct {
	client     *minio.Client
	bucket     string
	chunkSize  int64
	maxRetries int
}

// ObjectInfo holds metadata about a stored object
type ObjectInfo struct {
	CID       string    // Content ID (hash)
	Size      int64     // Object size in bytes
	Chunks    []string  // Chunk CIDs
	MerkleRoot string   // Merkle root hash
	Uploaded  time.Time // Upload timestamp
	Metadata  map[string]string // Additional metadata
}

// NewCAS creates a new CAS instance
func NewCAS(endpoint, accessKey, secretKey, bucket string, secure bool) (*CAS, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	cas := &CAS{
		client:     client,
		bucket:     bucket,
		chunkSize:  64 * 1024 * 1024, // 64MB chunks
		maxRetries: 3,
	}

	// Ensure bucket exists
	if err := cas.ensureBucket(); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket: %w", err)
	}

	return cas, nil
}

// ensureBucket creates the bucket if it doesn't exist
func (cas *CAS) ensureBucket() error {
	exists, err := cas.client.BucketExists(context.Background(), cas.bucket)
	if err != nil {
		return err
	}

	if !exists {
		err = cas.client.MakeBucket(context.Background(), cas.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
		log.Printf("Created bucket: %s", cas.bucket)
	}

	return nil
}

// Store stores data in CAS and returns the content ID
func (cas *CAS) Store(ctx context.Context, reader io.Reader, metadata map[string]string) (*ObjectInfo, error) {
	// Read all data
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	// Calculate content ID
	cid := cas.calculateCID(data)

	// Check if already exists
	if exists, err := cas.Exists(ctx, cid); err != nil {
		return nil, err
	} else if exists {
		// Return existing object info
		return cas.GetInfo(ctx, cid)
	}

	// Chunk the data
	chunks, merkleRoot, err := cas.chunkData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk data: %w", err)
	}

	// Upload chunks
	chunkCIDs := make([]string, len(chunks))
	for i, chunk := range chunks {
		chunkCID := cas.calculateCID(chunk)
		chunkCIDs[i] = chunkCID

		if err := cas.uploadChunk(ctx, chunkCID, chunk); err != nil {
			return nil, fmt.Errorf("failed to upload chunk %d: %w", i, err)
		}
	}

	// Create object info
	objInfo := &ObjectInfo{
		CID:        cid,
		Size:       int64(len(data)),
		Chunks:     chunkCIDs,
		MerkleRoot: merkleRoot,
		Uploaded:   time.Now(),
		Metadata:   metadata,
	}

	// Store object metadata
	if err := cas.storeObjectInfo(ctx, objInfo); err != nil {
		return nil, fmt.Errorf("failed to store object info: %w", err)
	}

	log.Printf("Stored object %s (%d bytes, %d chunks)", cid, len(data), len(chunks))
	return objInfo, nil
}

// Retrieve retrieves data from CAS by content ID
func (cas *CAS) Retrieve(ctx context.Context, cid string) (io.ReadCloser, error) {
	// Get object info
	objInfo, err := cas.GetInfo(ctx, cid)
	if err != nil {
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	// Download all chunks
	chunks := make([][]byte, len(objInfo.Chunks))
	for i, chunkCID := range objInfo.Chunks {
		chunk, err := cas.downloadChunk(ctx, chunkCID)
		if err != nil {
			return nil, fmt.Errorf("failed to download chunk %d: %w", i, err)
		}
		chunks[i] = chunk
	}

	// Verify Merkle root
	if !cas.verifyMerkleRoot(chunks, objInfo.MerkleRoot) {
		return nil, fmt.Errorf("Merkle root verification failed")
	}

	// Concatenate chunks
	var data []byte
	for _, chunk := range chunks {
		data = append(data, chunk...)
	}

	// Return as reader
	return io.NopCloser(strings.NewReader(string(data))), nil
}

// Exists checks if an object exists in CAS
func (cas *CAS) Exists(ctx context.Context, cid string) (bool, error) {
	_, err := cas.client.StatObject(ctx, cas.bucket, cas.getObjectKey(cid), minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetInfo gets object information
func (cas *CAS) GetInfo(ctx context.Context, cid string) (*ObjectInfo, error) {
	obj, err := cas.client.GetObject(ctx, cas.bucket, cas.getMetadataKey(cid), minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	var objInfo ObjectInfo
	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, err
	}

	// Simple JSON deserialization (in production, use proper serialization)
	// This is simplified - in production, use protobuf or similar
	return &objInfo, fmt.Errorf("metadata parsing not implemented")
}

// Delete removes an object from CAS
func (cas *CAS) Delete(ctx context.Context, cid string) error {
	objInfo, err := cas.GetInfo(ctx, cid)
	if err != nil {
		return err
	}

	// Delete all chunks
	for _, chunkCID := range objInfo.Chunks {
		if err := cas.client.RemoveObject(ctx, cas.bucket, cas.getChunkKey(chunkCID), minio.RemoveObjectOptions{}); err != nil {
			log.Printf("Failed to delete chunk %s: %v", chunkCID, err)
		}
	}

	// Delete metadata
	if err := cas.client.RemoveObject(ctx, cas.bucket, cas.getMetadataKey(cid), minio.RemoveObjectOptions{}); err != nil {
		return err
	}

	log.Printf("Deleted object %s", cid)
	return nil
}

// List lists objects in CAS with optional prefix
func (cas *CAS) List(ctx context.Context, prefix string) ([]*ObjectInfo, error) {
	// This is a simplified implementation
	// In production, maintain an index of objects
	return nil, fmt.Errorf("list operation not fully implemented")
}

// calculateCID calculates the content ID for data
func (cas *CAS) calculateCID(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// chunkData splits data into chunks and computes Merkle root
func (cas *CAS) chunkData(data []byte) ([][]byte, string, error) {
	var chunks [][]byte
	size := int64(len(data))

	for offset := int64(0); offset < size; offset += cas.chunkSize {
		end := offset + cas.chunkSize
		if end > size {
			end = size
		}
		chunk := data[offset:end]
		chunks = append(chunks, chunk)
	}

	// Compute Merkle root
	merkleRoot := cas.computeMerkleRoot(chunks)

	return chunks, merkleRoot, nil
}

// computeMerkleRoot computes the Merkle root of chunks
func (cas *CAS) computeMerkleRoot(chunks [][]byte) string {
	if len(chunks) == 0 {
		return ""
	}

	// Convert chunks to hashes
	hashes := make([]string, len(chunks))
	for i, chunk := range chunks {
		hashes[i] = cas.calculateCID(chunk)
	}

	// Build Merkle tree
	for len(hashes) > 1 {
		var nextLevel []string
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				combined := hashes[i] + hashes[i+1]
				hash := sha256.Sum256([]byte(combined))
				nextLevel = append(nextLevel, hex.EncodeToString(hash[:]))
			} else {
				nextLevel = append(nextLevel, hashes[i])
			}
		}
		hashes = nextLevel
	}

	return hashes[0]
}

// verifyMerkleRoot verifies chunks against Merkle root
func (cas *CAS) verifyMerkleRoot(chunks [][]byte, expectedRoot string) bool {
	computedRoot := cas.computeMerkleRoot(chunks)
	return computedRoot == expectedRoot
}

// uploadChunk uploads a chunk to storage
func (cas *CAS) uploadChunk(ctx context.Context, cid string, data []byte) error {
	key := cas.getChunkKey(cid)
	reader := strings.NewReader(string(data))

	_, err := cas.client.PutObject(ctx, cas.bucket, key, reader, int64(len(data)), minio.PutObjectOptions{})
	return err
}

// downloadChunk downloads a chunk from storage
func (cas *CAS) downloadChunk(ctx context.Context, cid string) ([]byte, error) {
	key := cas.getChunkKey(cid)
	obj, err := cas.client.GetObject(ctx, cas.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	return io.ReadAll(obj)
}

// storeObjectInfo stores object metadata
func (cas *CAS) storeObjectInfo(ctx context.Context, info *ObjectInfo) error {
	// Simplified - in production, serialize properly
	data := []byte(fmt.Sprintf("CID: %s, Size: %d", info.CID, info.Size))
	key := cas.getMetadataKey(info.CID)

	_, err := cas.client.PutObject(ctx, cas.bucket, key, strings.NewReader(string(data)), int64(len(data)), minio.PutObjectOptions{})
	return err
}

// getObjectKey returns the S3 key for an object
func (cas *CAS) getObjectKey(cid string) string {
	return filepath.Join("objects", cid[:2], cid[2:4], cid)
}

// getChunkKey returns the S3 key for a chunk
func (cas *CAS) getChunkKey(cid string) string {
	return filepath.Join("chunks", cid[:2], cid[2:4], cid)
}

// getMetadataKey returns the S3 key for metadata
func (cas *CAS) getMetadataKey(cid string) string {
	return filepath.Join("metadata", cid[:2], cid[2:4], cid+".json")
}
