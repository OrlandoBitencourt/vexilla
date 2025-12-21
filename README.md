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

**Real benchmark results (AMD Ryzen 5 5600G, 12 cores, Windows):**

| Metric | Flagr Direct | Vexilla (Local) | Vexilla (Remote) | Improvement |
|--------|--------------|-----------------|------------------|-------------|
| **Latency** | 50-200ms | **335 ns** (0.335 μs) | 50-200ms | **>400,000x faster** |
| **HTTP Requests** | 1 per eval | 0 per eval | 1 per eval | **100% reduction** |
| **Throughput** | ~2K req/s | **37.7M ops/s** | ~2K req/s | **18,850x higher** |
| **Memory Usage** | N/A | **448 B/eval** | ~1KB/flag | Configurable filtering |
| **Concurrent (12 cores)** | N/A | **85.85 ns** | N/A | **11.6M ops/sec** |
| **Client API** | N/A | **73 ns** | N/A | **13.7M ops/sec** |

[Full benchmark results →](benchmarks/results/REAL_RESULTS.md)

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

    "github.com/OrlandoBitencourt/vexilla"
)

func main() {
    // Create client with simple configuration
    client, err := vexilla.New(
        vexilla.WithFlagrEndpoint("http://localhost:18000"),
        vexilla.WithRefreshInterval(5 * time.Minute),
        vexilla.WithOnlyEnabled(true),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Start the client
    ctx := context.Background()
    if err := client.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Stop()

    // Evaluate a flag
    evalCtx := vexilla.NewContext("user-123").
        WithAttribute("country", "BR").
        WithAttribute("tier", "premium")

    enabled := client.Bool(ctx, "new-feature", evalCtx)
    log.Printf("Feature enabled: %v", enabled)
}
```

### Production Setup with Admin & Webhook

```go
client, err := vexilla.New(
    // Flagr connection
    vexilla.WithFlagrEndpoint("http://flagr:18000"),
    vexilla.WithRefreshInterval(5 * time.Minute),
    
    // Admin API for operations
    vexilla.WithAdminServer(vexilla.AdminConfig{
        Port: 19000,
    }),
    
    // Webhook for real-time updates
    vexilla.WithWebhookInvalidation(vexilla.WebhookConfig{
        Port:   18001,
        Secret: os.Getenv("WEBHOOK_SECRET"),
    }),
    
    // Resource optimization
    vexilla.WithServiceTag("user-service"),
    vexilla.WithOnlyEnabled(true),
    
    // Resilience
    vexilla.WithCircuitBreaker(3, 30*time.Second),
)

// Admin endpoints now available at http://localhost:19000
// Webhook listening at http://localhost:18001
```

### HTTP Middleware Integration

```go
// Automatic request context injection
handler := client.HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    userID := r.Header.Get("X-User-ID")
    evalCtx := vexilla.NewContext(userID)
    
    if client.Bool(r.Context(), "new-feature", evalCtx) {
        // New feature enabled
    }
}))

http.ListenAndServe(":8080", handler)
```

### Microservice Usage (with filtering)

```go
// Filter flags by service tag to reduce memory usage
client, err := vexilla.New(
    vexilla.WithFlagrEndpoint("http://localhost:18000"),
    vexilla.WithServiceTag("user-service"),      // Only cache flags tagged "user-service"
    vexilla.WithOnlyEnabled(true),               // Only cache enabled flags
    vexilla.WithAdditionalTags([]string{"production"}, "any"), // Environment filtering
)

// With 10,000 total flags but only 50 for user-service:
// Memory usage: ~10MB → ~500KB (95% reduction!)
```

- [Click and see more examples](examples/example_readme.md)
- [Deterministic Rollout (Bucket-based Evaluation)](examples/deterministic_rollout.md)

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
- **Memory savings** - Reduce memory footprint by 90-95% in microservices

### 🔔 Real-time Updates
- **Webhook support** - Instant flag updates from Flagr
- **Event-driven** - No polling overhead
- **Signature verification** - Secure HMAC-SHA256 webhook validation
- **Sub-second propagation** - Updates in <1s vs 5min refresh

### 🛠️ Operations & Management
- **Admin API** - REST endpoints for cache management
  - `GET /health` - Health check
  - `GET /admin/stats` - Cache metrics
  - `POST /admin/invalidate` - Invalidate specific flag
  - `POST /admin/invalidate-all` - Clear cache
  - `POST /admin/refresh` - Force refresh
- **HTTP middleware** - Drop-in request context injection
- **Graceful shutdown** - Clean resource cleanup

### 📊 Observability
- **Full OpenTelemetry** - Traces and metrics
- **Cache statistics** - Hit ratios, evictions, etc.
- **Evaluation tracking** - Local vs remote routing
- **Performance monitoring** - Latency, throughput, error rates

---

## 📖 API Examples

### Boolean Flag

```go
enabled := client.Bool(ctx, "new-feature", evalCtx)
```

### String Flag

```go
theme := client.String(ctx, "ui-theme", evalCtx, "light")
```

### Integer Flag

```go
maxItems := client.Int(ctx, "max-items", evalCtx, 10)
```

### Detailed Evaluation

```go
result, err := client.Evaluate(ctx, "ab-test", evalCtx)
if err == nil {
    fmt.Printf("Variant: %s\n", result.VariantKey)
    fmt.Printf("Reason: %s\n", result.EvaluationReason)
}

