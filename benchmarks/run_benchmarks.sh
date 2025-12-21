#!/bin/bash

# Vexilla Benchmark Runner
# Executes comprehensive benchmarks and generates formatted reports

set -e

echo "🚀 Vexilla Performance Benchmark Suite"
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create results directory
RESULTS_DIR="results"
mkdir -p "$RESULTS_DIR"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
RESULTS_FILE="$RESULTS_DIR/benchmark_$TIMESTAMP.txt"
SUMMARY_FILE="$RESULTS_DIR/summary_$TIMESTAMP.md"

echo -e "${BLUE}Results will be saved to: $RESULTS_FILE${NC}"
echo ""

# Check if benchstat is installed
if ! command -v benchstat &> /dev/null; then
    echo -e "${YELLOW}Installing benchstat for statistical analysis...${NC}"
    go install golang.org/x/perf/cmd/benchstat@latest
fi

# Run benchmarks
echo -e "${GREEN}Running benchmarks...${NC}"
echo "This may take a few minutes..."
echo ""

go test -bench=. -benchmem -benchtime=3s -timeout=30m | tee "$RESULTS_FILE"

# Generate summary
echo ""
echo -e "${GREEN}Generating summary report...${NC}"

cat > "$SUMMARY_FILE" <<EOF
# Vexilla Benchmark Results

**Date:** $(date)
**Go Version:** $(go version)
**OS:** $(uname -s)
**Architecture:** $(uname -m)

## Summary

EOF

# Parse results and generate summary
echo "### Local Evaluation Performance" >> "$SUMMARY_FILE"
echo "" >> "$SUMMARY_FILE"
echo "| Benchmark | Time/op | Allocs/op | Bytes/op |" >> "$SUMMARY_FILE"
echo "|-----------|---------|-----------|----------|" >> "$SUMMARY_FILE"

grep "BenchmarkLocalEvaluation" "$RESULTS_FILE" | while read -r line; do
    name=$(echo "$line" | awk '{print $1}' | sed 's/BenchmarkLocalEvaluation_//')
    time=$(echo "$line" | awk '{print $3}')
    bytes=$(echo "$line" | awk '{print $5}')
    allocs=$(echo "$line" | awk '{print $7}')
    echo "| $name | $time | $allocs | $bytes |" >> "$SUMMARY_FILE"
done

echo "" >> "$SUMMARY_FILE"
echo "### Cache Performance" >> "$SUMMARY_FILE"
echo "" >> "$SUMMARY_FILE"
echo "| Benchmark | Time/op | Allocs/op | Bytes/op |" >> "$SUMMARY_FILE"
echo "|-----------|---------|-----------|----------|" >> "$SUMMARY_FILE"

grep "BenchmarkCache\|BenchmarkStorage" "$RESULTS_FILE" | while read -r line; do
    name=$(echo "$line" | awk '{print $1}' | sed 's/Benchmark//')
    time=$(echo "$line" | awk '{print $3}')
    bytes=$(echo "$line" | awk '{print $5}')
    allocs=$(echo "$line" | awk '{print $7}')
    echo "| $name | $time | $allocs | $bytes |" >> "$SUMMARY_FILE"
done

echo "" >> "$SUMMARY_FILE"
echo "### Concurrent Performance" >> "$SUMMARY_FILE"
echo "" >> "$SUMMARY_FILE"
echo "| Benchmark | Time/op | Allocs/op | Bytes/op |" >> "$SUMMARY_FILE"
echo "|-----------|---------|-----------|----------|" >> "$SUMMARY_FILE"

grep "BenchmarkConcurrent" "$RESULTS_FILE" | while read -r line; do
    name=$(echo "$line" | awk '{print $1}' | sed 's/Benchmark//')
    time=$(echo "$line" | awk '{print $3}')
    bytes=$(echo "$line" | awk '{print $5}')
    allocs=$(echo "$line" | awk '{print $7}')
    echo "| $name | $time | $allocs | $bytes |" >> "$SUMMARY_FILE"
done

echo "" >> "$SUMMARY_FILE"
echo "### Large Scale Performance" >> "$SUMMARY_FILE"
echo "" >> "$SUMMARY_FILE"
echo "| Benchmark | Time/op | Allocs/op | Bytes/op |" >> "$SUMMARY_FILE"
echo "|-----------|---------|-----------|----------|" >> "$SUMMARY_FILE"

grep "BenchmarkLargeScale" "$RESULTS_FILE" | while read -r line; do
    name=$(echo "$line" | awk '{print $1}' | sed 's/Benchmark//')
    time=$(echo "$line" | awk '{print $3}')
    bytes=$(echo "$line" | awk '{print $5}')
    allocs=$(echo "$line" | awk '{print $7}')
    echo "| $name | $time | $allocs | $bytes |" >> "$SUMMARY_FILE"
done

echo "" >> "$SUMMARY_FILE"
echo "## Full Results" >> "$SUMMARY_FILE"
echo "" >> "$SUMMARY_FILE"
echo "\`\`\`" >> "$SUMMARY_FILE"
cat "$RESULTS_FILE" >> "$SUMMARY_FILE"
echo "\`\`\`" >> "$SUMMARY_FILE"

# Display summary
echo ""
echo -e "${GREEN}✅ Benchmarks complete!${NC}"
echo ""
echo "Results saved to:"
echo "  - Full results: $RESULTS_FILE"
echo "  - Summary: $SUMMARY_FILE"
echo ""

# Display key metrics
echo -e "${BLUE}Key Performance Metrics:${NC}"
echo ""

echo "Local Evaluation (Simple):"
grep "BenchmarkLocalEvaluation_Simple-" "$RESULTS_FILE" | head -1

echo ""
echo "Cache Hit:"
grep "BenchmarkCacheHit-" "$RESULTS_FILE" | head -1

echo ""
echo "Concurrent Evaluations:"
grep "BenchmarkConcurrentEvaluations-" "$RESULTS_FILE" | head -1

echo ""
echo "Memory Allocation:"
grep "BenchmarkMemoryAllocation-" "$RESULTS_FILE" | head -1

echo ""
echo -e "${YELLOW}View full summary: cat $SUMMARY_FILE${NC}"
