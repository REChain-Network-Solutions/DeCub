package main

import (
	"bytes"
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
)

// CLI represents the command-line interface
type CLI struct {
	serverURL string
	key       []byte
}

// NewCLI creates a new CLI instance
func NewCLI(serverURL string, key []byte) *CLI {
	return &CLI{
		serverURL: serverURL,
		key:       key,
	}
}

// upload uploads a file to the object storage
func (c *CLI) upload(filePath string, encrypt bool) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	// Compute SHA256
	hash := sha256.Sum256(data)
	sha256Str := hex.EncodeToString(hash[:])

	fmt.Printf("Uploading %s (SHA256: %s)\n", filePath, sha256Str)

	// Create request
	url := c.serverURL + "/chunk"
	if encrypt {
		url += "?encrypt=true"
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s", string(body))
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	fmt.Printf("Upload successful. SHA256: %s\n", result["sha256"])
	return nil
}

// download downloads a file from the object storage
func (c *CLI) download(sha256Str, outputPath string) error {
	fmt.Printf("Downloading %s to %s\n", sha256Str, outputPath)

	url := fmt.Sprintf("%s/chunk/%s", c.serverURL, sha256Str)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed: %s", string(body))
	}

	// Create output directory if needed
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Download successful. Saved to %s\n", outputPath)
	return nil
}

// verify verifies the integrity of a stored chunk
func (c *CLI) verify(sha256Str string) error {
	fmt.Printf("Verifying %s\n", sha256Str)

	url := fmt.Sprintf("%s/chunk/%s/verify", c.serverURL, sha256Str)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("verify failed: %s", string(body))
	}

	var result map[string]bool
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result["valid"] {
		fmt.Println("Verification successful: chunk is valid")
	} else {
		fmt.Println("Verification failed: chunk is corrupted")
	}

	return nil
}

// RunCLI runs the command-line interface
func RunCLI() {
	if len(os.Args) < 4 {
		fmt.Println("Usage:")
		fmt.Println("  upload <server-url> <file-path> [encrypt] [key]")
		fmt.Println("  download <server-url> <sha256> <output-path> [key]")
		fmt.Println("  verify <server-url> <sha256>")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  go run main.go cli upload http://localhost:8080 /path/to/file.txt true 0123456789abcdef...")
		fmt.Println("  go run main.go cli download http://localhost:8080 a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3 /tmp/downloaded.txt")
		fmt.Println("  go run main.go cli verify http://localhost:8080 a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3")
		os.Exit(1)
	}

	command := os.Args[2]
	serverURL := os.Args[3]

	var key []byte
	if len(os.Args) > 5 {
		keyStr := os.Args[5]
		if len(keyStr) != 64 {
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
	}

	cli := NewCLI(serverURL, key)

	switch command {
	case "upload":
		if len(os.Args) < 5 {
			log.Fatal("Usage: upload <server-url> <file-path> [encrypt]")
		}
		filePath := os.Args[4]
		encrypt := len(os.Args) > 5 && strings.ToLower(os.Args[5]) == "true"
		if err := cli.upload(filePath, encrypt); err != nil {
			log.Fatal(err)
		}

	case "download":
		if len(os.Args) < 6 {
			log.Fatal("Usage: download <server-url> <sha256> <output-path>")
		}
		sha256Str := os.Args[4]
		outputPath := os.Args[5]
		if err := cli.download(sha256Str, outputPath); err != nil {
			log.Fatal(err)
		}

	case "verify":
		if len(os.Args) < 5 {
			log.Fatal("Usage: verify <server-url> <sha256>")
		}
		sha256Str := os.Args[4]
		if err := cli.verify(sha256Str); err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatalf("Unknown command: %s", command)
	}
}