// Access custom variant data
tier := result.GetString("tier", "free")
limit := result.GetInt("limit", 100)
```

### Fluent Context Building

```go
evalCtx := vexilla.NewContext("user-456").
    WithAttribute("country", "US").
    WithAttribute("tier", "premium").
    WithAttribute("age", 25).
    WithEntityType("user")
```

---

## 🔧 Configuration Options

### Flagr Connection

```go
vexilla.WithFlagrEndpoint("http://localhost:18000")
vexilla.WithFlagrAPIKey("your-api-key")
vexilla.WithFlagrTimeout(5 * time.Second)
vexilla.WithFlagrMaxRetries(3)
```

### Cache Behavior

```go
vexilla.WithRefreshInterval(5 * time.Minute)
vexilla.WithInitialTimeout(10 * time.Second)
```

### Fallback Strategy

```go
vexilla.WithFallbackStrategy("fail_closed")  // Options: fail_open, fail_closed, error
```

### Circuit Breaker

```go
vexilla.WithCircuitBreaker(3, 30*time.Second)  // threshold, timeout
```

### Flag Filtering (Memory Optimization)

```go
// Only cache enabled flags
vexilla.WithOnlyEnabled(true)

// Only cache flags tagged for this service
vexilla.WithServiceTag("user-service")

// Only cache flags with specific tags
vexilla.WithAdditionalTags([]string{"production", "critical"}, "any")
```

### Server Features

```go
// Admin API for operations and monitoring
vexilla.WithAdminServer(vexilla.AdminConfig{
    Port: 19000,
})

// Webhook for real-time flag updates from Flagr
vexilla.WithWebhookInvalidation(vexilla.WebhookConfig{
    Port:   18001,
    Secret: "your-webhook-secret",
})
```

### Using a Config Struct

```go
cfg := vexilla.DefaultConfig()
cfg.Flagr.Endpoint = "http://localhost:18000"
cfg.Cache.RefreshInterval = 10 * time.Minute
cfg.Cache.Filter.OnlyEnabled = true
cfg.Cache.Filter.ServiceName = "payment-service"
cfg.Cache.Filter.RequireServiceTag = true

client, err := vexilla.New(vexilla.WithConfig(cfg))
```

---

## 🌐 Server Features

### Admin API

Monitor and manage your cache through REST endpoints:

```bash
# Health check
curl http://localhost:19000/health

# Get cache statistics
curl http://localhost:19000/admin/stats

# Invalidate specific flag
curl -X POST http://localhost:19000/admin/invalidate \
  -H "Content-Type: application/json" \
  -d '{"flag_key": "new-feature"}'

# Force refresh all flags
curl -X POST http://localhost:19000/admin/refresh

# Clear entire cache
curl -X POST http://localhost:19000/admin/invalidate-all
```

### Webhook Integration

Receive real-time updates from Flagr:

**Configure in Flagr:**
1. Go to Settings > Webhooks
2. Add URL: `http://vexilla:18001/webhook`
3. Set shared secret (must match Vexilla config)
4. Enable events: `flag.updated`, `flag.deleted`

**Payload Example:**
```json
{
  "event": "flag.updated",
  "flag_keys": ["new-feature", "beta-access"],
  "timestamp": "2025-12-20T10:30:00Z"
}
```

**Benefits:**
- **Sub-second updates** vs 5-minute polling
- **Reduced load** on Flagr (no constant polling)
- **Secure** with HMAC-SHA256 signature verification

### HTTP Middleware

Automatically inject request context into flag evaluations:

```go
mux := http.NewServeMux()

mux.HandleFunc("/api/features", func(w http.ResponseWriter, r *http.Request) {
    // Middleware automatically extracts context from request
    userID := r.Header.Get("X-User-ID")
    evalCtx := vexilla.NewContext(userID)
    
    features := map[string]bool{
        "new_dashboard": client.Bool(r.Context(), "new-dashboard", evalCtx),
        "beta_features": client.Bool(r.Context(), "beta-features", evalCtx),
    }
    
    json.NewEncoder(w).Encode(features)
})

// Wrap with middleware
handler := client.HTTPMiddleware(mux)
http.ListenAndServe(":8080", handler)
```

