# 🏴 Vexilla

![Vexilla Logo](https://raw.githubusercontent.com/OrlandoBitencourt/vexilla/refs/heads/main/media/logo.jpg)

> **High-performance caching layer for [Flagr](https://github.com/openflagr/flagr)**  
> Intelligent local/remote evaluation routing with smart flag filtering

[![Go Reference](https://pkg.go.dev/badge/github.com/OrlandoBitencourt/vexilla.svg)](https://pkg.go.dev/github.com/OrlandoBitencourt/vexilla)
[![Go Report Card](https://goreportcard.com/badge/github.com/OrlandoBitencourt/vexilla)](https://goreportcard.com/report/github.com/OrlandoBitencourt/vexilla)
[![codecov](https://codecov.io/gh/OrlandoBitencourt/vexilla/branch/main/graph/badge.svg)](https://codecov.io/gh/OrlandoBitencourt/vexilla)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## 🎯 What is Vexilla?

Vexilla is a **high-performance caching layer** for [Flagr](https://github.com/openflagr/flagr) that intelligently routes feature flag evaluations between local (cached) and remote (Flagr API) evaluation based on the flag's configuration.

### 📊 Performance Impact

| Metric | Flagr Direct | Vexilla (Local) | Vexilla (Remote) | Improvement |
|--------|--------------|-----------------|------------------|-------------|
| **Latency** | 50-200ms | <1ms | 50-200ms | **50-200x faster** |
| **HTTP Requests** | 1 per eval | 0 per eval | 1 per eval | **100% reduction** |
| **Throughput** | ~2K req/s | >200K req/s | ~2K req/s | **100x higher** |
| **Memory Usage** | N/A | ~1KB/flag | ~1KB/flag | Configurable filtering |

---

## 🤔 Why Vexilla?

### The Problem with Flagr

[Flagr](https://github.com/openflagr/flagr) is excellent for feature flagging with:
- ✅ Powerful A/B testing with consistent bucketing
- ✅ Dynamic configuration management
- ✅ Clean REST API and UI
- ✅ Segment-based targeting

**However**, every flag evaluation requires an HTTP request (50-200ms latency).

**At scale:**
- 10,000 requests/second = 10,000 HTTP calls to Flagr
- Flagr becomes a bottleneck
- Infrastructure costs increase
- Latency accumulates

### The Vexilla Solution

Vexilla solves this by **intelligently caching** flag definitions and **routing evaluation**:

**Key Insight:** Most flags are **deterministic** (e.g., `country == "BR"`). These don't need Flagr's stateful bucketing - they can be evaluated locally in <1ms with zero HTTP requests.

Only flags with **percentage-based rollouts** or **A/B testing** require Flagr's consistent hashing.

```
┌─────────────────────────────────────────┐
│  Your Application (10K req/s)           │
└──────────────┬──────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────┐
│           Vexilla Cache                  │
│  • In-memory cache (Ristretto)          │
│  • Smart local/remote routing            │
│  • <1ms for deterministic flags          │
│  • Optional flag filtering               │
└──────────────┬───────────────────────────┘
               │
        Only when needed
               │
               ▼
┌──────────────────────────────────────────┐
│      Flagr Server                        │
│  • A/B tests (consistent hash)           │
│  • Percentage rollouts                   │
│  • Complex variant assignments           │
└──────────────────────────────────────────┘
```

---

## ✨ Features

### 🚀 Core Features
- **Sub-millisecond evaluation** for deterministic flags (100% rollout)
- **Smart routing** - Automatically detects when Flagr is needed
- **Background refresh** - Keeps flags fresh without blocking
- **Circuit breaker** - Resilient to Flagr outages
- **Dual storage** - Ristretto (memory) + optional disk persistence

### 🎯 Flag Filtering (Resource Optimization)
- **OnlyEnabled filter** - Cache only enabled flags
- **Service-based filtering** - Cache only flags tagged for your service
- **Tag-based filtering** - Filter by environment (production, staging)
- **Memory savings** - Reduce memory footprint in microservice environments

### 🔔 Real-time Updates
- **Webhook support** - Instant flag updates from Flagr
- **Event-driven** - No polling overhead
- **Signature verification** - Secure webhook validation

### 🛠️ Operations
- **Admin API** - Management endpoints for ops teams
- **Health checks** - Monitor cache status
- **HTTP middleware** - Drop-in request context injection

### 📊 Observability
- **Full OpenTelemetry** - Traces and metrics
- **Cache statistics** - Hit ratios, evictions, etc.
- **Evaluation tracking** - Local vs remote routing

---

## 🚀 Quick Start

### Installation

```bash
go get github.com/OrlandoBitencourt/vexilla
```

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/OrlandoBitencourt/vexilla/pkg/cache"
    "github.com/OrlandoBitencourt/vexilla/pkg/domain"
    "github.com/OrlandoBitencourt/vexilla/pkg/evaluator"
    "github.com/OrlandoBitencourt/vexilla/pkg/flagr"
    "github.com/OrlandoBitencourt/vexilla/pkg/storage"
)

func main() {
    // 1. Create dependencies
    flagrClient := flagr.NewHTTPClient(flagr.Config{
        Endpoint:   "http://localhost:18000",
        Timeout:    5 * time.Second,
        MaxRetries: 3,
    })

    memStorage, _ := storage.NewMemoryStorage(storage.DefaultConfig())
    eval := evaluator.New()

    // 2. Create cache with options
    c, err := cache.New(
        cache.WithFlagrClient(flagrClient),
        cache.WithStorage(memStorage),
        cache.WithEvaluator(eval),
        cache.WithRefreshInterval(5 * time.Minute),
        cache.WithOnlyEnabled(true), // 🔥 Filter disabled flags
    )
    if err != nil {
        log.Fatal(err)
    }

    // 3. Start cache
    ctx := context.Background()
    if err := c.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer c.Stop()

    // 4. Evaluate flags
    evalCtx := domain.EvaluationContext{
        EntityID: "user_123",
        Context: map[string]interface{}{
            "country": "BR",
            "tier":    "premium",
        },
    }

    // Boolean evaluation
    enabled := c.EvaluateBool(ctx, "my-feature", evalCtx)
    log.Printf("Feature enabled: %v", enabled)

    // String evaluation
    theme := c.EvaluateString(ctx, "ui-theme", evalCtx, "light")
    log.Printf("Theme: %s", theme)

    // Full evaluation with details
    result, _ := c.Evaluate(ctx, "my-feature", evalCtx)
    log.Printf("Result: %+v", result)
}
```

### Simple Configuration (One-liner)

```go
c, err := cache.NewSimple(cache.SimpleConfig{
    FlagrEndpoint:   "http://localhost:18000",
    RefreshInterval: 5 * time.Minute,
    OnlyEnabled:     true,
    ServiceName:     "user-service", // Filter by service tag
})
```

---

## 🎓 Understanding Evaluation Strategies

Vexilla automatically determines the best evaluation strategy:

### ✅ Local Evaluation (Static Flags)

**Conditions:**
- `rollout_percent: 100` (everyone gets evaluated)
- Single distribution with `percentage: 100`
- Deterministic constraints only

**Example:**
```json
{
  "key": "premium_users",
  "segments": [{
    "rollout_percent": 100,
    "constraints": [
      {"property": "tier", "operator": "EQ", "value": "premium"}
    ],
    "distributions": [{"variant_key": "enabled", "percentage": 100}]
  }]
}
```

✅ **Evaluated locally** - Same inputs always produce same outputs

### ❌ Flagr Evaluation (Dynamic Flags)

**Requires Flagr when:**
- `rollout_percent < 100` (partial rollout)
- Multiple distributions (A/B test)
- Single distribution with `percentage < 100`

**Example:**
```json
{
  "key": "gradual_rollout",
  "segments": [{
    "rollout_percent": 30,  // ❌ Only 30% of users
    "distributions": [{"variant_key": "enabled", "percentage": 100}]
  }]
}
```

❌ **Requires Flagr** - Needs consistent hashing for sticky behavior

---

## 🔧 Configuration

### Full Configuration

```go
config := cache.Config{
    // Refresh behavior
    RefreshInterval: 5 * time.Minute,
    InitialTimeout:  10 * time.Second,
    
    // Fallback strategy
    FallbackStrategy: "fail_closed", // or "fail_open", "error"
    
    // Circuit breaker
    CircuitBreakerThreshold: 3,
    CircuitBreakerTimeout:   30 * time.Second,
    
    // 🔥 Flag Filtering (Resource Optimization)
    FilterConfig: cache.FilterConfig{
        OnlyEnabled:       true,              // Cache only enabled flags
        ServiceName:       "user-service",    // Your service name
        RequireServiceTag: true,              // Only flags tagged for this service
        AdditionalTags:    []string{"production"}, // Environment filter
        TagMatchMode:      "any",             // "any" or "all"
    },
}

c, err := cache.New(
    cache.WithConfig(config),
    // ... other options
)
```

### Filter Configuration Examples

#### Example 1: Filter by Service (Microservices)

```go
// Only cache flags tagged with "user-service"
c, err := cache.New(
    // ... dependencies
    cache.WithServiceTag("user-service", true),
    cache.WithOnlyEnabled(true),
)

// Memory savings: If you have 1000 flags but only 50 are for "user-service"
// You save: ~950 KB of memory (95% reduction)
```

#### Example 2: Filter by Environment

```go
// Only cache production flags
c, err := cache.New(
    // ... dependencies
    cache.WithAdditionalTags([]string{"production"}, "any"),
    cache.WithOnlyEnabled(true),
)
```

#### Example 3: Combined Filtering

```go
// Only cache enabled production flags for user-service
c, err := cache.New(
    // ... dependencies
    cache.WithOnlyEnabled(true),
    cache.WithServiceTag("user-service", true),
    cache.WithAdditionalTags([]string{"production"}, "all"),
)
```

---

## 🎯 Advanced Features

### Webhook Integration

```go
// Start webhook server for real-time updates
webhook := server.NewWebhookServer(c, 8081, "webhook-secret")
go webhook.Start()

// Flagr sends webhooks on flag updates:
// POST /webhook
// {
//   "event": "flag.updated",
//   "flag_keys": ["my-feature"]
// }
```

### Admin API

```go
// Start admin server
admin := server.NewAdminServer(c, 8082)
go admin.Start()

// Endpoints:
// GET  /health                - Health check
// GET  /admin/stats           - Cache statistics
// POST /admin/invalidate      - Invalidate specific flag
// POST /admin/invalidate-all  - Clear entire cache
// POST /admin/refresh         - Force refresh
```

### HTTP Middleware

```go
mw := server.NewMiddleware(c)

http.Handle("/api/", mw.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    evalCtx, _ := server.GetEvalContext(r.Context())
    cache, _ := server.GetCache(r.Context())
    
    enabled := cache.EvaluateBool(r.Context(), "feature", evalCtx)
    // ... use flag
})))
```

### Disk Persistence

```go
// Enable disk persistence for fast startup
diskStorage, err := storage.NewDiskStorage("./vexilla-cache")

c, err := cache.New(
    cache.WithStorage(diskStorage),
    // ... other options
)
```

---

## 📊 Monitoring & Observability

### Cache Metrics

```go
metrics := c.GetMetrics()

fmt.Printf("Storage Metrics:\n")
fmt.Printf("  Keys Added: %d\n", metrics.Storage.KeysAdded)
fmt.Printf("  Keys Evicted: %d\n", metrics.Storage.KeysEvicted)
fmt.Printf("  Hit Ratio: %.2f%%\n", metrics.Storage.HitRatio*100)

fmt.Printf("Circuit Breaker:\n")
fmt.Printf("  State: %v\n", metrics.CircuitOpen)
fmt.Printf("  Consecutive Fails: %d\n", metrics.ConsecutiveFails)
```

### OpenTelemetry Integration

```go
import (
    "github.com/OrlandoBitencourt/vexilla/pkg/telemetry"
    "go.opentelemetry.io/otel"
)

// Initialize OTel provider
provider, _ := telemetry.NewOTel()

// Use with cache (implementation-specific)
// Metrics automatically exported:
// - vexilla.cache.hits
// - vexilla.cache.misses
// - vexilla.evaluations
// - vexilla.refresh.duration
// - vexilla.circuit.state
```

---

## 🆚 Comparison: Vexilla vs Direct Flagr

### Scenario: Brazilian Regional Launch

**Flag Configuration:**
```json
{
  "key": "brazil_launch",
  "segments": [{
    "rollout_percent": 100,
    "constraints": [
      {"property": "country", "operator": "EQ", "value": "BR"}
    ]
  }]
}
```

**With Direct Flagr:**
```go
// Every request makes HTTP call
for req := range requests {
    result := flagrClient.PostEvaluation(ctx, &evalRequest)
    // 50-200ms latency
}
// 10K req/s = 10K HTTP calls to Flagr
```

**With Vexilla:**
```go
// Flags cached, evaluated locally
for req := range requests {
    enabled := cache.EvaluateBool(ctx, "brazil_launch", evalCtx)
    // <1ms latency, 0 HTTP requests
}
// 10K req/s = 0 HTTP calls to Flagr!
```

**Result:** 50-200x faster, zero Flagr load

---

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run with race detector
go test -race ./...

# Benchmarks
go test -bench=. -benchmem ./...
```

---

## 📖 Documentation

- [Architecture Guide](ARCHITECTURE.md) - Deep dive into design
- [API Reference](https://pkg.go.dev/github.com/OrlandoBitencourt/vexilla)
- [Examples](examples/) - Working code samples
- [Contributing](CONTRIBUTING.md) - How to contribute

---

## 🤝 Related Projects

- [**Flagr**](https://github.com/openflagr/flagr) - The feature flagging service Vexilla caches

---

## 📜 License

MIT License - see [LICENSE](LICENSE) for details

---

## 🙏 Credits

**Vexilla** stands on the shoulders of giants:

- [**Flagr**](https://github.com/openflagr/flagr) by Checkr/OpenFlagr - The feature flagging service
- [**Ristretto**](https://github.com/dgraph-io/ristretto) by DGraph - High-performance cache
- [**expr**](https://github.com/expr-lang/expr) by Anton Medvedev - Expression evaluation
- [**OpenTelemetry**](https://opentelemetry.io/) - Observability framework

---

**Vexilla** - From Latin *vexillum*, meaning "flag" or "standard" 🏴

```
██╗   ██╗███████╗██╗  ██╗██╗██╗     ██╗      █████╗ 
██║   ██║██╔════╝╚██╗██╔╝██║██║     ██║     ██╔══██╗
██║   ██║█████╗   ╚███╔╝ ██║██║     ██║     ███████║
╚██╗ ██╔╝██╔══╝   ██╔██╗ ██║██║     ██║     ██╔══██║
 ╚████╔╝ ███████╗██╔╝ ██╗██║███████╗███████╗██║  ██║
  ╚═══╝  ╚══════╝╚═╝  ╚═╝╚═╝╚══════╝╚══════╝╚═╝  ╚═╝
```

*Built with ❤️ for teams who need blazing-fast feature flags without compromising on Flagr's power*
