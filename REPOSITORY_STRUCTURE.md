# DeCube Repository Structure

This document provides an overview of the DeCube repository structure.

## Directory Structure

```
DeCub/
├── .github/                    # GitHub configuration
│   ├── workflows/              # CI/CD workflows
│   ├── ISSUE_TEMPLATE/        # Issue templates
│   ├── dependabot.yml         # Dependabot configuration
│   ├── FUNDING.yml            # Funding information
│   └── pull_request_template.md
├── benchmarks/                # Performance benchmarks
├── cmd/                       # Command-line tools
│   ├── decubectl/            # DeCube CLI
│   └── rechainctl/           # REChain CLI
├── config/                    # Configuration files
│   ├── config.example.yaml   # Example configuration
│   └── docker-compose.yml     # Docker Compose setup
├── decub-cas/                 # Content Addressable Storage
├── decub-catalog/             # CRDT Catalog Service
├── decub-control-plane/       # Control Plane
├── decub-crypto/              # Cryptographic utilities
├── decub-gcl/                 # Global Consensus Layer
│   ├── go/                   # Go implementation
│   └── rust/                 # Rust implementation
├── decub-gossip/              # Gossip Protocol
├── decub-object-storage/      # Object Storage Service
├── decub-snapshot/            # Snapshot Service
├── decube/                    # Main DeCube service
├── docs/                      # Documentation
│   ├── api.md                # API documentation
│   ├── architecture.md       # Architecture guide
│   ├── deployment.md         # Deployment guide
│   ├── development.md        # Development guide
│   ├── faq.md               # Frequently asked questions
│   ├── getting-started.md   # Getting started guide
│   ├── glossary.md          # Glossary of terms
│   ├── monitoring.md         # Monitoring guide
│   ├── performance.md        # Performance guide
│   ├── roadmap.md           # Project roadmap
│   ├── setup.md             # Setup guide
│   └── troubleshooting.md   # Troubleshooting guide
├── examples/                  # Example code
│   ├── quickstart/           # Quick start example
│   └── snapshot-example/     # Snapshot example
├── rechain/                  # REChain core
│   ├── api/                 # API definitions
│   ├── cmd/                 # Commands
│   ├── config/              # Configuration
│   ├── internal/            # Internal packages
│   └── pkg/                 # Public packages
├── scripts/                  # Utility scripts
│   ├── build-release.sh     # Build release binaries
│   ├── clean.sh             # Clean build artifacts
│   ├── generate-docs.sh     # Generate documentation
│   ├── health-check.sh      # Health check script
│   ├── setup-dev.sh         # Development setup
│   └── validate-config.sh   # Validate configuration
├── src/                      # Source code (legacy)
├── tests/                    # Integration tests
├── .dockerignore            # Docker ignore patterns
├── .editorconfig            # Editor configuration
├── .gitattributes           # Git attributes
├── .gitignore               # Git ignore patterns
├── .golangci.yml            # Go linter configuration
├── .cursorrules             # Cursor IDE rules
├── CHANGELOG.md             # Changelog
├── CODE_OF_CONDUCT.md       # Code of conduct
├── CONTRIBUTING.md          # Contributing guidelines
├── docker-compose.yml       # Main Docker Compose file
├── go.mod                   # Go module definition
├── LICENSE                  # License file
├── Makefile                 # Build automation
├── PROJECT_STATUS.md        # Project status
├── README.md                # Main README
├── REPOSITORY_STRUCTURE.md  # This file
├── SECURITY.md              # Security policy
└── TODO.md                  # TODO list

```

## Key Directories

### `.github/`
GitHub-specific configuration including CI/CD workflows, issue templates, and pull request templates.

### `cmd/`
Command-line tools and entry points for various components.

### `config/`
Configuration files and examples for different deployment scenarios.

### `decub-*/`
Individual DeCube components, each with its own module and functionality.

### `docs/`
Comprehensive documentation covering architecture, APIs, deployment, and more.

### `examples/`
Example code demonstrating how to use DeCube components.

### `rechain/`
Core REChain implementation with internal and public packages.

### `scripts/`
Utility scripts for development, building, and maintenance.

### `tests/`
Integration and end-to-end tests.

## File Types

### Configuration Files
- `*.yaml`, `*.yml`: YAML configuration files
- `*.json`: JSON configuration files
- `go.mod`: Go module definitions

### Documentation
- `*.md`: Markdown documentation files
- `docs/`: Comprehensive documentation directory

### Scripts
- `*.sh`: Shell scripts (Unix/Linux/macOS)
- `Makefile`: Build automation

### Code
- `*.go`: Go source code
- `*.rs`: Rust source code
- `*.proto`: Protocol buffer definitions

## Naming Conventions

- **Directories**: lowercase with hyphens (kebab-case)
- **Go files**: lowercase with underscores (snake_case)
- **Documentation**: UPPERCASE for important files (README.md, LICENSE)
- **Config files**: lowercase with dots (config.yaml)

## Component Organization

Each major component (`decub-*`) follows a similar structure:
- `main.go`: Entry point
- `go.mod`: Module definition
- `README.md`: Component-specific documentation
- Additional source files as needed

## Documentation Organization

- **Getting Started**: `docs/getting-started.md`
- **Architecture**: `docs/architecture.md`
- **API Reference**: `docs/api.md`
- **Deployment**: `docs/deployment.md`
- **Development**: `docs/development.md`
- **Troubleshooting**: `docs/troubleshooting.md`

## Contributing

When adding new files or directories:
1. Follow existing naming conventions
2. Update this document if adding major directories
3. Add appropriate documentation
4. Update relevant README files

## Questions?

If you're unsure where to place a file or have questions about the structure, open an issue or ask in discussions.

