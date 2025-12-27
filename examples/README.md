# DeCube Examples

This directory contains example code demonstrating how to use DeCube components.

## Examples

### Quick Start
- **Location**: `quickstart/`
- **Description**: Basic introduction to DeCube components
- **Run**: `cd quickstart && go run main.go`

### Snapshot Operations
- **Location**: `snapshot-example/`
- **Description**: Demonstrates snapshot creation, querying, and management
- **Run**: `cd snapshot-example && go run main.go`

## Prerequisites

Before running examples:

1. **Start DeCube services**:
   ```bash
   docker-compose up -d
   ```

2. **Wait for services to be ready**:
   ```bash
   ./scripts/health-check.sh
   ```

3. **Install dependencies**:
   ```bash
   go mod download
   ```

## Running Examples

Each example directory contains:
- `main.go`: Example code
- `README.md`: Example-specific documentation

To run an example:

```bash
cd examples/<example-name>
go run main.go
```

## Creating Your Own Examples

When creating new examples:

1. Create a new directory under `examples/`
2. Add a `main.go` file with your example code
3. Add a `README.md` explaining what the example demonstrates
4. Follow the existing code style and patterns
5. Update this README to include your example

## Example Structure

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    fmt.Println("Example: Description")
    
    // Example code here
    if err := doSomething(); err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Example completed successfully!")
}
```

## Contributing Examples

We welcome example contributions! See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

Examples should:
- Be clear and well-documented
- Demonstrate real-world use cases
- Include error handling
- Be easy to understand and modify

