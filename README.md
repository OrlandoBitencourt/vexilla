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
``

[Click and see more examples](examples/example_readme.md)

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

---

## 📊 Monitoring & Metrics

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

# Benchmarks
go test -bench=. -benchmem ./...
```

---

## 🤝 Related Projects

- [**Flagr**](https://github.com/openflagr/flagr) - The feature flagging service Vexilla caches

---

## 🚀 Performance Benchmarks

```
BenchmarkCache_Evaluate_Local-10    200000    5432 ns/op    0 allocs/op
BenchmarkCache_Evaluate_Remote-10     2000  150234 ns/op  384 allocs/op

Speedup: 27.6x faster for local evaluation
```

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
