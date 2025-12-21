# Vexilla Performance Benchmarks

Comprehensive benchmarks for measuring Vexilla's performance characteristics and generating real-world performance data for documentation.

## Quick Start

### Windows
```cmd
cd benchmarks
run_benchmarks.bat
```

### Linux/macOS
```bash
cd benchmarks
./run_benchmarks.sh
```

### Quick Test (1-2 minutes)
```bash
./quick_bench.sh  # Fast abbreviated results
```

## File Structure

```
benchmarks/
├── benchmark_test.go          # All benchmark implementations
├── README.md                  # This file
├── BENCHMARK_GUIDE.md         # Detailed execution guide
├── RESULTS_TEMPLATE.md        # Template for documenting results
├── run_benchmarks.sh          # Full benchmark runner (Linux/macOS)
├── run_benchmarks.bat         # Full benchmark runner (Windows)
├── quick_bench.sh             # Quick test runner
├── results/                   # Benchmark results directory
│   ├── example_results.md     # Example of completed results
│   └── benchmark_*.txt        # Raw benchmark outputs
└── .gitignore
```

## Running Benchmarks

### Run All Benchmarks

```bash
cd benchmarks
go test -bench=. -benchmem -benchtime=3s
```

### Run Specific Benchmarks

```bash
# Local evaluation
go test -bench=BenchmarkLocalEvaluation -benchmem

# Cache performance
go test -bench=BenchmarkCache -benchmem

# Concurrent performance
go test -bench=BenchmarkConcurrent -benchmem

# Storage operations
go test -bench=BenchmarkStorage -benchmem

# Large scale
go test -bench=BenchmarkLargeScale -benchmem
```

### Generate CPU Profile

```bash
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

### Generate Memory Profile

```bash
go test -bench=. -memprofile=mem.prof
go tool pprof mem.prof
```

### Generate Detailed Report

```bash
go test -bench=. -benchmem -benchtime=10s > benchmark_results.txt
```

## Benchmark Categories

### 1. Local Evaluation Benchmarks

Measure local flag evaluation performance:

- **Simple**: Boolean flag without constraints
- **WithConstraints**: Flag with 2-3 constraints
- **MultipleSegments**: Flag with 3+ segments
- **ComplexConstraints**: Flag with 5+ complex constraints

**Expected Results:**
- Simple: 1-5 μs/op, 0-2 allocs/op
- With constraints: 2-10 μs/op, 0-5 allocs/op
- Complex: 5-20 μs/op, 5-15 allocs/op

### 2. Cache Performance Benchmarks

Measure cache hit/miss performance:

- **CacheHit**: Ristretto lookup performance
- **ConcurrentEvaluations**: Parallel access patterns
- **StorageSet**: Write operation performance
- **StorageGet**: Read operation performance

**Expected Results:**
- Cache hit: <1 μs/op, 0 allocs/op
- Storage get: <1 μs/op, 0 allocs/op
- Storage set: 1-5 μs/op, 2-5 allocs/op

### 3. Deterministic Rollout Benchmarks

Measure bucket-based evaluation:

- **DeterministicRollout**: Bucket constraint matching

**Expected Results:**
- 2-10 μs/op, 0-5 allocs/op
- Zero HTTP requests

### 4. Large Scale Benchmarks

Measure performance with many flags:

- **1000Flags**: 1,000 cached flags
- **10000Flags**: 10,000 cached flags

**Expected Results:**
- Performance should scale linearly
- Memory usage: ~800 bytes per flag

### 5. Client API Benchmarks

Measure public API performance:

- **Bool()**: Simple boolean evaluation
- **Evaluate()**: Detailed evaluation

**Expected Results:**
- Bool: 1-10 μs/op
- Evaluate: 2-15 μs/op

## Performance Targets

### Latency Targets

| Operation | Target | Acceptable | Notes |
|-----------|--------|------------|-------|
| Local eval (simple) | <5 μs | <20 μs | No constraints |
| Local eval (complex) | <20 μs | <50 μs | 5+ constraints |
| Cache hit | <1 μs | <5 μs | Ristretto lookup |
| Storage get | <1 μs | <5 μs | Memory read |
| Storage set | <5 μs | <20 μs | Memory write |

### Throughput Targets

| Operation | Target | Acceptable | Notes |
|-----------|--------|------------|-------|
| Local evaluations | >200k/sec | >100k/sec | Single core |
| Concurrent evals | >1M/sec | >500k/sec | Multi-core |
| Cache operations | >10M/sec | >5M/sec | Ristretto |

### Memory Targets

| Operation | Target | Acceptable | Notes |
|-----------|--------|------------|-------|
| Per evaluation | <5 allocs | <10 allocs | Steady state |
| Per flag cached | <1 KB | <2 KB | Average |
| Cache overhead | <10% | <20% | Of total memory |

## Documentation Workflow

### 1. Run Benchmarks

```bash
# Full official run (20-30 min)
./run_benchmarks.sh

