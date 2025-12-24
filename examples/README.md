# Vexilla Examples

Comprehensive examples demonstrating all Vexilla features and use cases.

## Prerequisites

1. **Start Flagr** (if not already running):
```bash
docker run -d --name flagr -p 18000:18000 checkr/flagr
```

2. **Setup test flags** (run once):
```bash
cd examples
go run setup-flags.go
```

## Quick Start

Each example is self-contained and can be run independently:

```bash
cd examples/<example-name>
go run main.go
```

## Examples Overview

### [01-basic-usage](./01-basic-usage) - Getting Started

**What it demonstrates:**
- Creating and starting a Vexilla client
- Boolean, string, and integer flag evaluation
- Detailed evaluation with custom attachments
- Multiple user contexts
- A/B test distribution
- Cache metrics

**When to use this:** Learning Vexilla basics, first-time setup

**Key features:**
- ✅ Simple API usage
- ✅ Different flag types
- ✅ User context management
- ✅ Performance metrics

**Run it:**
```bash
cd 01-basic-usage
go run main.go
```

---

### [02-microservices](./02-microservices) - Microservice Optimization

**What it demonstrates:**
- Flag filtering for microservices
- Service-specific tag filtering
- Memory optimization (90-95% reduction)
- Premium feature gating
- Regional rollouts
- Gradual rollout strategies

**When to use this:** Microservice architectures, memory optimization

**Key features:**
- ✅ Service tag filtering
- ✅ 90-95% memory savings
- ✅ Production patterns
- ✅ Multi-region support

**Run it:**
```bash
cd 02-microservices
go run main.go
```

---

### [03-deterministic-rollout](./03-deterministic-rollout) - 🆕 Offline-First Evaluation

**What it demonstrates:**
- Pre-computed bucket-based rollouts
- 100% local evaluation (zero HTTP)
- CPF-based bucketing (Brazilian tax ID)
- User ID hash-based distribution
- Multi-variant A/B testing
- Regional + bucket combinations

**When to use this:** High-performance requirements, offline scenarios, mobile apps

**Key features:**
- ✅ <1ms evaluation latency
- ✅ Deterministic (same input = same result)
- ✅ Zero HTTP overhead
- ✅ Reproducible for debugging

**Performance:**
- Local: <1ms, 0 HTTP requests
- vs Remote: 50-200ms, 1 HTTP request
- **Speedup: 50-200x faster**

**Run it:**
```bash
cd 03-deterministic-rollout
go run main.go
```

---

### [04-webhook-invalidation](./04-webhook-invalidation) - 🆕 Real-Time Updates

**What it demonstrates:**
- Webhook-based cache invalidation
- Real-time flag updates from Flagr
- HMAC signature verification
- Event-driven architecture
- Sub-second update propagation

**When to use this:** Real-time requirements, instant flag updates

**Key features:**
- ✅ <1 second update latency
- ✅ Event-driven (no polling)
- ✅ Secure (HMAC-SHA256)
- ✅ Selective invalidation

**Setup required:**
```yaml
# Add to Flagr configuration
webhooks:
  - url: http://localhost:18001/webhook
    secret: my-webhook-secret-key
    events:
      - flag.updated
      - flag.deleted
```

**Run it:**
```bash
cd 04-webhook-invalidation
go run main.go
```

---

### [05-admin-server](./05-admin-server) - 🆕 Operational Management

**What it demonstrates:**
- Admin REST API for cache management
- Health checks and monitoring
- Manual cache invalidation
- Force refresh operations
- Cache statistics API

**When to use this:** Production operations, debugging, monitoring

**Key features:**
- ✅ REST API for management
- ✅ Health checks
- ✅ Manual invalidation
- ✅ Operational visibility

**API Endpoints:**
- `GET /health` - Health check
- `GET /admin/stats` - Cache statistics
- `POST /admin/invalidate` - Invalidate flag
- `POST /admin/invalidate-all` - Clear cache
- `POST /admin/refresh` - Force refresh

