# 🏴 Vexilla Examples

Complete, working examples demonstrating Vexilla usage in real-world scenarios.

## 📋 Prerequisites

### 1. Start Flagr Server

Choose one of these options:

**Option A: Docker (Recommended)**
```bash
docker run -it -p 18000:18000 ghcr.io/openflagr/flagr:latest
```

**Option B: Docker Compose**
```bash
cd vexilla
docker-compose up -d
```

**Option C: Local Binary**
Download from [Flagr Releases](https://github.com/openflagr/flagr/releases)

### 2. Create Test Flags

Run the setup script to create all example flags:

```bash
cd examples
go run setup-flags.go
```

This creates 10 flags for testing:
- ✅ `new_feature` - Simple boolean flag
- ✅ `dark_mode` - UI theme toggle
- ✅ `ui_theme` - A/B test (dark vs light)
- ✅ `max_items` - Integer configuration
- ✅ `premium_features` - Tier-based access
- ✅ `beta_access` - Beta program flag
- ✅ `button_color_test` - A/B test (blue vs red)
- ✅ `pricing_layout` - Multi-variant test (3 layouts)
- ✅ `gradual_rollout_30` - 30% gradual rollout
- ✅ `brazil_launch` - Regional launch flag

This is how your flagr UI should look like after running setup-flag.go:
![flagr ui](https://raw.githubusercontent.com/OrlandoBitencourt/vexilla/refs/heads/main/examples/after-setup-flags.png)

## 🚀 Examples

### Example 1: Basic Usage

**Location:** `01-basic-usage/main.go`

Learn fundamental concepts:
- Creating and starting a client
- Boolean, string, and integer flag evaluation
- Building evaluation contexts
- Accessing detailed results
- Reading cache metrics

**Run:**
```bash
cd 01-basic-usage
go run main.go
```

**Output:**
```
🏴 Vexilla Basic Example
======================================================================

📦 Creating Vexilla client...
🚀 Starting client and loading flags...
✅ Client started successfully!

Example 1: Boolean Flag Evaluation
----------------------------------------------------------------------
Flag: new_feature
User: user-123
Result: true

...
```

### Example 2: Microservices

**Location:** `02-microservices/main.go`

Advanced patterns for microservice architectures:
- Service-specific flag filtering
- Memory optimization with tags
- Regional feature rollouts
- Gradual rollout strategies
- Multi-variant A/B tests
- Performance monitoring

**Run:**
```bash
cd 02-microservices
go run main.go
```

**Output:**
```
🏴 Vexilla Microservice Example
======================================================================

Use Case 1: User Registration Features
----------------------------------------------------------------------
Beta Access Available: true
  → User can access beta features

Use Case 4: Gradual Rollout (30% in Brazil)
----------------------------------------------------------------------
Total Brazilian Users: 100
  ✅ Enabled: 28 (28%)
  ❌ Disabled: 72 (72%)

💾 Memory Optimization:
  Without filtering: ~9.77 MB (10,000 flags)
  With filtering: ~0.01 MB (10 flags)
  Memory saved: ~9.76 MB (99.9%)
```

## 📖 Common Patterns

### Pattern 1: Simple Feature Toggle

```go
client, _ := vexilla.New(
    vexilla.WithFlagrEndpoint("http://localhost:18000"),
)
client.Start(context.Background())
defer client.Stop()

ctx := vexilla.NewContext("user-123")
enabled := client.Bool(context.Background(), "new_feature", ctx)

if enabled {
    // Show new feature
}
```

### Pattern 2: User Attribute Targeting

```go
ctx := vexilla.NewContext("user-456").
    WithAttribute("country", "BR").
    WithAttribute("tier", "premium").
    WithAttribute("age", 25)

enabled := client.Bool(context.Background(), "premium_features", ctx)
```

### Pattern 3: A/B Testing

```go
result, _ := client.Evaluate(context.Background(), "button_test", ctx)

switch result.VariantKey {
case "blue":
    renderBlueButton()
case "red":
    renderRedButton()
default:
    renderDefaultButton()
}
```

### Pattern 4: Configuration Values

```go
// String configuration
theme := client.String(ctx, "ui_theme", evalCtx, "light")

// Integer configuration  
maxItems := client.Int(ctx, "page_size", evalCtx, 20)

// Detailed configuration
result, _ := client.Evaluate(ctx, "layout_config", evalCtx)
columns := result.GetInt("columns", 3)
spacing := result.GetString("spacing", "normal")
```

### Pattern 5: Microservice Filtering

```go
// Only cache flags for this service
client, _ := vexilla.New(
    vexilla.WithFlagrEndpoint("http://localhost:18000"),
    vexilla.WithServiceTag("user-service"),
    vexilla.WithOnlyEnabled(true),
)

// If you have 10,000 flags but only 50 for user-service:
// Memory: ~10MB → ~50KB (99.5% reduction!)
```

### Pattern 6: Regional Rollout

```go
regions := []string{"BR", "US", "UK", "JP"}

for _, region := range regions {
    ctx := vexilla.NewContext("user").
        WithAttribute("country", region)
    
    enabled := client.Bool(context.Background(), "regional_feature", ctx)
    fmt.Printf("%s: %v\n", region, enabled)
}
```

## 🔧 Configuration

### Environment Variables

```bash
# Flagr connection
export FLAGR_ENDPOINT="http://localhost:18000"
export FLAGR_API_KEY="your-api-key"  # Optional

# Cache settings
export REFRESH_INTERVAL="5m"
export INITIAL_TIMEOUT="10s"

# Filtering
export ONLY_ENABLED="true"
export SERVICE_NAME="user-service"
```

### Programmatic Configuration

```go
client, err := vexilla.New(
    // Required
    vexilla.WithFlagrEndpoint("http://localhost:18000"),
    
    // Optional - Authentication
    vexilla.WithFlagrAPIKey("secret-key"),
    
    // Optional - Cache behavior
    vexilla.WithRefreshInterval(5 * time.Minute),
    vexilla.WithInitialTimeout(10 * time.Second),
    
    // Optional - Fallback strategy
    vexilla.WithFallbackStrategy("fail_closed"), // or "fail_open", "error"
    
    // Optional - Circuit breaker
    vexilla.WithCircuitBreaker(3, 30*time.Second),
    
    // Optional - Filtering (memory optimization)
    vexilla.WithOnlyEnabled(true),
    vexilla.WithServiceTag("user-service"),
    vexilla.WithAdditionalTags([]string{"production"}, "any"),
)
```

## 📊 Monitoring

### Cache Metrics

```go
metrics := client.Metrics()

// Storage metrics
fmt.Printf("Keys Cached: %d\n", metrics.Storage.KeysAdded)
fmt.Printf("Hit Ratio: %.2f%%\n", metrics.Storage.HitRatio*100)
fmt.Printf("Keys Evicted: %d\n", metrics.Storage.KeysEvicted)

// Health metrics
fmt.Printf("Last Refresh: %s\n", metrics.LastRefresh)
fmt.Printf("Circuit Open: %v\n", metrics.CircuitOpen)
fmt.Printf("Failed Refreshes: %d\n", metrics.ConsecutiveFails)
```

### Performance Benchmarks

| Operation | Latency | Throughput |
|-----------|---------|------------|
| Local evaluation (cached) | < 1µs | > 1M ops/sec |
| Remote evaluation | 50-200ms | Depends on Flagr |
| Cache refresh | 100-500ms | Per refresh interval |

## 🐛 Troubleshooting

### Issue: Connection Refused

```
Error: failed to start client: connection refused
```

**Solution:**
```bash
# Check if Flagr is running
docker ps | grep flagr

# Start Flagr if not running
docker run -it -p 18000:18000 ghcr.io/openflagr/flagr:latest
```

### Issue: Flag Not Found

```
Error: flag not found: my-feature
```

**Solution:**
1. Run `go run setup-flags.go` to create test flags
2. Or create the flag manually in Flagr UI: http://localhost:18000
3. Check flag key spelling (exact match required)

### Issue: Circuit Breaker Open

```
Error: circuit breaker is open
```

**Solution:**
```go
// Check Flagr connectivity
client.HealthCheck(context.Background())

// Manually refresh cache
client.Sync(context.Background())

// Or restart client
client.Stop()
client.Start(context.Background())
```

### Issue: Wrong Evaluation Result

```
Expected: true, Got: false
```

**Solution:**
1. Check constraint attributes match exactly
2. Verify flag is enabled in Flagr UI
3. Check segment rollout percentage
4. Use detailed evaluation for debugging:

```go
result, _ := client.Evaluate(ctx, "flag-key", evalCtx)
fmt.Printf("Variant: %s\n", result.VariantKey)
fmt.Printf("Reason: %s\n", result.EvaluationReason)
```

## 💡 Best Practices

### 1. Use Service Filtering in Production

```go
// ❌ Bad: Caches ALL flags (wastes memory)
client, _ := vexilla.New(
    vexilla.WithFlagrEndpoint("http://localhost:18000"),
)

// ✅ Good: Only caches relevant flags
client, _ := vexilla.New(
    vexilla.WithFlagrEndpoint("http://localhost:18000"),
    vexilla.WithServiceTag("user-service"),
    vexilla.WithOnlyEnabled(true),
)
```

### 2. Choose Appropriate Refresh Intervals

```go
// High-traffic service (reduce load)
vexilla.WithRefreshInterval(10 * time.Minute)

// Low-latency requirements (faster updates)
vexilla.WithRefreshInterval(1 * time.Minute)

// Development (instant updates)
vexilla.WithRefreshInterval(10 * time.Second)
```

### 3. Monitor Cache Performance

```go
// Alert on low cache hit ratio
metrics := client.Metrics()
if metrics.Storage.HitRatio < 0.8 {
    log.Warn("Low cache hit ratio: %.2f%%", metrics.Storage.HitRatio*100)
}

// Alert on circuit breaker
if metrics.CircuitOpen {
    log.Error("Circuit breaker is open - degraded mode")
}
```

### 4. Use Fail-Safe Defaults

```go
// ❌ Bad: Can break on missing flag
enabled := client.Bool(ctx, "new_feature", evalCtx)
if enabled {
    // Might never execute if flag doesn't exist
}

// ✅ Good: Explicit fallback behavior
result, err := client.Evaluate(ctx, "new_feature", evalCtx)
if err != nil {
    log.Warn("Flag evaluation failed: %v", err)
    // Use safe default
    return
}
```

### 5. Tag Your Flags Properly

In Flagr UI, add meaningful tags:
- Service name: `user-service`, `payment-service`
- Environment: `production`, `staging`, `development`
- Team: `team-growth`, `team-platform`
- Category: `experiment`, `ops`, `feature`

## 📚 Resources

- **Documentation:** https://pkg.go.dev/github.com/OrlandoBitencourt/vexilla
- **GitHub:** https://github.com/OrlandoBitencourt/vexilla
- **Flagr UI:** http://localhost:18000 (when running)
- **Flagr Docs:** https://openflagr.github.io/flagr/

## 🤝 Contributing

Found a bug or have an improvement?

1. Open an issue: https://github.com/OrlandoBitencourt/vexilla/issues
2. Submit a PR with your example
3. Share your use case!

---

**Happy feature flagging! 🏴**
