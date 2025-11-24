# Conflict-free Replicated Data Types (CRDTs)

This package provides implementations of various Conflict-free Replicated Data Types (CRDTs) that can be used to build distributed systems with eventual consistency guarantees.

## Available CRDTs

### 1. LWWRegister (Last-Write-Wins Register)
A register that keeps the most recent write based on timestamps.

```go
import "github.com/rechain/rechain/pkg/crdt"

// Create a new LWWRegister
register := crdt.NewLWWRegister("node1")

// Set a value
register.Set("hello")

// Get the current value
value := register.GetValue()

// Merge with another register
other := crdt.NewLWWRegister("node2")
other.Set("world")
_ = register.Merge(other)
```

### 2. PNCounter (Positive-Negative Counter)
A counter that supports both increments and decrements.

```go
counter := crdt.NewPNCounter("node1")

// Increment the counter
counter.Increment(5)

// Decrement the counter
counter.Decrement(2)

// Get the current value
value := counter.Value() // 3
```

### 3. ORSet (Observed-Removed Set)
A set that supports adding and removing elements with conflict resolution.

```go
set := crdt.NewORSet("node1")

// Add elements
set.Add("a")
set.Add("b")

// Remove an element
set.Remove("a")

// Check if an element exists
hasA := set.Contains("a") // false
hasB := set.Contains("b") // true

// Get all elements
elements := set.Elements()
```

### 4. GCounter (Grow-only Counter)
A counter that only supports increments.

```go
counter := crdt.NewGCounter("node1")

// Increment the counter
counter.Increment(3)

// Get the current value
value := counter.Value() // 3
```

### 5. TwoPhaseSet (2P-Set, CmRDT)
A set that allows an element to be added and removed, but not re-added after removal.

```go
set := crdt.NewTwoPhaseSet("node1")

// Add elements
set.Add("a")
set.Add("b")

// Remove an element
set.Remove("a")

// Trying to re-add a removed element has no effect
set.Add("a")
hasA := set.Contains("a") // false
```

### 6. IDCounter (Increment-Decrement Counter, CvRDT)
An operation-based counter that supports both increments and decrements with conflict resolution.

```go
counter := crdt.NewIDCounter("node1")

// Increment and decrement
counter.Increment(5)
counter.Decrement(2)

// Get the current value
value := counter.Value() // 3

// Apply operations from a log
op1 := crdt.Operation{Type: "increment", Value: int64(3)}
op2 := crdt.Operation{Type: "decrement", Value: int64(1)}
_ = counter.ApplyOperation(op1)
_ = counter.ApplyOperation(op2)
value = counter.Value() // 5
```

## Usage in Distributed Systems

### Merging States

```go
// On node 1
node1 := crdt.NewPNCounter("node1")
node1.Increment(5)

// On node 2
node2 := crdt.NewPNCounter("node2")
node2.Increment(3)

// When nodes synchronize
_ = node1.Merge(node2)
_ = node2.Merge(node1)

// Both nodes now have the same state
node1.Value() // 8
node2.Value() // 8
```

### Serialization

```go
// Marshal to bytes
data, err := counter.Marshal()
if err != nil {
    // handle error
}

// Unmarshal to a new instance
newCounter := &crdt.PNCounter{}
err = newCounter.Unmarshal(data)
```

## Implementation Details

### LWWRegister (CvRDT)
- Uses timestamps to resolve conflicts
- Last write wins in case of concurrent updates
- Node ID is used as a tiebreaker for identical timestamps

### PNCounter (CvRDT)
- Uses separate counters for increments and decrements
- Each node maintains its own counters
- Merge takes the maximum value for each node's counters

### ORSet (CvRDT)
- Uses unique tags to track additions and removals
- An element is in the set if it has at least one add tag that's not in the remove set
- Supports concurrent adds and removes

### GCounter (CvRDT)
- Each node maintains its own counter
- Merge takes the maximum value for each node's counter
- Only supports increments, ensuring values always increase

### TwoPhaseSet (CmRDT)
- Uses two sets: one for added elements and one for removed elements
- Once an element is removed, it cannot be re-added
- Merge is a union of both added and removed sets
- Optimized for read-heavy workloads with caching

### IDCounter (CvRDT)
- Uses separate counters for increments and decrements per node
- Merge takes the maximum value for each node's increment and decrement counters
- Supports operation-based replication
- Optimized for high concurrency with lock-free reads

## Performance Characteristics

### TwoPhaseSet
- **Add/Remove**: O(1) average case
- **Contains**: O(1) average case
- **Merge**: O(n + m) where n and m are the sizes of the sets being merged
- **Memory**: O(n + m) where n is the number of added elements and m is the number of removed elements
- **Concurrency**: Optimized for concurrent reads with `sync.RWMutex` and `sync.Map`

### IDCounter
- **Increment/Decrement**: O(1) average case
- **Value**: O(n) where n is the number of nodes (cached for 100ms)
- **Merge**: O(n + m) where n and m are the number of nodes in each counter
- **Memory**: O(n) where n is the number of nodes
- **Concurrency**: Optimized for concurrent operations with `sync.Map` and lock-free reads

## Benchmarks

Run benchmarks with:

```bash
# Run all benchmarks
go test -bench=. -benchmem

# Run specific benchmark
go test -bench=BenchmarkTwoPhaseSet -benchmem
go test -bench=BenchmarkIDCounter -benchmem
go test -bench=BenchmarkCRDTMerge -benchmem
```

### Example Benchmark Results

```
BenchmarkTwoPhaseSet/SingleThread-100-8              50000     23456 ns/op    12345 B/op     456 allocs/op
BenchmarkTwoPhaseSet/Concurrent-1K-4-8              10000    123456 ns/op    56789 B/op    1234 allocs/op
BenchmarkIDCounter/SingleThread-10K-8               20000     98765 ns/op    23456 B/op     789 allocs/op
BenchmarkIDCounter/Concurrent-100K-4-8               2000   1234567 ns/op   123456 B/op    4567 allocs/op
BenchmarkCRDTMerge/4Nodes-1K-8                       5000    345678 ns/op   234567 B/op    6789 allocs/op
```

## Best Practices

1. **Choose the right CRDT for your use case**:
   - Use `LWWRegister` for simple values where last write should win
   - Use `PNCounter` or `IDCounter` for counting with increments/decrements
   - Use `ORSet` or `TwoPhaseSet` for collections with add/remove operations

2. **Optimize for your workload**:
   - For read-heavy workloads, consider enabling caching where available
   - For write-heavy workloads, batch operations when possible
   - For large clusters, consider sharding by key or node

3. **Error Handling**:
   - Always check errors from `Merge()` and `Unmarshal()` operations
   - Handle network partitions and node failures gracefully

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