# Quick test (1-2 min)
./quick_bench.sh
```

### 2. Document Results

```bash
# Copy template
cp RESULTS_TEMPLATE.md results/results_$(date +%Y%m%d).md

# Fill in with your results
# See results/example_results.md for reference
```

### 3. Update Main Documentation

Extract key metrics and update:

**README.md** (project root):
```markdown
### Performance

- **Local evaluation**: X.XX μs
- **Throughput**: XXX,XXX ops/sec
- **Cache hit**: X.XX μs
```

**ARCHITECTURE.md**:
```markdown
### Throughput

Benchmark Results (Hardware):
- Throughput: ~XXX,XXX eval/sec
```

### 4. Commit Results

```bash
git add results/
git commit -m "Add benchmark results for $(date +%Y-%m-%d)"
```

## Understanding Results

### Example Output
```
BenchmarkLocalEvaluation_Simple-8    500000    3542 ns/op    128 B/op    2 allocs/op
```

**Conversions**:
- `3542 ns/op` = `3.542 μs/op` (÷ 1000)
- Throughput = `1,000,000,000 ÷ 3542` = `282,326 ops/sec`

### Performance Targets

| Metric | Target | Excellent | Acceptable |
|--------|--------|-----------|------------|
| Local eval (simple) | <10 μs | <5 μs | <20 μs |
| Cache hit | <5 μs | <1 μs | <10 μs |
| Throughput (concurrent) | >100k/s | >1M/s | >50k/s |
| Memory per eval | <1 KB | <500 B | <2 KB |

## Interpreting Results

### Understanding Output

```
BenchmarkLocalEvaluation_Simple-8    500000    3542 ns/op    128 B/op    2 allocs/op
```

- **500000**: Number of iterations
- **3542 ns/op**: 3.542 μs per operation
- **128 B/op**: 128 bytes allocated per operation
- **2 allocs/op**: 2 memory allocations per operation

### Good vs Bad Results

**Good:**
- ✅ Latency: <10 μs/op for local evaluation
- ✅ Allocations: <5 allocs/op
- ✅ Memory: <500 B/op

**Needs Investigation:**
- ⚠️ Latency: >50 μs/op
- ⚠️ Allocations: >20 allocs/op
- ⚠️ Memory: >5 KB/op

## Comparing Results

### Before and After Changes

```bash
# Save baseline
go test -bench=. -benchmem > old.txt

# Make changes...

# Run new benchmarks
go test -bench=. -benchmem > new.txt

# Compare
go install golang.org/x/perf/cmd/benchstat@latest
benchstat old.txt new.txt
```

### Example Output

```
name                          old time/op    new time/op    delta
LocalEvaluation_Simple-8        3.54µs ± 2%    2.89µs ± 1%  -18.36%
LocalEvaluation_Complex-8       12.4µs ± 3%    10.2µs ± 2%  -17.74%

name                          old alloc/op   new alloc/op   delta
LocalEvaluation_Simple-8          128B ± 0%       96B ± 0%  -25.00%
LocalEvaluation_Complex-8         512B ± 0%      384B ± 0%  -25.00%
```

## Continuous Benchmarking

### GitHub Actions Integration

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
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run Benchmarks
        run: |
          cd benchmarks
          go test -bench=. -benchmem -benchtime=3s > results.txt
          cat results.txt

      - name: Upload Results
        uses: actions/upload-artifact@v3
        with:
          name: benchmark-results
          path: benchmarks/results.txt
```

## Performance Regression Testing

### Set Performance Budgets

