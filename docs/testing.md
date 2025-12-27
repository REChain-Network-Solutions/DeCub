# Testing Guide

This guide covers testing strategies and procedures for DeCube.

## Table of Contents

1. [Testing Strategy](#testing-strategy)
2. [Unit Testing](#unit-testing)
3. [Integration Testing](#integration-testing)
4. [End-to-End Testing](#end-to-end-testing)
5. [Performance Testing](#performance-testing)
6. [Chaos Testing](#chaos-testing)

## Testing Strategy

### Testing Pyramid

```
        /\
       /  \
      / E2E \        (Few, slow, expensive)
     /--------\
    /          \
   / Integration \  (Some, medium speed)
  /--------------\
 /                \
/   Unit Tests     \  (Many, fast, cheap)
/------------------\
```

### Test Types

- **Unit Tests**: Fast, isolated component tests
- **Integration Tests**: Component interaction tests
- **End-to-End Tests**: Full system tests
- **Performance Tests**: Load and stress tests
- **Chaos Tests**: Resilience and fault tolerance

## Unit Testing

### Go Unit Tests

#### Example

```go
package crdt

import (
    "testing"
)

func TestORSetAdd(t *testing.T) {
    orset := NewORSet()
    
    orset.Add("item1")
    
    if !orset.Contains("item1") {
        t.Errorf("Expected item1 to be in set")
    }
}
```

#### Running Tests

```bash
# Run all unit tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./rechain/pkg/crdt

# Run with verbose output
go test -v ./...
```

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    // Arrange
    setup := createTestSetup()
    
    // Act
    result := functionUnderTest(setup)
    
    // Assert
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

### Table-Driven Tests

```go
func TestORSetAdd(t *testing.T) {
    tests := []struct {
        name  string
        items []string
        want  bool
    }{
        {
            name:  "single item",
            items: []string{"item1"},
            want:  true,
        },
        {
            name:  "multiple items",
            items: []string{"item1", "item2"},
            want:  true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            orset := NewORSet()
            for _, item := range tt.items {
                orset.Add(item)
            }
            // Assertions...
        })
    }
}
```

## Integration Testing

### Component Integration

#### Example

```go
func TestCatalogSnapshotIntegration(t *testing.T) {
    // Setup
    catalog := NewCatalog()
    snapshotService := NewSnapshotService()
    
    // Create snapshot
    snapshot, err := snapshotService.Create("snapshot-001", nil)
    if err != nil {
        t.Fatal(err)
    }
    
    // Register in catalog
    err = catalog.Register(snapshot)
    if err != nil {
        t.Fatal(err)
    }
    
    // Verify
    found, err := catalog.Get("snapshot-001")
    if err != nil {
        t.Fatal(err)
    }
    
    if found.ID != snapshot.ID {
        t.Errorf("Expected %s, got %s", snapshot.ID, found.ID)
    }
}
```

### Running Integration Tests

```bash
# Run integration tests
go test -tags=integration ./tests/...

# With Docker services
docker-compose up -d
go test -tags=integration ./tests/...
docker-compose down
```

## End-to-End Testing

### E2E Test Example

```go
func TestSnapshotLifecycle(t *testing.T) {
    // Start services
    services := startServices(t)
    defer services.Stop()
    
    // Create snapshot
    snapshot, err := createSnapshot(services.Client, "e2e-snapshot")
    if err != nil {
        t.Fatal(err)
    }
    
    // Verify creation
    found, err := getSnapshot(services.Client, snapshot.ID)
    if err != nil {
        t.Fatal(err)
    }
    
    // Delete snapshot
    err = deleteSnapshot(services.Client, snapshot.ID)
    if err != nil {
        t.Fatal(err)
    }
    
    // Verify deletion
    _, err = getSnapshot(services.Client, snapshot.ID)
    if err == nil {
        t.Error("Expected snapshot to be deleted")
    }
}
```

### E2E Test Framework

```go
type TestServices struct {
    Client    *Client
    Catalog   *CatalogService
    Storage   *StorageService
    Cleanup   func()
}

func startServices(t *testing.T) *TestServices {
    // Start Docker services
    // Create clients
    // Return test services
}
```

## Performance Testing

### Benchmark Tests

```go
func BenchmarkORSetAdd(b *testing.B) {
    orset := NewORSet()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        orset.Add(fmt.Sprintf("item-%d", i))
    }
}
```

#### Running Benchmarks

```bash
# Run benchmarks
go test -bench=. -benchmem ./...

# Compare benchmarks
go test -bench=. -benchmem -benchcmp old.txt new.txt
```

### Load Testing

#### Using Apache Bench

```bash
# Basic load test
ab -n 10000 -c 100 http://localhost:8080/catalog/snapshots

# With authentication
ab -n 10000 -c 100 \
   -H "Authorization: Bearer $TOKEN" \
   http://localhost:8080/catalog/snapshots
```

#### Using wrk

```bash
# Load test
wrk -t4 -c100 -d30s http://localhost:8080/catalog/snapshots

# With Lua script
wrk -t4 -c100 -d30s -s script.lua http://localhost:8080/catalog/snapshots
```

### Stress Testing

```bash
# Stress test script
#!/bin/bash
for i in {1..10}; do
    ab -n 1000 -c 50 http://localhost:8080/catalog/snapshots &
done
wait
```

## Chaos Testing

### Chaos Scenarios

#### Network Partition

```go
func TestNetworkPartition(t *testing.T) {
    // Create cluster
    cluster := createCluster(t, 3)
    
    // Partition network
    partitionNetwork(cluster, 1)
    
    // Verify cluster continues operating
    // Verify eventual consistency
}
```

#### Node Failure

```go
func TestNodeFailure(t *testing.T) {
    cluster := createCluster(t, 5)
    
    // Kill one node
    killNode(cluster, 1)
    
    // Verify cluster continues
    // Verify data consistency
}
```

#### Slow Network

```go
func TestSlowNetwork(t *testing.T) {
    cluster := createCluster(t, 3)
    
    // Simulate slow network
    slowNetwork(cluster, 100*time.Millisecond)
    
    // Verify system handles delays
}
```

### Chaos Engineering Tools

- **Chaos Monkey**: Random instance termination
- **Network Chaos**: Network partition simulation
- **Pod Chaos**: Kubernetes pod failures

## Test Coverage

### Coverage Goals

- **Unit Tests**: >80% coverage
- **Integration Tests**: >60% coverage
- **Critical Paths**: 100% coverage

### Generating Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage
go tool cover -html=coverage.out

# Coverage by package
go test -cover ./... | grep coverage
```

## Continuous Testing

### CI/CD Integration

```yaml
# .github/workflows/test.yml
- name: Run tests
  run: |
    go test -v -race -coverprofile=coverage.out ./...
    
- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    file: ./coverage.out
```

### Test Automation

```bash
# Run all tests
make test

# Run specific test suite
make test-unit
make test-integration
make test-e2e
```

## Best Practices

### Test Organization

1. **Mirror Source Structure**
   - Tests alongside source
   - `*_test.go` files
   - Test packages

2. **Test Naming**
   - `TestFunctionName`
   - `TestFunctionName_Scenario`
   - Descriptive names

3. **Test Isolation**
   - Independent tests
   - No shared state
   - Cleanup after tests

### Test Data

1. **Use Fixtures**
   - Reusable test data
   - Consistent test setup
   - Easy to maintain

2. **Mock External Services**
   - Isolate unit tests
   - Fast execution
   - Predictable behavior

3. **Test Utilities**
   - Helper functions
   - Test builders
   - Common setup

## References

- [Development Guide](development.md)
- [Performance Tuning](performance-tuning.md)
- [Troubleshooting Guide](troubleshooting.md)

---

*Last updated: January 2024*

