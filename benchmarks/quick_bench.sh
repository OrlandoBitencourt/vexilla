#!/bin/bash

# Quick benchmark runner for fast results
# Use this for quick testing, not for official documentation

echo "🏃 Quick Benchmark Run"
echo "====================="
echo ""
echo "Running abbreviated benchmarks (1s each)..."
echo "For official results, use: ./run_benchmarks.sh"
echo ""

# Run quick benchmarks
go test -bench=. -benchmem -benchtime=1s -timeout=5m | tee quick_results.txt

echo ""
echo "Quick summary:"
echo ""

# Extract key metrics
echo "Local Evaluation:"
grep "BenchmarkLocalEvaluation_Simple" quick_results.txt | head -1

echo ""
echo "Cache Hit:"
grep "BenchmarkCacheHit" quick_results.txt | head -1

echo ""
echo "Concurrent:"
grep "BenchmarkConcurrentEvaluations" quick_results.txt | head -1

echo ""
echo "Full results saved to: quick_results.txt"
