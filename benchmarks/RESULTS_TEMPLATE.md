# Benchmark Results Template

Use this template to document benchmark results after running tests.

## Test Environment

- **Date**: YYYY-MM-DD
- **Hardware**:
  - CPU: [Model, cores, frequency]
  - RAM: [Size, type, speed]
- **Software**:
  - OS: [OS name and version]
  - Go Version: [from `go version`]
  - Vexilla Commit: [git commit SHA]

## Quick Summary

| Metric | Result | Target | Status |
|--------|--------|--------|--------|
| Local Eval (Simple) | X.XX μs | <10 μs | ✅/❌ |
| Local Eval (Complex) | XX.XX μs | <50 μs | ✅/❌ |
| Cache Hit | X.XX μs | <5 μs | ✅/❌ |
| Throughput (Concurrent) | XXX,XXX ops/sec | >100k | ✅/❌ |
| Memory per eval | XXX B | <1 KB | ✅/❌ |

## Detailed Results

### Local Evaluation Performance

| Benchmark | ns/op | μs/op | ops/sec | B/op | allocs/op |
|-----------|-------|-------|---------|------|-----------|
| Simple | | | | | |
| WithConstraints | | | | | |
| MultipleSegments | | | | | |
| ComplexConstraints | | | | | |

### Cache Performance

| Benchmark | ns/op | μs/op | ops/sec | B/op | allocs/op |
|-----------|-------|-------|---------|------|-----------|
| CacheHit | | | | | |
| StorageGet | | | | | |
| StorageSet | | | | | |

### Concurrent Performance

| Benchmark | ns/op | μs/op | ops/sec | B/op | allocs/op |
|-----------|-------|-------|---------|------|-----------|
| ConcurrentEvaluations | | | | | |

### Deterministic Rollout

| Benchmark | ns/op | μs/op | ops/sec | B/op | allocs/op |
|-----------|-------|-------|---------|------|-----------|
| DeterministicRollout | | | | | |

### Large Scale

| Benchmark | ns/op | μs/op | ops/sec | B/op | allocs/op |
|-----------|-------|-------|---------|------|-----------|
| 1000 Flags | | | | | |
| 10000 Flags | | | | | |

### Client API

| Benchmark | ns/op | μs/op | ops/sec | B/op | allocs/op |
|-----------|-------|-------|---------|------|-----------|
| Bool() | | | | | |
| Evaluate() | | | | | |

## Analysis

### Performance Highlights

- ✅ **[Highlight 1]**: [Description]
- ✅ **[Highlight 2]**: [Description]
- ⚠️ **[Concern 1]**: [Description]

### Comparison to Targets

- **Local Evaluation**: [% vs target]
- **Throughput**: [% vs target]
- **Memory Usage**: [% vs target]

### Scaling Analysis

- **Single → Multi-threaded**: XXx speedup (expected: ~cores)
- **100 → 1000 flags**: XX% slower (expected: <50%)
- **1000 → 10000 flags**: XX% slower (expected: <100%)

## Key Takeaways

1. [Takeaway 1]
2. [Takeaway 2]
3. [Takeaway 3]

## Recommendations

- [ ] [Recommendation 1]
- [ ] [Recommendation 2]
- [ ] [Recommendation 3]

## Raw Output

<details>
<summary>Click to expand full benchmark output</summary>

\`\`\`
[Paste full go test output here]
\`\`\`

</details>

---

## How to Use This Template

### 1. Run Benchmarks

```bash
cd benchmarks
./run_benchmarks.sh  # or run_benchmarks.bat on Windows
```

### 2. Fill in Environment Section

Copy from your system:

```bash
# Get CPU info
# Linux: lscpu | grep "Model name"
# macOS: sysctl -n machdep.cpu.brand_string
# Windows: wmic cpu get name

# Get RAM info
# Linux: free -h
# macOS: system_profiler SPHardwareDataType | grep Memory
# Windows: wmic memorychip get capacity

# Get Go version
go version

# Get commit SHA
git rev-parse --short HEAD
```

### 3. Extract Metrics

From benchmark output, calculate:

```python
# Convert ns/op to μs/op
microseconds = nanoseconds / 1000

# Calculate ops/sec
ops_per_sec = 1_000_000_000 / nanoseconds
```

Example:
```
Input:  3542 ns/op
μs/op:  3542 / 1000 = 3.542 μs
ops/s:  1,000,000,000 / 3542 = 282,326 ops/sec
```

### 4. Fill in Tables

For each benchmark result like:
```
BenchmarkLocalEvaluation_Simple-8    500000    3542 ns/op    128 B/op    2 allocs/op
```

Fill in the table:
| Benchmark | ns/op | μs/op | ops/sec | B/op | allocs/op |
|-----------|-------|-------|---------|------|-----------|
| Simple | 3542 | 3.542 | 282,326 | 128 | 2 |

### 5. Analyze Results

Compare to targets:
- ✅ Green: Within target
- ⚠️ Yellow: Acceptable but not ideal
- ❌ Red: Exceeds target, needs investigation

### 6. Write Analysis

Focus on:
- What's performing well
- What needs attention
- How it scales
- Memory efficiency

### 7. Save and Commit

```bash
# Save to results directory
cp RESULTS_TEMPLATE.md results/results_YYYYMMDD.md

# Edit the file with your results

# Commit
git add results/
git commit -m "Add benchmark results for YYYY-MM-DD"
```

## Example Filled Template

See `results/example_results.md` for a complete example of a filled-out template.