**For detailed server documentation, see:**
- [Server Features Guide](SERVER_FEATURES.md)
- [Server Examples](examples/example_server.go)

---

## 🎓 Understanding Evaluation Strategies

Vexilla automatically determines the best evaluation strategy:

### ✅ Local Evaluation (Static Flags)

**Conditions:**
- `rollout_percent: 100` (everyone gets evaluated)
- Single distribution with `percentage: 100`
- Deterministic constraints only

**Performance:** <1ms, 0 HTTP requests

### ❌ Flagr Evaluation (Dynamic Flags)

**Requires Flagr when:**
- `rollout_percent < 100` (partial rollout)
- Multiple distributions (A/B test)
- Single distribution with `percentage < 100`

**Performance:** 50-200ms, 1 HTTP request (in a real world scenario where we have lots of services hitting flagr it can take >1s to evaluate).

## Deterministic & Offline Rollouts

Most feature flag solutions rely on **random percentage-based rollouts**, which introduce
non-deterministic behavior, require HTTP calls, and make debugging difficult.

Vexilla supports **deterministic, bucket-based rollouts**, allowing feature decisions to be made
**entirely locally**, without randomness or network dependency.

### Key benefits

- ✅ Deterministic flag evaluation
- ✅ Offline-first / cache-friendly
- ✅ No HTTP calls during runtime
- ✅ Predictable and reproducible rollouts
- ✅ Easier debugging and incident analysis

Instead of rolling out features to "70% of users at random", Vexilla allows you to:

- Pre-process a user identifier (e.g. CPF, user ID, account ID)
- Convert it into a numeric bucket (`0–99`)
- Evaluate feature flags using simple numeric rules

👉 See the full example here:  
[Deterministic Rollout (Bucket-based Evaluation)](examples/deterministic_rollout.md)

---

## 📊 Monitoring & Metrics

### Programmatic Access

```go
metrics := client.Metrics()

fmt.Printf("Cache Performance:\n")
fmt.Printf("  Keys Added: %d\n", metrics.Storage.KeysAdded)
fmt.Printf("  Keys Evicted: %d\n", metrics.Storage.KeysEvicted)
fmt.Printf("  Hit Ratio: %.2f%%\n", metrics.Storage.HitRatio*100)

fmt.Printf("\nHealth Status:\n")
fmt.Printf("  Last Refresh: %s\n", metrics.LastRefresh)
fmt.Printf("  Circuit Open: %v\n", metrics.CircuitOpen)
fmt.Printf("  Consecutive Fails: %d\n", metrics.ConsecutiveFails)
```

### Admin API

```bash
# Get all metrics via HTTP
curl http://localhost:19000/admin/stats | jq
```

### OpenTelemetry

Vexilla exports metrics and traces via OpenTelemetry:

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/sdk/metric"
)

// Metrics exported:
// - vexilla.cache.hits
// - vexilla.cache.misses
// - vexilla.evaluations (with strategy label)
// - vexilla.refresh.duration
// - vexilla.circuit.state
```

---

## 🆚 Comparison: Direct Flagr vs Vexilla

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

### With Direct Flagr

```go
// Every request makes HTTP call
for req := range requests {
    result := flagrClient.PostEvaluation(ctx, &evalRequest)
    // 50-200ms latency
}
// 10K req/s = 10K HTTP calls to Flagr
```

### With Vexilla

```go
// Flags cached, evaluated locally
for req := range requests {
    enabled := client.Bool(ctx, "brazil_launch", evalCtx)
    // <1ms latency, 0 HTTP requests
}
// 10K req/s = 0 HTTP calls to Flagr!
```

**Result:** 50-200x faster, zero Flagr load

---

## 📚 Documentation

- **[Server Features Guide](SERVER_FEATURES.md)** - Admin API, webhooks, middleware
- **[Architecture Guide](ARCHITECTURE.md)** - Deep dive into design
- **[API Reference](https://pkg.go.dev/github.com/OrlandoBitencourt/vexilla)** - Complete API docs
- **[Examples](examples/)** - Working code samples

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

# Run benchmarks
cd benchmarks
./run_benchmarks.sh    # Linux/macOS
./run_benchmarks.bat   # Windows

# Quick benchmarks (1-2 minutes)
./quick_bench.sh
```

**Latest Benchmark Results:**
- Local evaluation (simple): **335 ns/op** (0.335 μs)
- Complex constraints: **625 ns/op** (0.625 μs)
- Concurrent throughput: **11.6M ops/sec** (37.7M single-core)
- Client API: **73 ns/op** (13.7M ops/sec)
- Memory per evaluation: **448 bytes**, 6 allocations
- Cache hit: **364.9 ns/op** (10.4M ops/sec)

See [benchmarks/results/REAL_RESULTS.md](benchmarks/results/REAL_RESULTS.md) for detailed performance data.

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