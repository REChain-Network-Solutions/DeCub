package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/syndtr/goleveldb/leveldb"
)

// CAS represents the Content-Addressed Storage
type CAS struct {
	minioClient *minio.Client
	bucket      string
	db          *leveldb.DB
}

// NewCAS creates a new CAS instance
func NewCAS(endpoint, accessKey, secretKey, bucket string) (*CAS, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false, // For local MinIO
	})
	if err != nil {
		return nil, err
	}

	// Create bucket if it doesn't exist
	err = minioClient.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(context.Background(), bucket)
		if errBucketExists != nil || !exists {
			return nil, err
		}
	}

	// Open LevelDB
	db, err := leveldb.OpenFile("./cas.db", nil)
	if err != nil {
		return nil, err
	}

	return &CAS{
		minioClient: minioClient,
		bucket:      bucket,
		db:          db,
	}, nil
}

// Store stores data and returns its content address (hash)
func (c *CAS) Store(ctx context.Context, data []byte) (string, error) {
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	// Check if already exists
	_, err := c.minioClient.StatObject(ctx, c.bucket, hashStr, minio.StatObjectOptions{})
	if err == nil {
		// Already exists
		return hashStr, nil
	}

	// Store in MinIO
	reader := strings.NewReader(string(data))
	_, err = c.minioClient.PutObject(ctx, c.bucket, hashStr, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return "", err
	}

	// Store metadata in LevelDB
	err = c.db.Put([]byte(hashStr), data, nil)
	if err != nil {
		return "", err
	}

	return hashStr, nil
}

// Retrieve retrieves data by its content address
func (c *CAS) Retrieve(ctx context.Context, hash string) ([]byte, error) {
	// First check LevelDB
	data, err := c.db.Get([]byte(hash), nil)
	if err == nil {
		return data, nil
	}

	// Fallback to MinIO
	obj, err := c.minioClient.GetObject(ctx, c.bucket, hash, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	data, err = io.ReadAll(obj)
	if err != nil {
		return nil, err
	}

	// Cache in LevelDB
	c.db.Put([]byte(hash), data, nil)

	return data, nil
}

// ChunkAndStore chunks large data and stores chunks
func (c *CAS) ChunkAndStore(ctx context.Context, data []byte, chunkSize int) ([]string, error) {
	var hashes []string
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := data[i:end]
		hash, err := c.Store(ctx, chunk)
		if err != nil {
			return nil, err
		}
		hashes = append(hashes, hash)
	}
	return hashes, nil
}

// RetrieveChunks retrieves and reassembles chunks
func (c *CAS) RetrieveChunks(ctx context.Context, hashes []string) ([]byte, error) {
	var data []byte
	for _, hash := range hashes {
		chunk, err := c.Retrieve(ctx, hash)
		if err != nil {
			return nil, err
		}
		data = append(data, chunk...)
	}
	return data, nil
}

// MerkleNode represents a node in the Merkle tree
type MerkleNode struct {
	Hash  string
	Left  *MerkleNode
	Right *MerkleNode
}

// BuildMerkleTree builds a Merkle tree from hashes
func BuildMerkleTree(hashes []string) *MerkleNode {
	if len(hashes) == 0 {
		return nil
	}

	nodes := make([]*MerkleNode, len(hashes))
	for i, h := range hashes {
		nodes[i] = &MerkleNode{Hash: h}
	}

	for len(nodes) > 1 {
		var newNodes []*MerkleNode
		for i := 0; i < len(nodes); i += 2 {
			left := nodes[i]
			var right *MerkleNode
			if i+1 < len(nodes) {
				right = nodes[i+1]
			} else {
				right = left // Duplicate for odd number
			}
			hash := sha256.Sum256([]byte(left.Hash + right.Hash))
			newNodes = append(newNodes, &MerkleNode{Hash: hex.EncodeToString(hash[:]), Left: left, Right: right})
		}
		nodes = newNodes
	}

	return nodes[0]
}

// GenerateMerkleProof generates a Merkle proof for a chunk
func GenerateMerkleProof(root *MerkleNode, index int) []string {
	var proof []string
	current := root
	for current.Left != nil || current.Right != nil {
		if index%2 == 0 {
			if current.Right != nil {
				proof = append(proof, current.Right.Hash)
			}
		} else {
			if current.Left != nil {
				proof = append(proof, current.Left.Hash)
			}
		}
		if index%2 == 0 {
			current = current.Left
		} else {
			current = current.Right
		}
		index /= 2
	}
	return proof
}

// VerifyMerkleProof verifies a Merkle proof
func VerifyMerkleProof(rootHash string, chunkHash string, proof []string, index int) bool {
	hash := chunkHash
	for _, p := range proof {
		if index%2 == 0 {
			hash = hex.EncodeToString(sha256.Sum256([]byte(hash + p)))
		} else {
			hash = hex.EncodeToString(sha256.Sum256([]byte(p + hash)))
		}
		index /= 2
	}
	return hash == rootHash
}

// API handlers
func (c *CAS) handleStore(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hash, err := c.Store(r.Context(), data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", hash)
}

func (c *CAS) handleRetrieve(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	data, err := c.Retrieve(r.Context(), hash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

func (c *CAS) handleChunkStore(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashes, err := c.ChunkAndStore(r.Context(), data, 1024*1024) // 1MB chunks
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	root := BuildMerkleTree(hashes)
	if root == nil {
		http.Error(w, "Failed to build Merkle tree", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"hashes": %q, "merkle_root": "%s"}`, hashes, root.Hash)
}

func (c *CAS) handleChunkRetrieve(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hashStr := vars["hashes"]

	var hashes []string
	err := json.Unmarshal([]byte(hashStr), &hashes)
	if err != nil {
		http.Error(w, "Invalid hashes format", http.StatusBadRequest)
		return
	}

	data, err := c.RetrieveChunks(r.Context(), hashes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

func (c *CAS) Close() error {
	return c.db.Close()
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run main.go <minio-endpoint> <access-key> <secret-key> [bucket]")
		os.Exit(1)
	}

	endpoint := os.Args[1]
	accessKey := os.Args[2]
	secretKey := os.Args[3]
	bucket := "decub-cas"
	if len(os.Args) > 4 {
		bucket = os.Args[4]
	}

	cas, err := NewCAS(endpoint, accessKey, secretKey, bucket)
	if err != nil {
		log.Fatalf("Failed to create CAS: %v", err)
	}
	defer cas.Close()

	r := mux.NewRouter()
	r.HandleFunc("/store", cas.handleStore).Methods("POST")
	r.HandleFunc("/retrieve/{hash}", cas.handleRetrieve).Methods("GET")
	r.HandleFunc("/chunk/store", cas.handleChunkStore).Methods("POST")
	r.HandleFunc("/chunk/retrieve/{hashes}", cas.handleChunkRetrieve).Methods("GET")

	fmt.Println("CAS server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
