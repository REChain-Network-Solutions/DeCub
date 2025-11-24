package crdt_test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/rechain/rechain/pkg/crdt"
)

func BenchmarkTwoPhaseSet(b *testing.B) {
	tests := []struct {
		name      string
		elements  int
		concurrent bool
	}{
		{"SingleThread-100", 100, false},
		{"SingleThread-1K", 1000, false},
		{"SingleThread-10K", 10000, false},
		{"Concurrent-100-4", 100, true},
		{"Concurrent-1K-4", 1000, true},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			set := crdt.NewTwoPhaseSet("benchmark-node")
			elements := generateElements(tt.elements)

			b.ResetTimer()

			if tt.concurrent {
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						for i := 0; i < tt.elements; i++ {
							if i%2 == 0 {
								set.Add(elements[i])
							} else {
								set.Remove(elements[i])
							}
						}
						_ = set.Elements()
					}
				})
			} else {
				for n := 0; n < b.N; n++ {
					for i := 0; i < tt.elements; i++ {
						if i%2 == 0 {
							set.Add(elements[i])
						} else {
							set.Remove(elements[i])
						}
					}
					_ = set.Elements()
				}
			}
		})
	}
}

func BenchmarkIDCounter(b *testing.B) {
	tests := []struct {
		name      string
		ops       int
		concurrent bool
	}{
		{"SingleThread-1K", 1000, false},
		{"SingleThread-10K", 10000, false},
		{"SingleThread-100K", 100000, false},
		{"Concurrent-10K-4", 10000, true},
		{"Concurrent-100K-4", 100000, true},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			counter := crdt.NewIDCounter("benchmark-node")
			ops := generateOperations(tt.ops)

			b.ResetTimer()

			if tt.concurrent {
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						for _, op := range ops {
							if op.Type == "increment" {
								counter.Increment(op.Value.(int64))
							} else {
								counter.Decrement(op.Value.(int64))
							}
						}
						_ = counter.Value()
					}
				})
			} else {
				for n := 0; n < b.N; n++ {
					for _, op := range ops {
						if op.Type == "increment" {
							counter.Increment(op.Value.(int64))
						} else {
							counter.Decrement(op.Value.(int64))
						}
					}
					_ = counter.Value()
				}
			}
		})
	}
}

func BenchmarkCRDTMerge(b *testing.B) {
	tests := []struct {
		name      string
		setSize   int
		numNodes  int
		concurrent bool
	}{
		{"2Nodes-1K", 1000, 2, false},
		{"4Nodes-1K", 1000, 4, false},
		{"8Nodes-1K", 1000, 8, false},
		{"2Nodes-10K", 10000, 2, false},
		{"Concurrent-4Nodes-1K", 1000, 4, true},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			nodes := make([]*crdt.TwoPhaseSet, tt.numNodes)
			elements := generateElements(tt.setSize)

			// Initialize nodes with different but overlapping data
			for i := 0; i < tt.numNodes; i++ {
				nodes[i] = crdt.NewTwoPhaseSet(fmt.Sprintf("node-%d", i))
				for j := 0; j < tt.setSize; j++ {
					// Each node gets a mix of unique and shared elements
					if j%(i+1) == 0 || rand.Intn(100) < 30 { // 30% chance of shared elements
						nodes[i].Add(elements[j])
					}
				}
			}

			b.ResetTimer()

			if tt.concurrent {
				var wg sync.WaitGroup
				for i := 0; i < b.N; i++ {
					// Merge all nodes into the first one
					for j := 1; j < tt.numNodes; j++ {
						wg.Add(1)
						go func(node *crdt.TwoPhaseSet) {
							_ = nodes[0].Merge(node)
							wg.Done()
						}(nodes[j])
					}
					wg.Wait()
				}
			} else {
				for n := 0; n < b.N; n++ {
					// Merge all nodes into the first one
					for j := 1; j < tt.numNodes; j++ {
						_ = nodes[0].Merge(nodes[j])
					}
				}
			}
		})
	}
}

// Helper functions

func generateElements(count int) []string {
	elements := make([]string, count)
	for i := 0; i < count; i++ {
		elements[i] = fmt.Sprintf("element-%d", i)
	}
	// Shuffle the elements
	rand.Shuffle(len(elements), func(i, j int) {
		elements[i], elements[j] = elements[j], elements[i]
	})
	return elements
}

func generateOperations(count int) []crdt.Operation {
	ops := make([]crdt.Operation, count)
	for i := 0; i < count; i++ {
		if i%2 == 0 {
			ops[i] = crdt.Operation{Type: "increment", Value: int64(rand.Intn(10) + 1)}
		} else {
			ops[i] = crdt.Operation{Type: "decrement", Value: int64(rand.Intn(5) + 1)}
		}
	}
	return ops
}

// Test helper to ensure benchmarks are correct
func TestBenchmarkHelpers(t *testing.T) {
	// Test generateElements
	elements := generateElements(10)
	if len(elements) != 10 {
		t.Errorf("generateElements(10) returned %d elements, want 10", len(elements))
	}

	// Test generateOperations
	ops := generateOperations(10)
	if len(ops) != 10 {
		t.Errorf("generateOperations(10) returned %d operations, want 10", len(ops))
	}
}
