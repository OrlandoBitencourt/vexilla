# Benchmark Execution Guide

Step-by-step guide to run benchmarks and update documentation with real performance data.

## Prerequisites

1. **Go 1.21+** installed
2. **Flagr** running (optional for some benchmarks)
3. **Stable environment** - close other applications

## Quick Start

### Windows

```cmd
cd benchmarks
run_benchmarks.bat
```

### Linux/macOS

```bash
cd benchmarks
chmod +x run_benchmarks.sh
./run_benchmarks.sh
```

## Step-by-Step Process

### 1. Prepare Environment

```bash
# Close unnecessary applications
# Disable CPU frequency scaling (Linux)
sudo cpupower frequency-set --governor performance

# Check system load
top

# Ensure clean state
go clean -testcache
```

### 2. Run Benchmarks

```bash
cd benchmarks

# Quick test (30 seconds)
go test -bench=. -benchmem -benchtime=1s

# Standard run (5-10 minutes)
go test -bench=. -benchmem -benchtime=3s

# Comprehensive run (20-30 minutes)
go test -bench=. -benchmem -benchtime=10s -count=5
```

### 3. Save Results

```bash
# Save to file
go test -bench=. -benchmem -benchtime=3s > results/benchmark_$(date +%Y%m%d).txt

# View results
cat results/benchmark_$(date +%Y%m%d).txt
```

### 4. Analyze Results

```bash
# Statistical analysis with benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Compare runs
benchstat results/old.txt results/new.txt

# Generate report
benchstat -html results/benchmark_*.txt > results/report.html
```

## Understanding Results

### Example Output

```
BenchmarkLocalEvaluation_Simple-8    500000    3542 ns/op    128 B/op    2 allocs/op
```

**Breakdown:**
- `BenchmarkLocalEvaluation_Simple`: Benchmark name
- `-8`: GOMAXPROCS (number of CPU cores)
- `500000`: Number of iterations
- `3542 ns/op`: **3.542 microseconds per operation**
- `128 B/op`: 128 bytes allocated per operation
- `2 allocs/op`: 2 memory allocations per operation

### Converting Units

| Unit | Conversion | Example |
|------|------------|---------|
| ns/op | ÷ 1,000 = μs/op | 3,542 ns = 3.542 μs |
| ns/op | ÷ 1,000,000 = ms/op | 1,500,000 ns = 1.5 ms |
| ops/sec | 1,000,000,000 ÷ ns/op | 3,542 ns → 282,000 ops/sec |

## Updating Documentation

### 1. Extract Key Metrics

From benchmark results, extract these metrics:

```bash
# Local evaluation latency
grep "BenchmarkLocalEvaluation_Simple" results/latest.txt

# Cache hit latency
grep "BenchmarkCacheHit" results/latest.txt

# Throughput (concurrent)
grep "BenchmarkConcurrentEvaluations" results/latest.txt

# Memory allocation
grep "BenchmarkMemoryAllocation" results/latest.txt
```

### 2. Update README.md

Open `../README.md` and update the Performance section:

```markdown
## Performance

Based on benchmarks run on [YOUR_HARDWARE]:

### Latency

- **Local evaluation**: X μs (Y ops/sec)
- **Cache hit**: X μs
- **Remote evaluation**: 50-200ms (depends on Flagr)

### Throughput

- **Single-threaded**: X,XXX evaluations/sec
- **Multi-threaded**: X,XXX,XXX evaluations/sec

### Memory

- **Per evaluation**: X bytes, Y allocations
- **Per flag cached**: ~800 bytes

[Full benchmark results](benchmarks/results/latest.txt)
```

### 3. Update ARCHITECTURE.md

Open `../ARCHITECTURE.md` and update the Performance Characteristics section:

```markdown
## Performance Characteristics

### Latency Comparison

| Operation | Latency | HTTP Requests | Notes |
|-----------|---------|---------------|-------|
| Local evaluation (static) | X μs | 0 | Actual benchmark result |
| Remote evaluation (dynamic) | 50-200ms | 1 | Full Flagr evaluation |
| Cache hit (Ristretto) | X μs | 0 | In-memory lookup |

### Throughput

\`\`\`
Benchmark Results ([YOUR_HARDWARE]):

Local Evaluation (Static Flags):
- 10,000 evaluations: ~Xms
- Average: ~X μs per evaluation
- Throughput: ~XXX,XXX eval/sec
```

### 4. Create Performance Badge

Add to main README:

```markdown
[![Benchmark](https://img.shields.io/badge/benchmark-XXX%20μs%2Fop-brightgreen)]()
```

## Hardware Specifications Template

When documenting results, include hardware specs:

