package crdt_test

import (
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/rechain/rechain/pkg/crdt"
)

// BenchmarkLWWRegister benchmarks the LWWRegister operations
func BenchmarkLWWRegister(b *testing.B) {
	nodeID := "benchmark-node"
	reg := crdt.NewLWWRegister(nodeID)

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			reg.Set(fmt.Sprintf("value-%d", i))
		}
	})

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = reg.GetValue()
		}
	})

	b.Run("Merge", func(b *testing.B) {
		other := crdt.NewLWWRegister("other-node")
		other.Set("other-value")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = reg.Merge(other)
		}
	})
}

// BenchmarkPNCounter benchmarks the PNCounter operations
func BenchmarkPNCounter(b *testing.B) {
	nodeID := "benchmark-node"
	counter := crdt.NewPNCounter(nodeID)

	b.Run("Increment", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			counter.Increment(1)
		}
	})

	b.Run("Decrement", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			counter.Decrement(1)
		}
	})

	b.Run("Value", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = counter.Value()
		}
	})

	b.Run("Merge", func(b *testing.B) {
		other := crdt.NewPNCounter("other-node")
		other.Increment(5)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = counter.Merge(other)
		}
	})
}

// BenchmarkORSet benchmarks the ORSet operations
func BenchmarkORSet(b *testing.B) {
	nodeID := "benchmark-node"
	set := crdt.NewORSet(nodeID)

	b.Run("Add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			set.Add(fmt.Sprintf("item-%d", i))
		}
	})

	b.Run("Remove", func(b *testing.B) {
		// First add some items to remove
		for i := 0; i < 1000; i++ {
			set.Add(fmt.Sprintf("item-%d", i))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			set.Remove(fmt.Sprintf("item-%d", i%1000))
		}
	})

	b.Run("Contains", func(b *testing.B) {
		// Add an item to check for
		set.Add("test-item")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = set.Contains("test-item")
		}
	})

	b.Run("Merge", func(b *testing.B) {
		other := crdt.NewORSet("other-node")
		for i := 0; i < 100; i++ {
			other.Add(fmt.Sprintf("other-item-%d", i))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = set.Merge(other)
		}
	})
}

// BenchmarkGCounter benchmarks the GCounter operations
func BenchmarkGCounter(b *testing.B) {
	nodeID := "benchmark-node"
	counter := crdt.NewGCounter(nodeID)

	b.Run("Increment", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			counter.Increment(1)
		}
	})

	b.Run("Value", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = counter.Value()
		}
	})

	b.Run("Merge", func(b *testing.B) {
		other := crdt.NewGCounter("other-node")
		other.Increment(5)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = counter.Merge(other)
		}
	})
}

// BenchmarkCRDTConcurrent benchmarks concurrent operations on CRDTs
func BenchmarkCRDTConcurrent(b *testing.B) {
	numGoroutines := []int{1, 2, 4, 8, 16, 32, 64, 128}

	for _, num := range numGoroutines {
		b.Run(fmt.Sprintf("LWWRegister-%d", num), func(b *testing.B) {
			reg := crdt.NewLWWRegister("benchmark-node")
			benchmarkConcurrent(b, num, func(i int) {
				reg.Set(fmt.Sprintf("value-%d", i))
			})
		})

		b.Run(fmt.Sprintf("PNCounter-%d", num), func(b *testing.B) {
			counter := crdt.NewPNCounter("benchmark-node")
			benchmarkConcurrent(b, num, func(i int) {
				if i%2 == 0 {
					counter.Increment(1)
				} else {
					counter.Decrement(1)
				}
			})
		})

		b.Run(fmt.Sprintf("ORSet-%d", num), func(b *testing.B) {
			set := crdt.NewORSet("benchmark-node")
			benchmarkConcurrent(b, num, func(i int) {
				item := fmt.Sprintf("item-%d", i%1000)
				if i%2 == 0 {
					set.Add(item)
				} else {
					set.Remove(item)
				}
			})
		})

		b.Run(fmt.Sprintf("GCounter-%d", num), func(b *testing.B) {
			counter := crdt.NewGCounter("benchmark-node")
			benchmarkConcurrent(b, num, func(i int) {
				counter.Increment(1)
			})
		})
	}
}

// benchmarkConcurrent runs a benchmark with concurrent goroutines
func benchmarkConcurrent(b *testing.B, numGoroutines int, fn func(int)) {
	b.Helper()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for g := 0; g < numGoroutines; g++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					fn(id*1000 + j)
				}
			}(g)
		}

		wg.Wait()
	}
}

// BenchmarkCRDTSerialization benchmarks the serialization/deserialization of CRDTs
func BenchmarkCRDTSerialization(b *testing.B) {
	// Create sample CRDTs
	lww := crdt.NewLWWRegister("node1")
	lww.Set("test-value")

	pn := crdt.NewPNCounter("node1")
	pn.Increment(5)
	pn.Decrement(2)

	orset := crdt.NewORSet("node1")
	for i := 0; i < 100; i++ {
		orset.Add(fmt.Sprintf("item-%d", i))
	}

	gcounter := crdt.NewGCounter("node1")
	gcounter.Increment(10)

	benchmarks := []struct {
		name  string
		crdt  crdt.CRDT
		value interface{}
	}{
		{"LWWRegister", lww, "test-value"},
		{"PNCounter", pn, int64(3)},
		{"ORSet", orset, 100},
		{"GCounter", gcounter, int64(10)},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name+"/Marshal", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = bm.crdt.Marshal()
			}
		})

		data, _ := bm.crdt.Marshal()

		b.Run(bm.name+"/Unmarshal", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tmp, _ := crdt.New(bm.crdt.Type(), "temp")
				_ = tmp.Unmarshal(data)
			}
		})

		b.Run(bm.name+"/Roundtrip", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				data, _ := bm.crdt.Marshal()
				tmp, _ := crdt.New(bm.crdt.Type(), "temp")
				_ = tmp.Unmarshal(data)
			}
		})
	}
}

// BenchmarkCRDTMemoryUsage benchmarks the memory usage of CRDTs
func BenchmarkCRDTMemoryUsage(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000, 100000}

	for _, size := range sizes {
		size := size // capture loop variable

		b.Run("ORSet/Size/"+strconv.Itoa(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				set := crdt.NewORSet("node1")
				for j := 0; j < size; j++ {
					set.Add(fmt.Sprintf("item-%d", j))
				}
				_ = set
			}
		})

		b.Run("GCounter/Size/"+strconv.Itoa(size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				counter := crdt.NewGCounter("node1")
				for j := 0; j < size; j++ {
					counter.Increment(1)
				}
				_ = counter
			}
		})
	}
}
