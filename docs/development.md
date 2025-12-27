# Development Guide

This guide covers development practices, coding standards, and workflow for contributing to DeCube.

## Table of Contents

1. [Development Environment](#development-environment)
2. [Code Style](#code-style)
3. [Testing](#testing)
4. [Git Workflow](#git-workflow)
5. [Debugging](#debugging)
6. [Performance Profiling](#performance-profiling)

## Development Environment

### Setup

```bash
# Clone repository
git clone https://github.com/REChain-Network-Solutions/DeCub.git
cd DeCub

# Install development tools
./scripts/setup-dev.sh

# Install pre-commit hooks (optional)
pre-commit install
```

### Required Tools

- **Go 1.19+**: Programming language
- **golangci-lint**: Linting tool
- **Docker**: Container runtime
- **Make**: Build automation
- **Git**: Version control

### IDE Setup

#### VS Code

Recommended extensions:
- Go extension
- YAML extension
- Docker extension
- Markdown extension

#### GoLand/IntelliJ

- Install Go plugin
- Configure Go SDK
- Enable Go modules

## Code Style

### Go Code Style

Follow standard Go conventions:

```go
// Good: Clear, concise function names
func CreateSnapshot(id string, metadata map[string]interface{}) error {
    // Implementation
}

// Good: Proper error handling
if err != nil {
    return fmt.Errorf("failed to create snapshot: %w", err)
}

// Good: Context usage for cancellation
func ProcessRequest(ctx context.Context, req *Request) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Process request
    }
}
```

### Naming Conventions

- **Packages**: lowercase, single word
- **Functions**: PascalCase for exported, camelCase for private
- **Variables**: camelCase
- **Constants**: PascalCase or UPPER_SNAKE_CASE
- **Interfaces**: End with `-er` when possible (e.g., `Reader`, `Writer`)

### Documentation

- All exported functions must have doc comments
- Use complete sentences
- Start with the function name

```go
// CreateSnapshot creates a new snapshot with the given ID and metadata.
// It returns an error if the snapshot already exists or creation fails.
func CreateSnapshot(id string, metadata map[string]interface{}) error {
    // ...
}
```

### Error Handling

- Always check errors
- Wrap errors with context using `fmt.Errorf` with `%w`
- Return errors, don't log and ignore

```go
// Good
if err := doSomething(); err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// Bad
if err := doSomething(); err != nil {
    log.Printf("error: %v", err)
    // Continue anyway
}
```

## Testing

### Unit Tests

```bash
# Run all unit tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests in verbose mode
go test -v ./...

# Run specific test
go test -run TestCreateSnapshot ./...
```

### Integration Tests

```bash
# Run integration tests (requires services)
make test-integration

# Or manually
docker-compose up -d
go test -tags=integration ./tests/...
```

### Test Structure

```go
func TestCreateSnapshot(t *testing.T) {
    tests := []struct {
        name    string
        id      string
        wantErr bool
    }{
        {
            name:    "valid snapshot",
            id:      "test-001",
            wantErr: false,
        },
        {
            name:    "duplicate id",
            id:      "test-001",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := CreateSnapshot(tt.id, nil)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateSnapshot() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Benchmarks

```bash
# Run benchmarks
go test -bench=. -benchmem ./...

# Compare benchmarks
go test -bench=. -benchmem -benchcmp old.txt new.txt
```

## Git Workflow

### Branch Naming

- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation updates
- `refactor/description` - Code refactoring
- `test/description` - Test additions/updates

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add snapshot compression support

- Implement gzip compression
- Add compression configuration option
- Update API documentation

Fixes #123
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test changes
- `chore`: Maintenance tasks

### Pull Request Process

1. Create feature branch
2. Make changes and commit
3. Push to your fork
4. Create pull request
5. Address review comments
6. Squash commits if requested
7. Merge after approval

## Debugging

### Using Delve

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug a program
dlv debug ./cmd/decube

# Attach to running process
dlv attach <pid>
```

### Logging

Use structured logging:

```go
import "github.com/sirupsen/logrus"

log.WithFields(logrus.Fields{
    "snapshot_id": id,
    "size": size,
}).Info("Snapshot created")
```

### Debug Mode

Enable debug logging:

```yaml
logging:
  level: "debug"
```

## Performance Profiling

### CPU Profiling

```bash
# Build with profiling
go build -o bin/decube ./cmd/decube

# Run with profiling
./bin/decube -cpuprofile=cpu.prof

# Analyze
go tool pprof cpu.prof
```

### Memory Profiling

```bash
# Run with memory profiling
./bin/decube -memprofile=mem.prof

# Analyze
go tool pprof mem.prof
```

### Trace Analysis

```bash
# Generate trace
./bin/decube -trace=trace.out

# View trace
go tool trace trace.out
```

## Code Review Checklist

Before submitting a PR, ensure:

- [ ] Code follows style guidelines
- [ ] All tests pass
- [ ] New code has tests
- [ ] Documentation is updated
- [ ] No linter errors
- [ ] Commit messages follow conventions
- [ ] PR description is clear
- [ ] Related issues are referenced

## Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Testing Best Practices](https://golang.org/doc/effective_go#testing)
- [Conventional Commits](https://www.conventionalcommits.org/)

