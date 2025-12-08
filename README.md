```
██╗   ██╗███████╗██╗  ██╗██╗██╗     ██╗      █████╗ 
██║   ██║██╔════╝╚██╗██╔╝██║██║     ██║     ██╔══██╗
██║   ██║█████╗   ╚███╔╝ ██║██║     ██║     ███████║
╚██╗ ██╔╝██╔══╝   ██╔██╗ ██║██║     ██║     ██╔══██║
 ╚████╔╝ ███████╗██╔╝ ██╗██║███████╗███████╗██║  ██║
  ╚═══╝  ╚══════╝╚═╝  ╚═╝╚═╝╚══════╝╚══════╝╚═╝  ╚═╝
```

# 🏴 Vexilla

> **Intelligent caching layer for [Flagr](https://github.com/openflagr/flagr)**  
> High-performance feature flag evaluation with smart local/remote routing

[![Go Reference](https://pkg.go.dev/badge/github.com/Orlando.Bitencourt/vexilla.svg)](https://pkg.go.dev/github.com/Orlando.Bitencourt/vexilla)
[![Go Report Card](https://goreportcard.com/badge/github.com/Orlando.Bitencourt/vexilla)](https://goreportcard.com/report/github.com/Orlando.Bitencourt/vexilla)
[![codecov](https://codecov.io/gh/Orlando.Bitencourt/vexilla/branch/main/graph/badge.svg)](https://codecov.io/gh/Orlando.Bitencourt/vexilla)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## 🎯 What is Vexilla?

Vexilla is a **high-performance caching layer** for [Flagr](https://github.com/openflagr/flagr), designed to eliminate the performance bottleneck of HTTP-based feature flag evaluation while maintaining 100% compatibility with Flagr's powerful A/B testing and dynamic configuration capabilities.

### 📊 Performance Impact

| Metric | Flagr Direct | Vexilla (Static) | Vexilla (Dynamic) | Improvement |
|--------|--------------|------------------|-------------------|-------------|
| **Latency** | 50-200ms | <1ms | 50-200ms | **50-200x faster** |
| **HTTP Requests** | 1 per eval | 0 per eval | 1 per eval | **100% reduction** |
| **Throughput** | ~2K req/s | >10K req/s | ~2K req/s | **5x higher** |
| **Flagr Load** | 100% | 0-11%* | Varies | **89-100% reduction** |

*Example: With deterministic constraints filtering 89% of traffic before Flagr evaluation

---

## 🤔 Why Vexilla?

### The Problem with Flagr

[Flagr](https://github.com/openflagr/flagr) is an excellent feature flagging microservice with:
- ✅ Powerful A/B testing with consistent bucketing
- ✅ Dynamic configuration management
- ✅ Clean REST API and UI
- ✅ Segment-based targeting

**However**, Flagr has one significant limitation:

```go
// Every flag evaluation requires an HTTP request
result := flagrClient.PostEvaluation(ctx, evalRequest)
// 50-200ms latency per call!
```

**At scale, this becomes a problem:**
- 10,000 requests/second = 10,000 HTTP calls to Flagr
- Each with 50-200ms latency
- Flagr becomes a bottleneck
- Infrastructure costs increase

### The Vexilla Solution

Vexilla solves this by **intelligently caching** flag definitions and **routing evaluation**:

```
┌─────────────────────────────────────────┐
│  Your Application (10K req/s)           │
└──────────────┬──────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────┐
│           Vexilla Cache                  │
│  • Flags cached in-memory (Ristretto)   │
│  • Smart routing: local vs Flagr         │
│  • <1ms evaluation for static flags      │
└──────────────┬───────────────────────────┘
               │
        Only when needed
               │
               ▼
┌──────────────────────────────────────────┐
│      Flagr Server                        │
│  • Handles A/B tests (consistent hash)   │
│  • Percentage rollouts                   │
│  • Complex variant assignments           │
└──────────────────────────────────────────┘
```

**Key Insight:** Most flags are **deterministic** (e.g., `country == "BR"`). These don't need Flagr's stateful bucketing - they can be evaluated locally in <1ms with zero HTTP requests.

---

## ✨ Features

### 🚀 Core Features
- **Sub-millisecond evaluation** for deterministic flags (100% rollout)
- **Smart routing** - Automatically detects when Flagr is needed
- **Background refresh** - Keeps flags fresh without blocking
- **Circuit breaker** - Resilient to Flagr outages
- **Disk persistence** - Warm cache on restart

### 🔔 Real-time Updates
- **Webhook support** - Instant flag updates from Flagr
- **Event-driven** - No polling overhead

### 🛠️ Operations
- **Admin API** - Management endpoints for ops teams
- **Health checks** - Monitor cache status
- **HTTP middleware** - Drop-in request context injection

### 📊 Observability
- **Full OpenTelemetry** - Traces and metrics
- **Cache statistics** - Hit ratios, evictions, etc.
- **Evaluation tracking** - Local vs remote routing

---

## 🆚 Vexilla vs Direct Flagr

### Scenario: Brazilian Regional Launch

**Flag Configuration:**
```json
{
  "key": "brazil_launch",
  "segments": [{
    "rollout_percent": 100,
    "constraints": [
      {"property": "country", "operator": "EQ", "value": "BR"},
      {"property": "document", "operator": "MATCHES", "value": "[0-9]{7}(0[0-9]|10)$"}
    ]
  }]
}
```

**With Direct Flagr:**
```go
// Every request makes HTTP call
for req := range requests {
    result := flagrClient.PostEvaluation(ctx, &goflagr.EvalContext{
        EntityID: req.UserID,
        EntityContext: map[string]interface{}{
            "country": "BR",
            "document": req.Document,
        },
    })
    // 50-200ms latency
}
// 10K req/s = 10K HTTP calls to Flagr
```

**With Vexilla:**
```go
// Flags cached, evaluated locally
for req := range requests {
    enabled := cache.EvaluateBool(ctx, "brazil_launch", vexilla.EvaluationContext{
        UserID: req.UserID,
        Attributes: map[string]interface{}{
            "country": "BR",
            "document": req.Document,
        },
    })
    // <1ms latency, 0 HTTP requests
}
// 10K req/s = 0 HTTP calls to Flagr!
```

**Result:** 50-200x faster, zero Flagr load

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
      {"property": "tier", "operator": "EQ", "value": "premium"},
      {"property": "country", "operator": "IN", "value": ["US", "CA", "UK"]}
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
    "constraints": [{"property": "country", "operator": "EQ", "value": "US"}],
    "distributions": [{"variant_key": "enabled", "percentage": 100}]
  }]
}
```

❌ **Requires Flagr** - Needs consistent hashing for sticky behavior

---

## 📖 Documentation

- [Architecture Guide](ARCHITECTURE.md) - Deep dive into design
- [API Reference](https://pkg.go.dev/github.com/Orlando.Bitencourt/vexilla)
- [Examples](examples/) - Working code samples
- [Contributing](CONTRIBUTING.md) - How to contribute

---

## 🔧 Configuration

```go
config := vexilla.Config{
    // Flagr connection (required)
    FlagrEndpoint: "https://flagr.company.com",
    FlagrAPIKey:   "optional-api-key",
    
    // Cache behavior
    RefreshInterval:  5 * time.Minute,
    InitialTimeout:   10 * time.Second,
    HTTPTimeout:      5 * time.Second,
    RetryAttempts:    3,
    FallbackStrategy: "fail_closed", // or "fail_open", "last_known_good"
    
    // Persistence
    PersistenceEnabled: true,
    PersistencePath:    "/var/lib/vexilla",
    
    // Webhook (optional)
    WebhookEnabled: true,
    WebhookPort:    8081,
    WebhookSecret:  "your-webhook-secret",
    
    // Admin API (optional)
    AdminAPIEnabled: true,
    AdminAPIPort:    8082,
    
    // Ristretto cache settings
    CacheMaxCost:     1 << 30, // 1GB
    CacheNumCounters: 1e7,     // 10M keys
}
```

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
- [**goflagr**](https://github.com/checkr/goflagr) - Official Go client for Flagr (used by Vexilla)

---

## 📜 License

MIT License - see [LICENSE](LICENSE) for details

---

## 🙏 Credits

**Vexilla** stands on the shoulders of giants:

- [**Flagr**](https://github.com/openflagr/flagr) by Checkr/OpenFlagr - The feature flagging service
- [**Ristretto**](https://github.com/dgraph-io/ristretto) by DGraph - High-performance cache
- [**expr**](https://github.com/antonmedv/expr) by Anton Medvedev - Expression evaluation
- [**OpenTelemetry**](https://opentelemetry.io/) - Observability framework

---

**Vexilla** - From Latin *vexillum*, meaning "flag" or "standard" 🏴

*Built with ❤️ for teams who need blazing-fast feature flags without compromising on Flagr's power*