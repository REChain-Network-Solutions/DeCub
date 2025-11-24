package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)

// ObjectStorage represents the object storage service
type ObjectStorage struct {
	dataDir string
	db      *bolt.DB
	key     []byte // AES-256 key
}

// ChunkMetadata represents metadata for a stored chunk
type ChunkMetadata struct {
	SHA256    string `json:"sha256"`
	Size      int64  `json:"size"`
	Encrypted bool   `json:"encrypted"`
}

// NewObjectStorage creates a new object storage instance
func NewObjectStorage(dataDir string, key []byte) (*ObjectStorage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Join(dataDir, "chunks"), 0755); err != nil {
		return nil, err
	}

	db, err := bolt.Open(filepath.Join(dataDir, "metadata.db"), 0600, nil)
	if err != nil {
		return nil, err
	}

	// Create buckets
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("chunks"))
		return err
	})
	if err != nil {
		return nil, err
	}

	return &ObjectStorage{
		dataDir: dataDir,
		db:      db,
		key:     key,
	}, nil
}

// computeSHA256 computes SHA256 hash of data
func (os *ObjectStorage) computeSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// encrypt encrypts data using AES-256-GCM
func (os *ObjectStorage) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(os.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-256-GCM
func (os *ObjectStorage) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(os.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// storeChunk stores a chunk with optional encryption
func (os *ObjectStorage) storeChunk(data []byte, encrypt bool) (string, error) {
	var finalData []byte
	var encrypted bool

	if encrypt {
		encryptedData, err := os.encrypt(data)
		if err != nil {
			return "", err
		}
		finalData = encryptedData
		encrypted = true
	} else {
		finalData = data
		encrypted = false
	}

	// Compute SHA256 of original data for integrity
	sha256 := os.computeSHA256(data)

	// Store file
	filePath := filepath.Join(os.dataDir, "chunks", sha256)
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := file.Write(finalData); err != nil {
		return "", err
	}

	// Store metadata
	metadata := ChunkMetadata{
		SHA256:    sha256,
		Size:      int64(len(data)),
		Encrypted: encrypted,
	}

	err = os.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("chunks"))
		jsonData, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(sha256), jsonData)
	})

	if err != nil {
		return "", err
	}

	return sha256, nil
}

// retrieveChunk retrieves a chunk by SHA256
func (os *ObjectStorage) retrieveChunk(sha256 string) ([]byte, error) {
	// Get metadata
	var metadata ChunkMetadata
	err := os.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("chunks"))
		data := bucket.Get([]byte(sha256))
		if data == nil {
			return fmt.Errorf("chunk not found")
		}
		return json.Unmarshal(data, &metadata)
	})
	if err != nil {
		return nil, err
	}

	// Read file
	filePath := filepath.Join(os.dataDir, "chunks", sha256)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Decrypt if necessary
	if metadata.Encrypted {
		data, err = os.decrypt(data)
		if err != nil {
			return nil, err
		}
	}

	// Verify integrity
	computedSHA256 := os.computeSHA256(data)
	if computedSHA256 != sha256 {
		return nil, fmt.Errorf("integrity check failed")
	}

	return data, nil
}

// verifyChunk verifies a chunk's integrity
func (os *ObjectStorage) verifyChunk(sha256 string) (bool, error) {
	data, err := os.retrieveChunk(sha256)
	if err != nil {
		return false, err
	}

	computedSHA256 := os.computeSHA256(data)
	return computedSHA256 == sha256, nil
}

// Close closes the object storage
func (os *ObjectStorage) Close() error {
	return os.db.Close()
}

// API handlers
func (os *ObjectStorage) handlePutChunk(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	encrypt := r.URL.Query().Get("encrypt") == "true"

	sha256, err := os.storeChunk(data, encrypt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"sha256": sha256}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (os *ObjectStorage) handleGetChunk(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sha256 := vars["sha256"]

	data, err := os.retrieveChunk(sha256)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

func (os *ObjectStorage) handleVerifyChunk(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sha256 := vars["sha256"]

	valid, err := os.verifyChunk(sha256)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]bool{"valid": valid}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run main.go <data-dir> [encryption-key]  # Start server")
		fmt.Println("  go run main.go cli <command> ...            # CLI mode")
		os.Exit(1)
	}

	if len(os.Args) > 2 && os.Args[1] == "cli" {
		RunCLI()
		return
	}

	dataDir := os.Args[1]

	var key []byte
	if len(os.Args) > 2 {
		keyStr := os.Args[2]
		if len(keyStr) != 64 { // 32 bytes * 2 for hex
			log.Fatal("Encryption key must be 64 hex characters (32 bytes)")
		}
		var err error
		key, err = hex.DecodeString(keyStr)
		if err != nil {
			log.Fatal("Invalid encryption key format")
		}
	} else {
		// Generate a random key for demo
		key = make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			log.Fatal("Failed to generate encryption key")
		}
		fmt.Printf("Generated encryption key: %s\n", hex.EncodeToString(key))
	}

	os, err := NewObjectStorage(dataDir, key)
	if err != nil {
		log.Fatalf("Failed to create object storage: %v", err)
	}
	defer os.Close()

	r := mux.NewRouter()
	r.HandleFunc("/chunk", os.handlePutChunk).Methods("PUT")
	r.HandleFunc("/chunk/{sha256}", os.handleGetChunk).Methods("GET")
	r.HandleFunc("/chunk/{sha256}/verify", os.handleVerifyChunk).Methods("GET")

	fmt.Println("Object storage server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
