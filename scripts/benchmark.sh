#!/bin/bash
set -e

# DeCube Benchmark Script
# Runs performance benchmarks

echo "DeCube Benchmark Suite"
echo "======================"
echo ""

# Run Go benchmarks
echo "Running Go benchmarks..."
for dir in rechain/pkg/crdt rechain/pkg/merkle rechain/internal/storage; do
    if [ -d "$dir" ]; then
        echo "  Benchmarking $dir..."
        cd "$dir"
        go test -bench=. -benchmem -benchtime=5s ./... > "../../../benchmarks/$(basename $dir).txt" 2>&1 || true
        cd - > /dev/null
    fi
done

# Run API benchmarks
echo "Running API benchmarks..."
if command -v ab >/dev/null 2>&1; then
    echo "  Testing REST API..."
    ab -n 10000 -c 100 http://localhost:8080/health > benchmarks/api-rest.txt 2>&1 || true
fi

# Run storage benchmarks
echo "Running storage benchmarks..."
if [ -f "tests/storage_benchmark_test.go" ]; then
    go test -bench=BenchmarkStorage -benchmem ./tests/... > benchmarks/storage.txt 2>&1 || true
fi

echo ""
echo "Benchmark complete!"
echo "Results saved to benchmarks/"

