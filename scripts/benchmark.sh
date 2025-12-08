#!/bin/bash

echo "⚡ Running benchmarks..."

echo ""
echo "=== Evaluator Benchmarks ==="
go test -bench=. -benchmem ./pkg/evaluator/

echo ""
echo "=== Storage Benchmarks ==="
go test -bench=. -benchmem ./pkg/storage/

echo ""
echo "=== Circuit Breaker Benchmarks ==="
go test -bench=. -benchmem ./pkg/circuit/

echo ""
echo "✅ Benchmarks complete!"
