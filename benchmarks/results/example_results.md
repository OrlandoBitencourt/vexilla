# Benchmark Results - Example

> This is an example of how results should be documented. Replace with actual values from your benchmarks.

## Test Environment

- **Date**: 2024-01-15
- **Hardware**:
  - CPU: Apple M1 Pro (10 cores: 8 performance + 2 efficiency)
  - RAM: 16GB LPDDR5-6400
- **Software**:
  - OS: macOS 14.2 (23C64)
  - Go Version: go1.21.5 darwin/arm64
  - Vexilla Commit: abc123f

## Quick Summary

| Metric | Result | Target | Status |
|--------|--------|--------|--------|
| Local Eval (Simple) | 2.85 μs | <10 μs | ✅ |
| Local Eval (Complex) | 18.46 μs | <50 μs | ✅ |
| Cache Hit | 0.87 μs | <5 μs | ✅ |
| Throughput (Concurrent) | 2,850,000 ops/sec | >100k | ✅ |
| Memory per eval | 128 B | <1 KB | ✅ |

## Detailed Results

### Local Evaluation Performance

| Benchmark | ns/op | μs/op | ops/sec | B/op | allocs/op |
|-----------|-------|-------|---------|------|-----------|
| Simple | 2847 | 2.85 | 351,264 | 128 | 2 |
| WithConstraints | 8234 | 8.23 | 121,482 | 384 | 5 |
| MultipleSegments | 12456 | 12.46 | 80,282 | 512 | 8 |
| ComplexConstraints | 18456 | 18.46 | 54,182 | 896 | 12 |

### Cache Performance

| Benchmark | ns/op | μs/op | ops/sec | B/op | allocs/op |
|-----------|-------|-------|---------|------|-----------|
| CacheHit | 867 | 0.87 | 1,153,403 | 0 | 0 |
| StorageGet | 892 | 0.89 | 1,121,076 | 0 | 0 |
| StorageSet | 3421 | 3.42 | 292,343 | 256 | 3 |

### Concurrent Performance

| Benchmark | ns/op | μs/op | ops/sec | B/op | allocs/op |
|-----------|-------|-------|---------|------|-----------|
| ConcurrentEvaluations | 351 | 0.35 | 2,849,003 | 128 | 2 |

### Deterministic Rollout

| Benchmark | ns/op | μs/op | ops/sec | B/op | allocs/op |
|-----------|-------|-------|---------|------|-----------|
| DeterministicRollout | 5234 | 5.23 | 191,058 | 256 | 4 |

### Large Scale

| Benchmark | ns/op | μs/op | ops/sec | B/op | allocs/op |
|-----------|-------|-------|---------|------|-----------|
| 1000 Flags | 3124 | 3.12 | 320,102 | 128 | 2 |
| 10000 Flags | 3867 | 3.87 | 258,620 | 128 | 2 |

### Client API

| Benchmark | ns/op | μs/op | ops/sec | B/op | allocs/op |
|-----------|-------|-------|---------|------|-----------|
| Bool() | 3542 | 3.54 | 282,326 | 128 | 2 |
| Evaluate() | 4123 | 4.12 | 242,630 | 256 | 4 |

## Analysis

### Performance Highlights

- ✅ **Sub-microsecond cache hits**: Cache lookups complete in <1μs with zero allocations
- ✅ **Excellent concurrency**: 8.1x speedup on 10-core CPU (81% parallel efficiency)
- ✅ **Minimal allocations**: Simple evaluations use only 2 allocations and 128 bytes
- ✅ **Scales well**: 10,000 flags only 24% slower than 1,000 flags

### Comparison to Targets

- **Local Evaluation**: 71% faster than target (2.85μs vs <10μs)
- **Throughput**: 28.5x higher than target (2.85M vs >100k ops/sec)
- **Memory Usage**: 87% below target (128B vs <1KB)

### Scaling Analysis

- **Single → Multi-threaded**: 8.1x speedup (expected: ~8-10 cores utilized)
- **100 → 1000 flags**: 9.6% slower (expected: <50%)
- **1000 → 10000 flags**: 23.8% slower (expected: <100%)

### Memory Efficiency

- **Zero-allocation cache hits**: Pure memory lookups with no GC pressure
- **Consistent per-eval memory**: 128B regardless of cache size
- **Efficient complex evaluations**: 896B even with 5+ constraints

