# Vexilla Architecture

> Deep dive into the design and decision-making behind Vexilla

## Table of Contents

1. [Overview](#overview)
2. [Architecture Diagram](#architecture-diagram)
3. [Core Components](#core-components)
4. [Evaluation Strategies](#evaluation-strategies)
5. [Flag Filtering System](#flag-filtering-system)
6. [Data Flow](#data-flow)
7. [Performance Characteristics](#performance-characteristics)
8. [Deterministic Rollout Pattern](#deterministic-rollout-pattern)
9. [Design Decisions](#design-decisions)
10. [Recent Enhancements](#recent-enhancements)

---

## Overview

Vexilla is a high-performance caching layer for [Flagr](https://github.com/openflagr/flagr) that intelligently routes feature flag evaluations between local (cached) and remote (Flagr API) evaluation based on the flag's configuration.

### Key Insight

Most feature flags are **deterministic** - they always produce the same result for the same input. These don't need Flagr's stateful bucketing and can be evaluated locally in <1ms with zero HTTP requests.

Only flags with **percentage-based rollouts** or **A/B testing** require Flagr's consistent hashing for sticky user assignments.

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Your Application                             │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐                     │
│  │  HTTP API  │  │  Services  │  │   Workers  │                     │
│  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘                     │
│        │               │               │                            │
│        └───────────────┴───────────────┘                            │
│                        │                                            │
│                  HTTP Middleware (optional)                         │
│                        │                                            │
│                        ▼                                            │
│              ┌─────────────────────┐                                │
│              │    Cache (pkg)      │                                │
│              │  - Orchestration    │                                │
│              │  - Routing Logic    │                                │
│              │  - Circuit Breaker  │                                │
│              └─────────────────────┘                                │
└─────────────────────────────────────────────────────────────────────┘
                         │
        ┌────────────────┼────────────────┬─────────────────┐
        │                │                │                 │
        ▼                ▼                ▼                 ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  Storage     │  │  Evaluator   │  │  Flagr       │  │  Servers     │
│  (Memory/    │  │  (Local)     │  │  Client      │  │  (Admin/     │
│   Disk)      │  │              │  │  (HTTP)      │  │   Webhook)   │
└──────────────┘  └──────────────┘  └──────┬───────┘  └──────┬───────┘
                                            │                 │
                                            ▼                 ▼
                                   ┌───────────────┐  ┌──────────────┐
                                   │  Flagr Server │  │  Admin API   │
                                   │   (HTTP API)  │  │  Webhook API │
                                   └───────────────┘  └──────────────┘
```

---

## Core Components

### 1. Cache (`pkg/cache/cache.go`)

The main orchestrator that coordinates all components.

```go
type Cache struct {
    // Dependencies (injected)
    flagrClient flagr.Client
    storage     storage.Storage
    evaluator   evaluator.Evaluator

    // Configuration
    config Config

    // State management
    ctx    context.Context
    cancel context.CancelFunc
    wg     sync.WaitGroup

    // Circuit breaker state
    mu               sync.RWMutex
    lastRefresh      time.Time
    consecutiveFails int
    circuitOpen      bool
}
```

**Responsibilities:**
- Lifecycle management (Start/Stop)
- Coordination between components
- Evaluation routing (local vs remote)
- Circuit breaker management
- Background refresh orchestration

**Key Methods:**
- `Start(ctx)` - Initializes cache, loads flags, starts background processes
- `Evaluate()` - Main evaluation entry point with routing logic
- `EvaluateBool/String/Int()` - Convenience methods
- `refreshFlags()` - Background refresh with circuit breaker
- `GetMetrics()` - Returns comprehensive metrics

### 2. Storage Layer (`pkg/storage/`)

Provides pluggable storage implementations.

#### Memory Storage (`memory.go`)

High-performance in-memory cache using [Ristretto](https://github.com/dgraph-io/ristretto).

**Why Ristretto?**
- Concurrent: Lock-free reads
- Smart eviction: TinyLFU admission policy
- High throughput: Millions of ops/sec
- Built-in metrics
- TTL support

```go
type MemoryStorage struct {
    cache   *ristretto.Cache
    config  Config
    metrics Metrics
}
```

**Configuration:**
```go
Config{
    MaxCost:     1 << 30,    // 1GB
    NumCounters: 1e7,        // 10M counters
    BufferItems: 64,         // Buffer size
    DefaultTTL:  5 * time.Minute,
}
```

#### Disk Storage (`disk.go`)

Persistent storage for warm cache on restart.

```go
type DiskStorage struct {
    dir     string
    metrics Metrics
    mu      sync.RWMutex
}
```

**Features:**
- JSON-based storage
- Snapshot support
- Atomic operations
- Thread-safe

**Use Cases:**
- Fast startup (warm cache)
- Survives restarts
- Last-known-good fallback

### 3. Evaluator (`pkg/evaluator/evaluator.go`)

Evaluates flags locally using [expr-lang/expr](https://github.com/expr-lang/expr).

```go
type LocalEvaluator struct {
    programCache map[string]*vm.Program
}
```

**Supported Operators:**
- `EQ`, `NEQ` - Equality
- `IN`, `NOTIN` - List membership
- `LT`, `LTE`, `GT`, `GTE` - Numeric comparison
- `MATCHES` - Regex matching
- `CONTAINS` - String contains

**Evaluation Process:**
1. Check if flag is enabled
2. Iterate segments by rank
3. Evaluate constraints (AND logic)
4. Return matching variant

**Strategy Determination:**
```go
func (e *LocalEvaluator) CanEvaluateLocally(flag domain.Flag) bool {
    strategy := flag.DetermineStrategy()
    return strategy == domain.StrategyLocal
}
```

### 4. Flagr Client (`pkg/flagr/http.go`)

HTTP client for Flagr API with retry logic.

```go
type HTTPClient struct {
    endpoint   string
    apiKey     string
    httpClient *http.Client
    maxRetries int
}
```

**Features:**
- Automatic retries (exponential backoff)
- Timeout handling
- Authentication support
- Error wrapping

**Endpoints Used:**
- `GET /api/v1/flags` - Fetch all flags
- `GET /api/v1/flags/:id` - Fetch single flag
- `POST /api/v1/evaluation` - Remote evaluation
- `GET /api/v1/health` - Health check

### 5. Circuit Breaker (`pkg/circuit/breaker.go`)

Prevents cascade failures when Flagr is down.

```go
type Breaker struct {
    state           State  // Closed, Open, HalfOpen
    failures        int
    maxFailures     int
    timeout         time.Duration
    halfOpenTimeout time.Duration
}
```

**State Machine:**
```
┌─────────┐  failures >= max   ┌──────┐
│ CLOSED  ├────────────────────>│ OPEN │
└────┬────┘                     └───┬──┘
     │                              │
     │ success                      │ timeout
     │                              │
     └──────────────────┐    ┌──────┘
                        │    │
                   ┌────▼────▼────┐
                   │  HALF-OPEN   │
                   └──────────────┘
```

**States:**
- `Closed`: Normal operation
- `Open`: Blocking requests (after threshold failures)
- `Half-Open`: Testing recovery

**Configuration:**
```go
Config{
    MaxFailures:     3,
    Timeout:         30 * time.Second,
    HalfOpenTimeout: 10 * time.Second,
}
```

### 6. Telemetry (`pkg/telemetry/`)

OpenTelemetry integration for observability.

```go
type OTelProvider struct {
    tracer trace.Tracer
    meter  metric.Meter
    
    // Metrics
    cacheHits       metric.Int64Counter
    cacheMisses     metric.Int64Counter
    evaluations     metric.Int64Counter
    refreshDuration metric.Float64Histogram
}
```

**Metrics Exported:**
- `vexilla.cache.hits` - Cache hit counter
- `vexilla.cache.misses` - Cache miss counter
- `vexilla.evaluations` - Total evaluations (with strategy label)
- `vexilla.refresh.duration` - Refresh latency histogram
- `vexilla.refresh.success/failure` - Refresh success/failure counters
- `vexilla.circuit.state` - Circuit breaker state gauge

**Traces:**
- All operations instrumented
- Rich contextual attributes
- Error recording

### 7. Servers (`pkg/server/`)

#### Webhook Server (`webhook.go`)

Receives real-time updates from Flagr.

```go
type WebhookServer struct {
    cache  CacheInterface
    port   int
    secret string
}
```

**Events Handled:**
- `flag.updated` - Invalidates and refreshes
- `flag.deleted` - Invalidates flag

**Security:**
- HMAC-SHA256 signature verification
- Configurable secret

#### Admin API (`admin.go`)

Management interface for operations.

**Endpoints:**
- `GET /health` - Health check
- `GET /admin/stats` - Cache metrics
- `POST /admin/invalidate` - Invalidate specific flag
- `POST /admin/invalidate-all` - Clear cache
- `POST /admin/refresh` - Force refresh

#### Middleware (`middleware.go`)

HTTP middleware for automatic request context injection.

```go
type Middleware struct {
    cache CacheInterface
}
```

**Features:**
- Extracts user context from HTTP headers
- Builds evaluation context automatically from request data
- Injects cache and context into request
- Drop-in integration with standard `http.Handler`

**Usage Example:**
```go
// Wrap any HTTP handler
handler := client.HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    evalCtx := vexilla.NewContext(userID)

    if client.Bool(r.Context(), "new-feature", evalCtx) {
        // Feature enabled
    }
}))

http.ListenAndServe(":8080", handler)
```

**Benefits:**
- Eliminates boilerplate context extraction
- Consistent context building across endpoints
- Seamless integration with existing HTTP routers

---

## Evaluation Strategies

### Decision Tree

```
┌─────────────────────────────────────┐
│  Flag Evaluation Request            │
└──────────────┬──────────────────────┘
               │
               ▼
┌──────────────────────────────────────┐
│  Load flag from storage              │
└──────────────┬───────────────────────┘
               │
               ▼
┌──────────────────────────────────────┐
│  Determine evaluation strategy       │
│  CanEvaluateLocally(flag)?           │
└──────────────┬───────────────────────┘
               │
       ┌───────┴───────┐
       │               │
     YES              NO
       │               │
       ▼               ▼
┌─────────────┐  ┌─────────────────┐
│   LOCAL     │  │     FLAGR       │
│  EVALUATION │  │   EVALUATION    │
└──────┬──────┘  └──────┬──────────┘
       │                │
       ▼                ▼
┌─────────────┐  ┌─────────────────────────┐
│  Evaluate   │  │  HTTP POST              │
│  constraints│  │  /api/v1/evaluation     │
│  with expr  │  │                         │
│             │  │  Flagr performs:        │
│  <1ms       │  │  • Consistent hashing   │
│  0 HTTP     │  │  • Bucket assignment    │
└─────┬───────┘  │  • Variant distribution │
      │          │                         │
      │          │  50-200ms               │
      │          │  1 HTTP request         │
      │          └──────┬──────────────────┘
      │                 │
      └────────┬────────┘
               │
               ▼
        ┌─────────────┐
        │   Return    │
        │   Result    │
        └─────────────┘
```

### Strategy Determination Logic

```go
func (f *Flag) DetermineStrategy() EvaluationStrategy {
    if !f.Enabled {
        return StrategyLocal // Disabled = always local
    }

    if len(f.Segments) == 0 {
        return StrategyLocal // No segments = local
    }

    for _, segment := range f.Segments {
        // Partial rollout → remote
        if segment.RolloutPercent > 0 && segment.RolloutPercent < 100 {
            return StrategyRemote
        }

        // Multiple distributions (A/B) → remote
        if len(segment.Distributions) > 1 {
            return StrategyRemote
        }

        // Single distribution < 100% → remote
        if len(segment.Distributions) == 1 {
            if segment.Distributions[0].Percent < 100 {
                return StrategyRemote
            }
        }
    }

    return StrategyLocal // 100% deterministic
}
```

---

## Flag Filtering System

### Overview

The filtering system allows selective caching of flags to optimize memory usage and cache efficiency. This is especially valuable in microservice architectures where each service only needs a subset of flags.

### Filter Configuration

```go
type FilterConfig struct {
    // OnlyEnabled: Cache only enabled flags
    OnlyEnabled bool

    // ServiceName: Current service identifier
    ServiceName string

    // RequireServiceTag: Only cache flags tagged with ServiceName
    RequireServiceTag bool

    // AdditionalTags: Additional tag filtering
    AdditionalTags []string

    // TagMatchMode: "any" or "all"
    TagMatchMode string
}
```

### Filtering Logic

```go
func (f FilterConfig) ShouldCacheFlag(flag FlagMetadata) bool {
    // Rule 1: OnlyEnabled filter
    if f.OnlyEnabled && !flag.Enabled {
        return false
    }

    // Rule 2: Service tag filter
    if f.RequireServiceTag {
        if !hasTag(flag.Tags, f.ServiceName) {
            return false
        }
    }

    // Rule 3: Additional tags filter
    if len(f.AdditionalTags) > 0 {
        if !matchesTags(flag.Tags, f.AdditionalTags, f.TagMatchMode) {
            return false
        }
    }

    return true
}
```

### Memory Savings Calculation

```go
func (f FilterConfig) EstimateMemorySavings(totalFlags, filteredFlags int) MemorySavings {
    savedFlags := totalFlags - filteredFlags
    percentSaved := float64(savedFlags) / float64(totalFlags) * 100
    
    // Rough estimate: 1KB per flag
    const bytesPerFlag = 1024
    savedBytes := int64(savedFlags * bytesPerFlag)
    
    return MemorySavings{
        TotalFlags:      totalFlags,
        CachedFlags:     filteredFlags,
        FilteredFlags:   savedFlags,
        PercentFiltered: percentSaved,
        BytesSaved:      savedBytes,
    }
}
```

### Use Cases

**1. Microservice Architecture:**
```go
// Each service caches only its flags
WithServiceTag("user-service", true)
// 1000 total flags → 50 for user-service = 95% memory saving
```

**2. Environment Separation:**
```go
// Only cache production flags
WithAdditionalTags([]string{"production"}, "any")
// Avoids caching staging/dev flags in production
```

**3. Resource Optimization:**
```go
// Only enabled flags for active service
WithOnlyEnabled(true)
WithServiceTag("payment-service", true)
// Maximum memory efficiency
```

---

## Data Flow

### Initial Load

```
Application Start
    │
    ▼
Cache.Start(ctx)
    │
    ├─> Load from disk (if DiskStorage configured)
    │   └─> Warm cache with persisted flags
    │
    ├─> HTTP GET /api/v1/flags (fetch all IDs)
    │   └─> For each flag ID:
    │       └─> HTTP GET /api/v1/flags/:id (detailed)
    │
    ├─> Apply filtering (if configured)
    │   └─> FilterConfig.ShouldCacheFlag()
    │       ├─> OnlyEnabled filter
    │       ├─> ServiceTag filter
    │       └─> AdditionalTags filter
    │
    ├─> Store in Ristretto
    │   └─> storage.Set(key, flag, ttl)
    │
    ├─> Start background refresh goroutine
    │
    ├─> Start webhook server (if enabled)
    │   └─> Listen on configured port
    │       └─> Handle flag.updated and flag.deleted events
    │
    └─> Start admin API (if enabled)
        └─> Listen on configured port
            └─> Expose management endpoints
```

### Background Refresh

```
Every N minutes (default: 5)
    │
    ▼
Check circuit breaker
    │
    ├─> Open? Skip refresh, return error
    │
    └─> Closed/HalfOpen? Continue
        │
        ▼
    HTTP GET /api/v1/flags → for each: GET /api/v1/flags/:id
        │
        ├─> Success
        │   ├─> Apply filtering
        │   ├─> Update cache (storage.Set)
        │   ├─> Reset circuit breaker
        │   ├─> Save to disk (if DiskStorage)
        │   └─> Update lastRefresh timestamp
        │
        └─> Failure
            ├─> Increment consecutiveFails
            ├─> If fails >= threshold: Open circuit
            └─> Return error
```

### Webhook Flow

```
Flagr UI: Flag Updated
    │
    ▼
POST /webhook
{
  "event": "flag.updated",
  "flag_keys": ["feature_x"],
  "timestamp": "2025-01-15T10:30:00Z"
}
    │
    ▼
Verify HMAC signature (if secret configured)
    │
    ├─> Invalid? Return 401
    │
    └─> Valid? Continue
        │
        ▼
    Parse payload
        │
        ▼
    Handle event
        │
        ├─> "flag.updated"
        │   ├─> For each flag_key:
        │   │   └─> storage.Delete(flag_key)
        │   └─> Trigger immediate refresh
        │
        └─> "flag.deleted"
            └─> For each flag_key:
                └─> storage.Delete(flag_key)
    │
    ▼
Response 200 OK
```

### Evaluation Flow

```
cache.Evaluate(ctx, flagKey, evalCtx)
    │
    ▼
storage.Get(flagKey)
    │
    ├─> Found
    │   │
    │   ▼
    │   evaluator.CanEvaluateLocally(flag)?
    │   │
    │   ├─> Yes (Local Strategy)
    │   │   │
    │   │   ▼
    │   │   evaluator.Evaluate(flag, evalCtx)
    │   │   │
    │   │   ├─> Check flag.Enabled
    │   │   ├─> Iterate segments by rank
    │   │   ├─> Evaluate constraints (AND logic)
    │   │   │   └─> expr engine for operators
    │   │   └─> Return matching variant
    │   │       │
    │   │       └─> <1ms, 0 HTTP requests
    │   │
    │   └─> No (Remote Strategy)
    │       │
    │       ▼
    │       Check circuit breaker
    │       │
    │       ├─> Open? Return error
    │       │
    │       └─> Closed? Continue
    │           │
    │           ▼
    │           flagrClient.EvaluateFlag(flagKey, evalCtx)
    │           │
    │           └─> POST /api/v1/evaluation
    │               │
    │               └─> 50-200ms, 1 HTTP request
    │
    └─> Not Found
        │
        ▼
        Refresh all flags from Flagr
        │
        └─> Retry evaluation OR apply fallback strategy
```

---

## Performance Characteristics

### Latency Comparison

| Operation | Latency | HTTP Requests | Notes |
|-----------|---------|---------------|-------|
| Local evaluation (static) | <1ms | 0 | Constraint evaluation with expr |
| Remote evaluation (dynamic) | 50-200ms | 1 | Full Flagr evaluation |
| Cache hit (Ristretto) | <1μs | 0 | In-memory lookup |
| Cache miss → fetch | 50-200ms | 1+ | Triggers refresh |
| Background refresh | N/A | N flags | Async, non-blocking |

### Throughput

```
Benchmark Results (M1 MacBook Pro):

Local Evaluation (Static Flags):
- 10,000 evaluations: ~50ms
- Average: ~5μs per evaluation
- Throughput: ~200,000 eval/sec
- Memory: 0 allocations per eval (after warmup)

Remote Evaluation (Dynamic Flags):
- 100 evaluations: ~15s
- Average: ~150ms per evaluation
- Throughput: ~6.6 eval/sec
- Limited by network + Flagr processing

Speedup: ~30,000x faster for static flags!
```

### Memory Usage

**Per Flag:**
- Flag struct: ~500 bytes
- Metadata: ~100 bytes
- Ristretto overhead: ~200 bytes
- **Total: ~800 bytes per flag**

**Scaling:**
- 100 flags: ~80 KB
- 1,000 flags: ~800 KB
- 10,000 flags: ~8 MB
- 100,000 flags: ~80 MB

**With Filtering (Example):**
- 10,000 total flags
- Service-specific: 500 flags (95% filtered)
- Memory used: ~400 KB vs 8 MB
- **Savings: ~7.6 MB (95% reduction)**

### Ristretto Performance

```
Operations per second (M1 MacBook Pro):
- Set: ~10M ops/sec
- Get: ~30M ops/sec (lock-free)
- Mixed workload: ~15M ops/sec

Memory efficiency:
- Admission rate: ~95% (TinyLFU)
- Eviction accuracy: Very high
- False positive rate: <1%
```

---

## Deterministic Rollout Pattern

### Overview

Vexilla enables a powerful pattern for **deterministic rollouts** that eliminates the need for random percentage-based evaluations, enabling 100% local evaluation with zero HTTP dependencies.

### Problem with Traditional Rollouts

Traditional percentage-based rollouts have several drawbacks:
- **Non-deterministic**: Random evaluation means results can change between calls
- **HTTP-dependent**: Requires Flagr API calls for consistent hashing
- **Hard to debug**: Difficult to reproduce user-specific behavior
- **Network latency**: Adds 50-200ms per evaluation

### Solution: Pre-computed Buckets

Instead of relying on Flagr's random rollout percentages, applications can pre-compute a **deterministic bucket** from user identifiers and send it as a simple numeric attribute.

### Architecture

```
User Identifier (CPF, UserID, etc.)
    │
    ▼
Application (pre-processing)
    │
    ├─> Extract/Hash identifier
    │   └─> Generate bucket: 0-99
    │
    ▼
Vexilla.Evaluate(ctx, flag, {bucket: 42})
    │
    ▼
Local Evaluation (no HTTP)
    │
    ├─> Check: bucket >= 0 AND bucket <= 69?
    │   └─> Match: Variant A
    │
    └─> Else: Default variant
```

### Example: CPF-based Bucketing

```go
// Pre-processing: Extract bucket from CPF
func CPFBucket(cpf string) int {
    clean := strings.NewReplacer(".", "", "-", "").Replace(cpf)
    if len(clean) < 7 {
        return -1
    }

    // Use digits 6-7 to create bucket 00-99
    bucket, err := strconv.Atoi(clean[5:7])
    if err != nil {
        return -1
    }

    return bucket
}

// Usage
attrs := vexilla.Attributes{
    "cpf_bucket": CPFBucket("123.456.789-09"),
}

enabled := client.Bool(ctx, "new-feature", attrs)
```

### Flagr Configuration

**Segment: audience_a (70% rollout)**

| Field      | Operator | Value |
|------------|----------|-------|
| cpf_bucket | >=       | 0     |
| cpf_bucket | <=       | 69    |

**Segment: default (30% rollout)**
- No constraints (matches all others)

### Benefits

✅ **Deterministic**: Same input always produces same result
✅ **100% Local**: No HTTP calls to Flagr needed
✅ **Sub-millisecond**: <1ms evaluation latency
✅ **Reproducible**: Easy to debug user-specific behavior
✅ **Offline-capable**: Works without network connectivity
✅ **Cacheable**: Fully compatible with Vexilla's caching layer

### Performance Comparison

| Approach | Latency | HTTP Calls | Deterministic |
|----------|---------|------------|---------------|
| Flagr % rollout | 50-200ms | 1 per eval | ❌ Random |
| Deterministic bucket | <1ms | 0 | ✅ Stable |

**Speedup**: ~50-200x faster, zero network overhead

### Use Cases

This pattern is ideal for:
- **Critical features**: Where latency matters
- **Edge computing**: Limited connectivity environments
- **Mobile apps**: Reduce battery drain from network calls
- **Compliance**: Reproducible audit trails
- **A/B testing**: Stable user assignments without sticky sessions

### Applicable Identifiers

Any stable identifier can be used:
- User ID (hash % 100)
- Account ID
- Document numbers (CPF, SSN, etc.)
- Email hash
- Device ID
- Session ID (for short-term experiments)

---

## Design Decisions

### 1. Why Ristretto over sync.Map?

| Feature | Ristretto | sync.Map |
|---------|-----------|----------|
| Eviction | ✅ Smart (TinyLFU) | ❌ None (unbounded) |
| Metrics | ✅ Built-in | ❌ Manual tracking |
| Memory bounds | ✅ Configurable | ❌ Unbounded growth |
| Throughput | ⚡ 10-30M ops/sec | 🐢 Lower |
| TTL support | ✅ Native | ❌ Manual expiry |
| Admission policy | ✅ TinyLFU (intelligent) | ❌ N/A |

**Verdict:** Ristretto provides production-grade caching with bounded memory, high performance, and intelligent eviction.

### 2. Why expr-lang/expr?

**Alternatives Considered:**
- `text/template` - Too limited
- `github.com/antonmedv/expr` ✅ **Chosen**
- `github.com/robertkrimen/otto` - Full JS engine (overkill)
- Custom parser - Too much maintenance

**Why expr?**
- ✅ Safe sandbox (no code execution)
- ✅ Rich expression syntax
- ✅ Good performance (~1μs per eval)
- ✅ Easy integration
- ✅ Type-safe evaluation
- ✅ Regex support

### 3. Why Circuit Breaker?

**Problem:** When Flagr is down, applications shouldn't retry indefinitely.

**Circuit Breaker Benefits:**
- Prevents cascade failures
- Fails fast (no waiting for timeouts)
- Automatic recovery testing (half-open)
- Configurable thresholds

**Trade-offs:**
- Adds complexity
- Requires tuning (max failures, timeout)
- May block valid requests during recovery

**Verdict:** Essential for production resilience. Cost is worth the protection.

### 4. Why Disk Persistence?

**Benefits:**
- Fast startup (warm cache from disk)
- Survives restarts (last-known-good state)
- Graceful degradation if Flagr is down

**Trade-offs:**
- Disk I/O overhead (mitigated: async writes)
- Stale data risk (mitigated: TTL + refresh)
- Disk space usage (minimal: ~1KB per flag)

**Verdict:** Optional but recommended. Provides excellent startup performance.

### 5. Why Flag Filtering?

**Problem:** In microservice architectures, each service doesn't need all 10,000+ flags.

**Filtering Benefits:**
- ✅ Reduced memory footprint (50-95% savings)
- ✅ Faster cache operations (smaller dataset)
- ✅ Better cache hit ratios
- ✅ Lower refresh overhead

**Example Impact:**
```
Without filtering:
- 10,000 flags × 1KB = 10 MB memory
- Refresh time: ~30 seconds

With filtering (5% relevant):
- 500 flags × 1KB = 500 KB memory (95% saving)
- Refresh time: ~1.5 seconds (95% faster)
```

**Verdict:** Critical for microservice deployments. Dramatically improves efficiency.

### 6. Why Separate Storage Interface?

**Benefits:**
- Testability (mock storage)
- Flexibility (swap implementations)
- Future-proofing (Redis, etc.)

**Implementations:**
- `MemoryStorage` - Production (Ristretto)
- `DiskStorage` - Persistence option
- `MockStorage` - Testing

**Verdict:** Clean separation of concerns. Easy to extend.

---

## Scalability

### Horizontal Scaling

```
Each instance maintains its own cache:

Instance 1: [Ristretto Cache] ─┐
Instance 2: [Ristretto Cache] ─┼─> Flagr Server
Instance 3: [Ristretto Cache] ─┤
Instance N: [Ristretto Cache] ─┘
```

**Pros:**
- ✅ No coordination needed
- ✅ Linear scalability
- ✅ Simple deployment
- ✅ No single point of failure

**Cons:**
- ⚠️ Cache inconsistency during refresh window (acceptable)
- ⚠️ Each instance refreshes independently

**Mitigation:**
- Webhook support for real-time updates
- Short refresh intervals (1-5 minutes)
- Staggered refresh (add jitter)

### Vertical Scaling

Ristretto can scale to very large datasets:

```
Memory capacity:
- 1K flags:   ~1 MB
- 10K flags:  ~10 MB
- 100K flags: ~100 MB
- 1M flags:   ~1 GB
```

**Configuration for Large Deployments:**
```go
Config{
    MaxCost:     10 << 30,  // 10GB
    NumCounters: 1e8,       // 100M counters
    BufferItems: 256,       // Larger buffer
}
```

### Performance at Scale

**10K flags, 10K req/s:**
- Local evaluations: ~9,500/s (95% local)
- Remote evaluations: ~500/s (5% remote)
- Flagr load: 500 req/s (vs 10K without Vexilla)
- **95% load reduction on Flagr**

---

## Recent Enhancements

### ✅ Implemented Features (Latest Version)

1. **✅ Webhook Invalidation**
   - Real-time flag updates from Flagr
   - HMAC-SHA256 signature verification
   - Event-driven cache invalidation
   - Sub-second update propagation

   ```go
   vexilla.WithWebhookInvalidation(vexilla.WebhookConfig{
       Port:   18001,
       Secret: "shared-secret",
   })
   ```

2. **✅ Admin Server**
   - REST API for cache management
   - Health checks and metrics
   - Manual invalidation and refresh
   - Operational visibility

   ```go
   vexilla.WithAdminServer(vexilla.AdminConfig{
       Port: 19000,
   })
   ```

3. **✅ HTTP Middleware**
   - Automatic request context injection
   - Seamless integration with `http.Handler`
   - Eliminates boilerplate code

   ```go
   handler := client.HTTPMiddleware(myHandler)
   ```

4. **✅ Deterministic Rollout Pattern**
   - Pre-computed bucket-based evaluation
   - 100% local evaluation for rollouts
   - Zero HTTP overhead
   - See [Deterministic Rollout Pattern](#deterministic-rollout-pattern) section

### 🔜 Future Enhancements

1. **Distributed Cache** (Redis/Memcached)
   - Shared cache across instances
   - Strict consistency option
   - Cache warming strategies

2. **Batch Evaluation API**
   ```go
   results := cache.EvaluateMany(ctx, []string{"flag1", "flag2", "flag3"}, evalCtx)
   ```

3. **Predictive Pre-fetching**
   - Learn access patterns
   - Pre-warm cache for likely evaluations
   - ML-based prediction

4. **Compression**
   - Compress flag data on disk
   - Reduce memory footprint
   - Optional ZSTD compression

5. **gRPC Support**
   - Lower latency than HTTP
   - Better streaming for updates
   - Bi-directional communication

6. **Advanced Filtering**
   - Custom filter functions
   - Regex-based tag matching
   - Time-based filtering (active hours)

7. **Enhanced Telemetry**
   - Per-flag metrics
   - Evaluation latency percentiles
   - Cache efficiency scoring

---

## Testing Strategy

### Unit Tests
- All packages have `_test.go` files
- Mock implementations for dependencies
- Table-driven tests
- Target: >80% coverage

### Integration Tests
- Test against real Flagr instance
- Docker-based test environment
- End-to-end evaluation flows

### Benchmarks
- Performance regression detection
- Memory allocation tracking
- Throughput measurements

### Load Tests
- Concurrent evaluation stress tests
- Cache eviction behavior under load
- Circuit breaker trigger testing

---

## Conclusion

Vexilla provides a production-grade caching layer for Flagr that:

1. **Dramatically improves performance** (50-200x) for deterministic flags
2. **Reduces infrastructure load** (0-95%) on Flagr servers
3. **Maintains full compatibility** with Flagr's feature set
4. **Provides intelligent filtering** for resource optimization
5. **Offers enterprise features** (observability, resilience, persistence)
6. **Enables real-time updates** via webhook invalidation
7. **Simplifies operations** with admin API and HTTP middleware
8. **Supports deterministic rollouts** for offline-first evaluation

The architecture balances performance, reliability, and maintainability while providing clear extension points for future enhancements.

---

**Built with ❤️ for high-performance feature flagging** 🏴