**Run it:**
```bash
cd 05-admin-server
go run main.go
```

**Test it:**
```bash
curl http://localhost:19000/health
curl http://localhost:19000/admin/stats
curl -X POST http://localhost:19000/admin/refresh
```

---

### [06-http-middleware](./06-http-middleware) - 🆕 HTTP Integration

**What it demonstrates:**
- HTTP middleware for context injection
- Feature gating in REST APIs
- A/B testing in HTTP handlers
- Regional feature detection
- Request-based evaluation

**When to use this:** REST APIs, web applications, HTTP services

**Key features:**
- ✅ Automatic context extraction
- ✅ Clean HTTP integration
- ✅ Feature gating
- ✅ A/B testing support

**API Examples:**
- Feature gating: Check premium access
- A/B testing: Different pricing layouts
- Regional: Brazil-specific features

**Run it:**
```bash
cd 06-http-middleware
go run main.go
```

**Test it:**
```bash
curl -H 'X-User-ID: user-123' -H 'X-User-Tier: premium' \
     http://localhost:8080/api/dashboard

curl -H 'X-User-ID: user-456' -H 'X-Country: BR' \
     http://localhost:8080/api/pricing
```

---

### [07-disk-persistence](./07-disk-persistence) - 🆕 Fast Startup & Recovery

**What it demonstrates:**
- Disk-based cache persistence
- Warm cache on restart
- Recovery when Flagr is down
- Fast startup (10-20x faster)
- Last-known-good state preservation

**When to use this:** Fast startup requirements, resilience, recovery scenarios

**Key features:**
- ✅ 50-100ms warm startup
- ✅ Works offline
- ✅ Survives restarts
- ✅ Last-known-good fallback

**Performance:**
- Cold start: 500-2000ms (fetch from Flagr)
- Warm start: 50-100ms (load from disk)
- **Speedup: 10-20x faster**

**Run it:**
```bash
cd 07-disk-persistence
go run main.go
```

---

### [08-circuit-breaker](./08-circuit-breaker) - 🆕 Resilience & Protection

**What it demonstrates:**
- Circuit breaker state machine
- Protection against Flagr outages
- Fail-fast behavior
- Automatic recovery testing
- Cascade failure prevention

**When to use this:** Production resilience, high availability requirements

**Key features:**
- ✅ Fail fast (<1ms vs 30s timeout)
- ✅ Automatic recovery
- ✅ Three states: CLOSED, OPEN, HALF-OPEN
- ✅ Configurable thresholds

**States:**
- **CLOSED**: Normal operation
- **OPEN**: Failing fast (Flagr down)
- **HALF-OPEN**: Testing recovery

**Run it:**
```bash
cd 08-circuit-breaker
go run main.go
```

**Manual test:**
```bash
# Open circuit: Stop Flagr
docker stop flagr

# Watch circuit OPEN after 3 failures

# Close circuit: Start Flagr
docker start flagr

# Watch recovery: OPEN → HALF-OPEN → CLOSED
```

---

### [09-telemetry](./09-telemetry) - 🆕 Observability & Monitoring

**What it demonstrates:**
- OpenTelemetry integration
- Built-in metrics access
- Cache performance analysis
- Health status monitoring
- Production monitoring patterns

**When to use this:** Production monitoring, performance analysis, debugging

**Key features:**
- ✅ OpenTelemetry metrics
- ✅ Prometheus integration
- ✅ Grafana dashboards
- ✅ Alerting rules

**Metrics exported:**
- `vexilla.cache.hits` - Cache hit counter
- `vexilla.cache.misses` - Cache miss counter
- `vexilla.evaluations` - Total evaluations
- `vexilla.refresh.duration` - Refresh latency
- `vexilla.circuit.state` - Circuit breaker state

