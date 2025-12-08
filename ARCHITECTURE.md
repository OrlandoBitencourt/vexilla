# Vexilla Architecture

> Deep dive into the design and decision-making behind Vexilla

## Table of Contents

1. [Overview](#overview)
2. [Architecture Diagram](#architecture-diagram)
3. [Core Components](#core-components)
4. [Evaluation Strategies](#evaluation-strategies)
5. [Data Flow](#data-flow)
6. [Performance Characteristics](#performance-characteristics)
7. [Design Decisions](#design-decisions)

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
│                        ▼                                            │
│              ┌─────────────────────┐                                │
│              │   Vexilla Client    │                                │
│              └─────────────────────┘                                │
└─────────────────────────────────────────────────────────────────────┘
                         │
        ┌────────────────┼────────────────┐
        │                │                │
        ▼                ▼                ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│   Ristretto  │  │  Background  │  │   Optional   │
│    Cache     │  │   Refresh    │  │   Servers    │
│  (in-memory) │  │   Worker     │  │ (admin/hook) │
└──────────────┘  └──────┬───────┘  └──────────────┘
                         │
                         ▼
                 ┌───────────────┐
                 │  Flagr Server │
                 │   (HTTP API)  │
                 └───────────────┘
```

---

## Core Components

### 1. Client (`pkg/vexilla/client.go`)

The main entry point. Orchestrates all other components.

```go
type Client struct {
    memoryStore   *storage.MemoryStore    // Ristretto cache
    diskStore     *storage.DiskStore      // Persistence
    flagrClient   *client.FlagrClient     // HTTP client
    evaluator     *evaluator.Evaluator    // Local evaluation
    strategy      *evaluator.Determiner   // Strategy detection
    breaker       *circuit.Breaker        // Circuit breaker
    // ... telemetry, servers, etc.
}
```

**Responsibilities:**
- Lifecycle management (Start/Stop)
- Coordination between components
- Evaluation routing (local vs remote)
- Admin/webhook interface implementation

### 2. Memory Store (`pkg/storage/memory.go`)

High-performance in-memory cache using [Ristretto](https://github.com/dgraph-io/ristretto).

**Why Ristretto?**
- Concurrent: Lock-free reads
- Smart eviction: TinyLFU admission policy
- High throughput: Millions of ops/sec
- Built-in metrics

### 3. Strategy Determiner (`pkg/evaluator/strategy.go`)

Decides if a flag can be evaluated locally.

```go
func (d *Determiner) CanEvaluateLocally(flag Flag) bool {
    for _, segment := range flag.Segments {
        // Partial rollout? → Needs Flagr
        if segment.RolloutPercent > 0 && segment.RolloutPercent < 100 {
            return false
        }
        
        // Multiple distributions (A/B test)? → Needs Flagr
        if len(segment.Distributions) > 1 {
            return false
        }
        
        // Single distribution < 100%? → Needs Flagr
        if len(segment.Distributions) == 1 && 
           segment.Distributions[0].Percent < 100 {
            return false
        }
    }
    return true // 100% deterministic → local
}
```

### 4. Evaluator (`pkg/evaluator/evaluator.go`)

Evaluates flags locally using [antonmedv/expr](https://github.com/antonmedv/expr).

**Supports:**
- Operators: `EQ`, `NEQ`, `IN`, `NOTIN`, `MATCHES`
- Regex matching
- Multiple constraints (AND logic)
- Attribute-based targeting

### 5. Flagr Client (`pkg/client/flagr.go`)

HTTP client for Flagr API with:
- Request retries
- Timeout handling
- OpenTelemetry tracing
- Error wrapping

### 6. Circuit Breaker (`pkg/circuit/breaker.go`)

Prevents cascade failures when Flagr is down.

**States:**
- `Closed`: Normal operation
- `Open`: Blocking requests (after threshold failures)
- `Half-Open`: Testing recovery

### 7. Telemetry (`pkg/telemetry/`)

OpenTelemetry integration:
- **Traces**: All operations instrumented
- **Metrics**: Cache hits/misses, refresh latency, evaluation counts
- **Attributes**: Rich contextual data

### 8. Servers (`pkg/server/`)

**Webhook Server:**
- Receives `flag.updated` / `flag.deleted` events
- Triggers immediate cache invalidation
- Secret validation

**Admin API:**
- GET `/admin/stats` - Cache statistics
- POST `/admin/invalidate` - Manual invalidation
- POST `/admin/refresh` - Force refresh
- GET `/health` - Health check

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
│  Load flag from cache                │
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

---

## Data Flow

### Initial Load

```
Application Start
    │
    ▼
Client.Start()
    │
    ├─> Load from disk (if enabled)
    │   └─> Warm cache
    │
    ├─> HTTP GET /api/v1/flags
    │   └─> Fetch all flags
    │
    ├─> Store in Ristretto
    │
    ├─> Start background refresh
    │
    ├─> Start webhook server (if enabled)
    │
    └─> Start admin API (if enabled)
```

### Background Refresh

```
Every N minutes (default: 5)
    │
    ▼
Check circuit breaker
    │
    ├─> Open? Skip refresh
    │
    └─> Closed? Continue
        │
        ▼
    HTTP GET /api/v1/flags
        │
        ├─> Success
        │   ├─> Update cache
        │   ├─> Reset circuit breaker
        │   └─> Save to disk
        │
        └─> Failure
            ├─> Increment fail counter
            └─> Open circuit if threshold reached
```

### Webhook Flow

```
Flagr UI: Flag Updated
    │
    ▼
POST /webhook
{
  "event": "flag.updated",
  "flag_keys": ["feature_x"]
}
    │
    ▼
Verify secret
    │
    ▼
Invalidate cache keys
    │
    ▼
Trigger immediate refresh
    │
    ▼
Response 200 OK
```

---

## Performance Characteristics

### Latency Comparison

| Operation | Latency | HTTP Requests |
|-----------|---------|---------------|
| Local evaluation (static flag) | <1ms | 0 |
| Remote evaluation (dynamic flag) | 50-200ms | 1 |
| Cache hit | <1μs | 0 |
| Cache miss | N/A | Varies |

### Throughput

```
Benchmark Results (local machine):

Static Flags (Local Evaluation):
- 10,000 evaluations: ~50ms
- Average: ~5μs per evaluation
- Throughput: ~200,000 eval/sec

Dynamic Flags (Flagr Evaluation):
- 100 evaluations: ~15s
- Average: ~150ms per evaluation
- Throughput: ~6.6 eval/sec

Speedup: ~30,000x faster for static flags!
```

### Memory Usage

- **Per flag**: ~1KB (struct + metadata)
- **1,000 flags**: ~1MB
- **Ristretto overhead**: ~10-20MB
- **Total for 1K flags**: ~30MB

---

## Design Decisions

### Why Ristretto over sync.Map?

| Feature | Ristretto | sync.Map |
|---------|-----------|----------|
| Eviction | ✅ Smart (TinyLFU) | ❌ None |
| Metrics | ✅ Built-in | ❌ Manual |
| Memory bounds | ✅ Configurable | ❌ Unbounded |
| Throughput | ⚡ Higher | 🐢 Lower |

### Why antonmedv/expr?

- Safe sandbox (no code execution)
- Rich expression syntax
- Good performance
- Easy integration

### Why Circuit Breaker?

Prevents:
- Cascade failures when Flagr is down
- Request pile-up
- Resource exhaustion

Allows:
- Graceful degradation
- Automatic recovery
- Operational visibility

### Why Disk Persistence?

Benefits:
- Fast startup (warm cache)
- Survives restarts
- Last-known-good fallback

Trade-offs:
- Disk I/O overhead (async, acceptable)
- Stale data risk (mitigated by TTL)

---

## Scalability

### Horizontal Scaling

```
Each instance maintains its own cache:

Instance 1: [Ristretto Cache] ─┐
Instance 2: [Ristretto Cache] ─┼─> Flagr Server
Instance 3: [Ristretto Cache] ─┘
```

**Pros:**
- No coordination needed
- Linear scalability
- Simple deployment

**Cons:**
- Cache inconsistency during refresh window (acceptable)
- Each instance hits Flagr independently (reduced by caching)

### Vertical Scaling

Ristretto cache can grow to GB+ sizes:
- 10K flags: ~30MB
- 100K flags: ~300MB
- 1M flags: ~3GB

---

## Future Enhancements

1. **Shared cache layer** (Redis) for strict consistency
2. **Batch evaluation** API for multiple flags
3. **Predictive pre-fetching** based on access patterns
4. **Compression** for disk persistence
5. **gRPC support** for lower latency

---

**Built with ❤️ for high-performance feature flagging** 🏴