```bash
# Create performance budget file
cat > performance_budget.yaml <<EOF
benchmarks:
  LocalEvaluation_Simple:
    max_ns_per_op: 10000      # 10 μs
    max_allocs_per_op: 5
    max_bytes_per_op: 500

  LocalEvaluation_Complex:
    max_ns_per_op: 50000      # 50 μs
    max_allocs_per_op: 15
    max_bytes_per_op: 2000
EOF
```

### Automated Regression Detection

```bash
# Script to check for regressions
#!/bin/bash

BUDGET_NS=10000  # 10 μs max for simple eval

RESULT=$(go test -bench=BenchmarkLocalEvaluation_Simple -benchtime=3s | \
  grep "BenchmarkLocalEvaluation_Simple" | \
  awk '{print $3}' | \
  sed 's/ns\/op//')

if [ "$RESULT" -gt "$BUDGET_NS" ]; then
  echo "❌ Performance regression detected!"
  echo "Expected: <${BUDGET_NS}ns/op, Got: ${RESULT}ns/op"
  exit 1
fi

echo "✅ Performance within budget"
```

## Profiling for Optimization

### CPU Profiling

```bash
# Generate profile
go test -bench=BenchmarkLocalEvaluation -cpuprofile=cpu.prof

# Analyze
go tool pprof -http=:8080 cpu.prof

# Or text mode
go tool pprof -top cpu.prof
```

### Memory Profiling

```bash
# Generate profile
go test -bench=BenchmarkLocalEvaluation -memprofile=mem.prof

# Analyze
go tool pprof -http=:8080 mem.prof

# Show allocations
go tool pprof -alloc_space mem.prof
```

### Trace Analysis

```bash
# Generate trace
go test -bench=BenchmarkConcurrent -trace=trace.out

# Visualize
go tool trace trace.out
```

## Best Practices

1. **Consistent Environment**
   - Run on same hardware
   - Close other applications
   - Disable CPU frequency scaling
   - Use benchtime=10s for stability

2. **Multiple Runs**
   - Run at least 3 times
   - Use benchstat for statistical analysis
   - Watch for variance

3. **Realistic Workloads**
   - Test with production-like data
   - Include worst-case scenarios
   - Vary input sizes

4. **Regression Prevention**
   - Benchmark before/after changes
   - Set performance budgets
   - Automate in CI/CD

## Troubleshooting

### High Variance

```bash
# Increase benchmark time
go test -bench=. -benchtime=10s

# Increase iterations
go test -bench=. -count=10

# Check system load
top
```

### Unexpected Results

```bash
# Check for GC interference
go test -bench=. -benchmem -gcflags="-m"

# Profile to find bottlenecks
go test -bench=. -cpuprofile=cpu.prof
go tool pprof -top cpu.prof
```

## Tools and Scripts

### run_benchmarks.sh / .bat
Full comprehensive benchmark suite. Use for official documentation.

**Output**:
- `results/benchmark_TIMESTAMP.txt` - Raw results
- `results/summary_TIMESTAMP.md` - Formatted summary

**Duration**: 20-30 minutes

### quick_bench.sh
Fast abbreviated benchmarks for development.

**Output**: `quick_results.txt`

**Duration**: 1-2 minutes

### benchstat
Statistical analysis tool for comparing results.

```bash
# Install
go install golang.org/x/perf/cmd/benchstat@latest

# Compare two runs
benchstat old.txt new.txt
```

## Detailed Documentation

- **[BENCHMARK_GUIDE.md](BENCHMARK_GUIDE.md)** - Complete execution guide
- **[RESULTS_TEMPLATE.md](RESULTS_TEMPLATE.md)** - How to document results
- **[results/example_results.md](results/example_results.md)** - Example completed results

## Contributing Benchmarks

When adding new benchmarks:

1. Follow naming convention: `Benchmark<Feature>_<Scenario>`
2. Include both simple and complex cases
3. Add to appropriate category
4. Document expected results
5. Update this README
6. Add to BENCHMARK_GUIDE.md if needed

## Quick Reference

```bash
# Run all benchmarks (official)
./run_benchmarks.sh

# Quick test
./quick_bench.sh

# Run specific benchmark
go test -bench=BenchmarkLocalEvaluation -benchmem

# Run with profiling
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Compare results
benchstat old.txt new.txt

# Generate HTML report
go test -bench=. -benchmem > results.txt
benchstat -html results.txt > report.html
```
