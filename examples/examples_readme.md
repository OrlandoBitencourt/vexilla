# Vexilla Examples

This directory contains working examples demonstrating various Vexilla use cases.

## Prerequisites

Before running the examples, you need a running Flagr instance:

### Option 1: Using Docker

```bash
docker run -it -p 18000:18000 ghcr.io/openflagr/flagr:latest
```

### Option 2: Using Docker Compose

```bash
docker-compose up -d
```

The Flagr UI will be available at: http://localhost:18000

## Examples

### 1. Basic Example

**Location:** `examples/basic/main.go`

Demonstrates fundamental Vexilla usage:
- Creating a client
- Evaluating boolean, string, and integer flags
- Using evaluation contexts
- Accessing detailed evaluation results
- Monitoring cache metrics

**Run:**
```bash
cd examples/basic
go run main.go
```

**What you'll learn:**
- Basic client setup
- Flag evaluation methods
- Context building with attributes
- Reading cache metrics

---

### 2. Microservice Example

**Location:** `examples/microservice/main.go`

Demonstrates Vexilla in a microservice architecture:
- Service-specific flag filtering
- Memory optimization with tags
- Regional feature rollouts
- A/B test evaluation
- Performance monitoring

**Run:**
```bash
cd examples/microservice
go run main.go
```

**What you'll learn:**
- Flag filtering by service tag
- Memory savings calculation
- Environment-based filtering
- Multi-region evaluation

---

## Creating Flags in Flagr

To run the examples, you'll need to create some flags in Flagr. Here's how:

### 1. Access Flagr UI

Open http://localhost:18000 in your browser.

### 2. Create a Simple Boolean Flag

**Flag Key:** `new-checkout`

**Segments:**
- Rollout: 100%
- Constraints:
  - Property: `country`
  - Operator: `EQ`
  - Value: `BR`

**Variants:**
- Variant 1: `enabled`
  - Attachment: `{"enabled": true}`

### 3. Create a String Flag

**Flag Key:** `ui-theme`

**Segments:**
- Rollout: 100%
- Constraints:
  - Property: `tier`
  - Operator: `EQ`
  - Value: `premium`

**Variants:**
- Variant 1: `dark`
  - Attachment: `{"value": "dark"}`
- Variant 2: `light`
  - Attachment: `{"value": "light"}`

### 4. Create an A/B Test Flag

**Flag Key:** `onboarding-ab-test`

**Segments:**
- Rollout: 50% (requires Flagr evaluation)
- No constraints (all users)

**Variants:**
- Variant 1: `control`
  - Attachment: `{"flow_type": "standard", "step_count": 3}`
- Variant 2: `treatment`
  - Attachment: `{"flow_type": "guided", "step_count": 5}`

### 5. Tag Your Flags

For the microservice example, add tags to your flags:

1. Click on the flag
2. Go to "Tags" section
3. Add tags:
   - `user-service` (for user-related flags)
   - `production` (for production flags)

---

## Configuration Options

All examples support configuration via environment variables:

```bash
# Flagr connection
export FLAGR_ENDPOINT="http://localhost:18000"
export FLAGR_API_KEY="your-api-key"  # Optional

# Cache behavior
export REFRESH_INTERVAL="5m"
export INITIAL_TIMEOUT="10s"

# Filtering
export SERVICE_NAME="user-service"
export ONLY_ENABLED="true"
```

---

## Common Scenarios

### Scenario 1: Feature Flag

```go
// Check if feature is enabled for user
ctx := vexilla.NewContext("user-123").
    WithAttribute("country", "BR")

enabled := client.Bool(context.Background(), "new-feature", ctx)
if enabled {
    // Show new feature
}
```

### Scenario 2: Configuration Value

```go
// Get configuration value
ctx := vexilla.NewContext("user-456")

theme := client.String(context.Background(), "ui-theme", ctx, "light")
maxItems := client.Int(context.Background(), "page-size", ctx, 20)
```

### Scenario 3: A/B Test

```go
// Get A/B test variant
ctx := vexilla.NewContext("user-789")

result, _ := client.Evaluate(context.Background(), "ab-test", ctx)
variant := result.VariantKey

if variant == "treatment" {
    // Show treatment version
} else {
    // Show control version
}
```

### Scenario 4: Regional Rollout

```go
// Different behavior per region
regions := []string{"BR", "US", "UK"}

for _, region := range regions {
    ctx := vexilla.NewContext("user").
        WithAttribute("country", region)
    
    enabled := client.Bool(context.Background(), "regional-feature", ctx)
    fmt.Printf("%s: %v\n", region, enabled)
}
```

---

## Performance Tips

### 1. Use Service Filtering in Microservices

```go
client, _ := vexilla.New(
    vexilla.WithFlagrEndpoint("http://localhost:18000"),
    vexilla.WithServiceTag("user-service"),
    vexilla.WithOnlyEnabled(true),
)

// If you have 10,000 flags but only 100 for user-service:
// Memory: ~10MB → ~100KB (99% reduction!)
```

### 2. Choose Appropriate Refresh Intervals

```go
// High-traffic service: longer interval
vexilla.WithRefreshInterval(10 * time.Minute)

// Low-traffic service: shorter interval
vexilla.WithRefreshInterval(1 * time.Minute)
```

### 3. Monitor Cache Performance

```go
metrics := client.Metrics()

if metrics.Storage.HitRatio < 0.8 {
    log.Warn("Cache hit ratio is low: %.2f%%", metrics.Storage.HitRatio*100)
}

if metrics.CircuitOpen {
    log.Error("Circuit breaker is open!")
}
```

---

## Troubleshooting

### Connection Failed

```
Error: failed to start client: connection refused
```

**Solution:** Ensure Flagr is running on http://localhost:18000

```bash
docker ps | grep flagr
```

### Flags Not Found

```
Error: flag not found: my-feature
```

**Solution:** Create the flag in Flagr UI or check the flag key spelling.

### Circuit Breaker Open

```
Error: circuit breaker is open
```

**Solution:** Check Flagr connectivity and restart the client:

```go
client.Stop()
client.Start(ctx)
```

---

## Next Steps

1. **Read the Documentation**
   - [API Reference](https://pkg.go.dev/github.com/OrlandoBitencourt/vexilla)
   - [Architecture Guide](../ARCHITECTURE.md)

2. **Try Advanced Features**
   - Webhook integration for real-time updates
   - OpenTelemetry metrics
   - Circuit breaker configuration

3. **Build Your Application**
   - Start with the basic example
   - Add service filtering for microservices
   - Monitor performance with metrics

---

## Contributing

Found a bug or have an improvement? 

1. Open an issue: https://github.com/OrlandoBitencourt/vexilla/issues
2. Submit a PR with your example
3. Share your use case!

---

**Happy flagging! 🏴**