**Run it:**
```bash
cd 09-telemetry
go run main.go
```

---

### [10-advanced-filtering](./10-advanced-filtering) - 🆕 Memory Optimization

**What it demonstrates:**
- Advanced filtering strategies
- Service-specific tag filtering
- Multiple tag matching (ANY/ALL)
- Combined filtering
- Memory savings calculation (50-95% reduction)

**When to use this:** Large-scale deployments, microservices, memory optimization

**Key features:**
- ✅ 90-95% memory reduction
- ✅ Service tag filtering
- ✅ Environment filtering
- ✅ Combined strategies

**Filtering options:**
1. **OnlyEnabled**: 20-40% savings
2. **ServiceTag**: 90-95% savings
3. **AdditionalTags (ANY)**: 50-80% savings
4. **AdditionalTags (ALL)**: 70-90% savings
5. **Combined**: 95-97% savings

**Run it:**
```bash
cd 10-advanced-filtering
go run main.go
```

---

### [99-complete-api](./99-complete-api) - 🎯 Complete Demo Application

**What it demonstrates:**
- Full-stack application (Go + Next.js)
- Real-world shopping API with checkout flows
- Deterministic rollout with visual explanation
- User context simulation
- Admin operations and metrics
- Rate limiting with feature flags
- Kill switches for emergency rollback
- Production-ready patterns

**When to use this:** Understanding complete Vexilla integration, demo/presentation, reference architecture

**Key features:**
- ✅ Complete backend API (Go + Gin)
- ✅ Interactive frontend (Next.js + TypeScript)
- ✅ 6 production feature flags
- ✅ Visual rollout explanation
- ✅ User simulator (5 personas)
- ✅ Admin panel with metrics
- ✅ Cache invalidation UI

**Feature Flags:**
1. **api.checkout.v2** - New checkout experience toggle
2. **api.checkout.rollout** - Gradual rollout (0-100%)
3. **api.rate_limit.enabled** - Dynamic rate limiting
4. **api.kill_switch** - Emergency shutdown
5. **frontend.new_ui** - New UI components
6. **frontend.beta_banner** - Beta tester banner

**Quick Start:**
```bash
cd 99-complete-api

# Setup flags (run once)
bash scripts/setup-flags.sh
# or: go run ../setup-flags.go

# Terminal 1: Start backend
cd backend
go run main.go

# Terminal 2: Start frontend
cd frontend
npm install
npm run dev
```

**Access:**
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Flagr Admin: http://localhost:18000

**Demo Flow:**
1. Switch between users (Alice, Bruno, Carlos, Diana, Enterprise)
2. Watch deterministic rollout in action
3. See different checkout versions (V1 vs V2)
4. Toggle flags in Flagr admin
5. Invalidate cache from admin panel
6. Observe instant updates

---

## Feature Matrix

| Feature | Examples |
|---------|----------|
| **Basic Usage** | 01 |
| **Microservice Filtering** | 02, 10 |
| **Deterministic Rollouts** | 03, 99 |
| **Real-time Updates** | 04, 99 |
| **Admin API** | 05, 99 |
| **HTTP Integration** | 06, 99 |
| **Disk Persistence** | 07 |
| **Circuit Breaker** | 08 |
| **Telemetry** | 09 |
| **Memory Optimization** | 02, 10 |
| **Full-Stack Demo** | 99 |
| **Kill Switches** | 99 |
| **Rate Limiting** | 99 |

## Use Case Guide

### Choose the Right Example

**Learning Vexilla:**
- Start with: [01-basic-usage](./01-basic-usage)
- Then: [02-microservices](./02-microservices)
- Complete demo: [99-complete-api](./99-complete-api) - Full-stack reference

**Presentations & Demos:**
- [99-complete-api](./99-complete-api) - Interactive full-stack demo with visual UI