### Deterministic Rollout Performance

- **83% faster than remote**: 5.23μs local vs ~150ms remote (28,700x speedup)
- **Bucket evaluation overhead**: Only 2.38μs vs simple evaluation
- **Production-ready**: Can handle 191k deterministic evaluations per second

## Key Takeaways

1. **Vexilla delivers on performance promises**: All metrics significantly exceed targets
2. **Cache is extremely efficient**: Sub-microsecond lookups with zero allocations
3. **Scales linearly**: Performance degradation well within acceptable bounds at scale
4. **Deterministic rollouts are fast**: Enables 100% local evaluation with minimal overhead
5. **Memory footprint is tiny**: 128 bytes per evaluation, 0 bytes for cache hits

## Recommendations

- [x] **Production ready**: Performance metrics validate production deployment
- [x] **Use deterministic rollouts**: For latency-sensitive applications requiring rollouts
- [x] **Enable filtering**: With 10k flags performing well, filtering will optimize further
- [ ] **Monitor in production**: Set up alerts for >50μs p99 evaluation latency
- [ ] **Profile large deployments**: Test with >100k flags to find upper limits

## Performance vs Remote Evaluation

| Scenario | Vexilla (Local) | Flagr (Remote) | Speedup |
|----------|----------------|----------------|---------|
| Simple flag | 2.85 μs | ~150 ms | 52,632x |
| Complex flag | 18.46 μs | ~200 ms | 10,834x |
| Deterministic rollout | 5.23 μs | ~150 ms | 28,700x |
| 1000 evals/sec | 2.85 ms total | 150 sec total | 52,632x |

## Raw Output

<details>
<summary>Click to expand full benchmark output</summary>

```
goos: darwin
goarch: arm64
pkg: github.com/OrlandoBitencourt/vexilla/benchmarks
BenchmarkLocalEvaluation_Simple-10                       500000              2847 ns/op             128 B/op          2 allocs/op
BenchmarkLocalEvaluation_WithConstraints-10              200000              8234 ns/op             384 B/op          5 allocs/op
BenchmarkLocalEvaluation_MultipleSegments-10             100000             12456 ns/op             512 B/op          8 allocs/op
BenchmarkLocalEvaluation_ComplexConstraints-10            50000             18456 ns/op             896 B/op         12 allocs/op
BenchmarkCacheHit-10                                    2000000               867 ns/op               0 B/op          0 allocs/op
BenchmarkConcurrentEvaluations-10                       5000000               351 ns/op             128 B/op          2 allocs/op
BenchmarkDeterministicRollout-10                         300000              5234 ns/op             256 B/op          4 allocs/op
BenchmarkMultipleFlagEvaluation-10                       400000              3245 ns/op             128 B/op          2 allocs/op
BenchmarkStorageSet-10                                   500000              3421 ns/op             256 B/op          3 allocs/op
BenchmarkStorageGet-10                                  1500000               892 ns/op               0 B/op          0 allocs/op
BenchmarkEvaluatorConstraintMatching-10                  200000              7896 ns/op             384 B/op          5 allocs/op
BenchmarkMemoryAllocation-10                             500000              2847 ns/op             128 B/op          2 allocs/op
BenchmarkLargeScaleCache_1000Flags-10                    400000              3124 ns/op             128 B/op          2 allocs/op
BenchmarkLargeScaleCache_10000Flags-10                   300000              3867 ns/op             128 B/op          2 allocs/op
BenchmarkClientAPI_Bool-10                               400000              3542 ns/op             128 B/op          2 allocs/op
BenchmarkClientAPI_Evaluate-10                           300000              4123 ns/op             256 B/op          4 allocs/op
BenchmarkBooleanVsDetailedEvaluation_Bool-10             400000              3542 ns/op             128 B/op          2 allocs/op
BenchmarkBooleanVsDetailedEvaluation_Evaluate-10         300000              4123 ns/op             256 B/op          4 allocs/op
PASS
ok      github.com/OrlandoBitencourt/vexilla/benchmarks    45.123s
```

</details>

---

**Performance Validated**: All benchmarks meet or exceed targets. Vexilla is production-ready for high-performance deployments.
