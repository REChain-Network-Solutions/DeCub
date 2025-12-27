# DeCube Scripts

Utility scripts for development, deployment, and maintenance.

## Available Scripts

### Development

#### `setup-dev.sh`
Sets up the development environment.

```bash
./scripts/setup-dev.sh
```

**What it does:**
- Checks prerequisites (Go, Docker, etc.)
- Installs development tools
- Downloads dependencies
- Creates necessary directories
- Builds components

#### `clean.sh`
Cleans build artifacts and temporary files.

```bash
./scripts/clean.sh
```

**What it does:**
- Removes build artifacts (`bin/`, `dist/`)
- Cleans Go cache
- Removes Docker containers and volumes
- Cleans test artifacts

#### `validate-config.sh`
Validates configuration files.

```bash
./scripts/validate-config.sh
```

**Requirements:**
- `yamllint` (install with `pip install yamllint`)

**What it does:**
- Validates YAML syntax
- Checks configuration structure

#### `generate-docs.sh`
Generates API documentation from code.

```bash
./scripts/generate-docs.sh
```

**Requirements:**
- `godoc` (install with `go install golang.org/x/tools/cmd/godoc@latest`)
- `protoc-gen-doc` (optional, for proto documentation)

**What it does:**
- Generates Go API documentation
- Generates protocol buffer documentation

### Operations

#### `health-check.sh`
Checks the health of DeCube services.

```bash
./scripts/health-check.sh
```

**Environment variables:**
- `CATALOG_ENDPOINT`: Catalog service endpoint (default: http://localhost:8080)
- `GOSSIP_ENDPOINT`: Gossip service endpoint (default: http://localhost:8000)
- `CAS_ENDPOINT`: CAS service endpoint (default: http://localhost:9000)

**What it does:**
- Checks Docker status
- Verifies service health endpoints
- Checks port usage

#### `build-release.sh`
Builds release binaries for multiple platforms.

```bash
./scripts/build-release.sh [version]
```

**What it does:**
- Builds binaries for Linux, macOS, and Windows
- Creates checksums
- Outputs to `dist/` directory

## Making Scripts Executable

On Unix-like systems:

```bash
chmod +x scripts/*.sh
```

## Adding New Scripts

When adding new scripts:

1. Place them in the `scripts/` directory
2. Make them executable
3. Add a shebang (`#!/bin/bash`)
4. Include error handling (`set -e`)
5. Document in this README
6. Follow existing script patterns

## Script Guidelines

- Use `set -e` for error handling
- Provide clear output messages
- Check prerequisites before running
- Use colors for better readability (optional)
- Include help/usage information
- Handle edge cases gracefully

## Troubleshooting

### Script won't run
- Ensure it's executable: `chmod +x scripts/script-name.sh`
- Check shebang is correct: `#!/bin/bash`
- Verify line endings are Unix-style (LF, not CRLF)

### Permission denied
- Run `chmod +x scripts/script-name.sh`
- Or run with `bash scripts/script-name.sh`

### Script fails silently
- Remove `set -e` temporarily to see errors
- Add `set -x` for debug output
- Check script output carefully