**High Performance:**
- [03-deterministic-rollout](./03-deterministic-rollout) - Offline evaluation
- [07-disk-persistence](./07-disk-persistence) - Fast startup

**Production Operations:**
- [04-webhook-invalidation](./04-webhook-invalidation) - Real-time updates
- [05-admin-server](./05-admin-server) - Management API
- [08-circuit-breaker](./08-circuit-breaker) - Resilience
- [09-telemetry](./09-telemetry) - Monitoring
- [99-complete-api](./99-complete-api) - Complete reference architecture

**Memory Optimization:**
- [02-microservices](./02-microservices) - Basic filtering
- [10-advanced-filtering](./10-advanced-filtering) - Advanced strategies

**HTTP APIs:**
- [06-http-middleware](./06-http-middleware) - REST integration
- [99-complete-api](./99-complete-api) - Full REST API + Frontend

## Running All Examples

Run all CLI examples in sequence:

```bash
#!/bin/bash
# run-all-examples.sh

examples=(
    "01-basic-usage"
    "02-microservices"
    "03-deterministic-rollout"
    "04-webhook-invalidation"
    "05-admin-server"
    "06-http-middleware"
    "07-disk-persistence"
    "08-circuit-breaker"
    "09-telemetry"
    "10-advanced-filtering"
)

for example in "${examples[@]}"; do
    echo "Running $example..."
    cd "$example"
    go run main.go
    cd ..
    echo ""
done
```

**Note:** Example [99-complete-api](./99-complete-api) is a full-stack application that requires separate backend and frontend processes. See its [README](./99-complete-api/README.md) for instructions.

## Common Setup

All examples use the same Flagr setup. Run this once:

```bash
# 1. Start Flagr
docker run -d --name flagr -p 18000:18000 checkr/flagr

# 2. Setup test flags
cd examples
go run setup-flags.go

# 3. Verify
curl http://localhost:18000/api/v1/flags
```

## Performance Benchmarks

| Feature | Latency | Throughput | Notes |
|---------|---------|------------|-------|
| Local evaluation | <1ms | 200,000/sec | Deterministic flags |
| Remote evaluation | 50-200ms | ~6/sec | Percentage rollouts |
| Cache hit | <1μs | Millions/sec | Ristretto lookup |
| Warm startup | 50-100ms | N/A | With disk persistence |
| Cold startup | 500-2000ms | N/A | Fetch from Flagr |

## Memory Usage

| Scenario | Flags Cached | Memory |
|----------|--------------|--------|
| No filtering | 10,000 | 9.77 MB |
| OnlyEnabled | 7,000 | 6.84 MB |
| Service tags | 500 | 0.49 MB |
| Combined | 300 | 0.29 MB |

## Troubleshooting

### Flagr Not Running

```bash
# Check if Flagr is running
curl http://localhost:18000/api/v1/health

# Start Flagr
docker run -d --name flagr -p 18000:18000 checkr/flagr
```

### No Flags Available

```bash
# Setup test flags
cd examples
go run setup-flags.go
```

### Port Already in Use

```bash
# Webhook server (04-webhook-invalidation)
# Change port in code or stop conflicting service

# Admin server (05-admin-server)
# Change port in code or stop conflicting service

# HTTP server (06-http-middleware)
# Change port in code or stop conflicting service
```

## Next Steps

After exploring the examples:

1. **Read the [Architecture Document](../ARCHITECTURE.md)** for deep dive
2. **Check the [Main README](../README.md)** for library documentation
3. **Review the [API Documentation](https://pkg.go.dev/github.com/OrlandoBitencourt/vexilla)**
4. **Integrate into your application**

## Contributing

Found an issue or have an improvement?

1. Open an issue: https://github.com/OrlandoBitencourt/vexilla/issues
2. Submit a PR with example improvements
3. Share your use case for new examples

## License

See [LICENSE](../LICENSE) in the root directory.

---

**Built with ❤️ for high-performance feature flagging**