```markdown
**Test Environment:**
- **CPU**: [e.g., Apple M1, Intel i7-12700K, AMD Ryzen 9 5950X]
- **RAM**: [e.g., 16GB DDR4-3200]
- **OS**: [e.g., macOS 14.0, Ubuntu 22.04, Windows 11]
- **Go Version**: [e.g., 1.21.5]
- **Date**: [YYYY-MM-DD]
```

## Benchmark Scenarios

### Scenario 1: Local Evaluation Performance

**Goal**: Measure pure local evaluation speed

```bash
go test -bench=BenchmarkLocalEvaluation -benchmem -benchtime=5s
```

**Expected ranges:**
- Simple: 1-10 μs/op
- With constraints: 5-20 μs/op
- Complex: 10-50 μs/op

### Scenario 2: Cache Performance

**Goal**: Measure cache hit/miss performance

```bash
go test -bench=BenchmarkCache -benchmem -benchtime=5s
```

**Expected ranges:**
- Cache hit: <1 μs/op
- Storage get: <1 μs/op
- Storage set: 1-10 μs/op

### Scenario 3: Concurrent Performance

**Goal**: Measure scalability under load

```bash
go test -bench=BenchmarkConcurrent -benchmem -benchtime=5s
```

**Expected:**
- Should scale near-linearly with cores
- Minimal lock contention

### Scenario 4: Large Scale

**Goal**: Measure performance with many flags

```bash
go test -bench=BenchmarkLargeScale -benchmem -benchtime=5s
```

**Expected:**
- 1,000 flags: similar performance to 100
- 10,000 flags: <2x slower than 100

## Troubleshooting

### High Variance

**Symptom**: Results vary significantly between runs

**Solutions:**
```bash
# Increase benchmark time
go test -bench=. -benchtime=10s

# Multiple runs with benchstat
go test -bench=. -count=10 > results.txt
benchstat results.txt
```

### Unexpected Slowness

**Symptom**: Results much slower than expected

**Check:**
1. System load: `top` or Task Manager
2. CPU throttling: Check power settings
3. Background processes: Close applications
4. GC interference: Add `-gcflags="-m"`

### Out of Memory

**Symptom**: Benchmarks crash with OOM

**Solutions:**
```bash
# Reduce benchmark time
go test -bench=. -benchtime=1s

# Run specific benchmarks
go test -bench=BenchmarkLocalEvaluation

# Increase memory limit (if needed)
GOMEMLIMIT=4GiB go test -bench=.
```

## Continuous Benchmarking

### GitHub Actions Workflow

Create `.github/workflows/benchmark.yml`:

```yaml
name: Benchmarks

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  benchmark:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run Benchmarks
        run: |
          cd benchmarks
          go test -bench=. -benchmem -benchtime=3s > results.txt
          cat results.txt

      - name: Comment PR with Results
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v6
        with:
          script: |
            const fs = require('fs');
            const results = fs.readFileSync('benchmarks/results.txt', 'utf8');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '## Benchmark Results\n\n```\n' + results + '\n```'
            });
```

## Best Practices

1. **Run multiple times**: At least 3 runs, use benchstat
2. **Stable environment**: Close applications, consistent power
3. **Document hardware**: Always include specs
4. **Version control**: Save results with git commit SHA
5. **Compare fairly**: Same hardware, same Go version
6. **Realistic data**: Use production-like flag configurations

## Reporting Issues

When reporting performance issues, include:

1. Full benchmark output
2. Hardware specifications
3. Go version: `go version`
4. OS and version
5. Steps to reproduce
6. Comparison with expected results

## Example Full Report

```markdown
# Performance Report - 2024-01-15

## Environment

- **CPU**: Apple M1 Pro (10 cores)
- **RAM**: 16GB LPDDR5
- **OS**: macOS 14.2
- **Go**: 1.21.5
- **Commit**: abc123f

## Results

### Local Evaluation

| Benchmark | Time/op | Allocs/op | Bytes/op |
|-----------|---------|-----------|----------|
| Simple | 2.847 μs | 2 | 128 B |
| WithConstraints | 8.234 μs | 5 | 384 B |
| Complex | 18.456 μs | 12 | 896 B |

### Throughput

- Single-threaded: 351,000 eval/sec
- Multi-threaded: 2,850,000 eval/sec

## Analysis

✅ All metrics within expected ranges
✅ Linear scaling with CPU cores
✅ Memory allocations minimal
```

## Next Steps

After running benchmarks:

1. ✅ Save results to `results/` directory
2. ✅ Update README.md with key metrics
3. ✅ Update ARCHITECTURE.md with detailed data
4. ✅ Commit results to git
5. ✅ Create PR with performance improvements (if applicable)
