# DeCube Object Storage

S3-compatible object storage service for DeCube with client-side encryption and integrity verification.

## Features

- **S3-compatible endpoints**: PUT /chunk, GET /chunk/{sha256}
- **Content addressing**: Files stored under `chunks/<sha256>`
- **Integrity verification**: SHA256 computation and verification
- **Client-side encryption**: AES-256-GCM with provided key
- **Metadata index**: BoltDB for fast lookups
- **CLI tool**: Upload, download, and verify operations

## API Endpoints

- `PUT /chunk`: Store a chunk
  - Query param: `encrypt=true` for encryption
  - Returns: `{"sha256": "hash"}`

- `GET /chunk/{sha256}`: Retrieve a chunk

- `GET /chunk/{sha256}/verify`: Verify chunk integrity
  - Returns: `{"valid": true/false}`

## Running the Server

```bash
go mod tidy
go run main.go <data-dir> [encryption-key]
```

Example:
```bash
go run main.go ./data 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
```

If no key is provided, a random key is generated and printed.

## CLI Usage

### Upload a file
```bash
go run main.go cli upload http://localhost:8080 /path/to/file.txt true [key]
```

### Download a file
```bash
go run main.go cli download http://localhost:8080 <sha256> /path/to/output.txt [key]
```

### Verify integrity
```bash
go run main.go cli verify http://localhost:8080 <sha256>
```

## Example Workflow

1. **Start the server**:
   ```bash
   go run main.go ./data
   # Generated encryption key: abc123...
   ```

2. **Upload a file with encryption**:
   ```bash
   go run main.go cli upload http://localhost:8080 README.md true abc123...
   # Uploading README.md (SHA256: a665a459...)
   # Upload successful. SHA256: a665a459...
   ```

3. **Download the file**:
   ```bash
   go run main.go cli download http://localhost:8080 a665a459... /tmp/downloaded.md abc123...
   # Downloading a665a459... to /tmp/downloaded.md
   # Download successful. Saved to /tmp/downloaded.md
   ```

4. **Verify integrity**:
   ```bash
   go run main.go cli verify http://localhost:8080 a665a459...
   # Verifying a665a459...
   # Verification successful: chunk is valid
   ```

## Storage Structure

```
data-dir/
├── chunks/
│   ├── a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3
│   └── ...
└── metadata.db (BoltDB)
```

## Security

- **Encryption**: AES-256-GCM for confidentiality
- **Integrity**: SHA256 for tamper detection
- **Key Management**: Client-provided keys (HSM integration planned)
