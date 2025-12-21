# Advanced Filtering Example

This example demonstrates **advanced flag filtering** to reduce memory usage by 50-95% in microservice architectures.

## Problem

In large-scale deployments:
- 10,000+ feature flags in Flagr
- Each microservice only needs ~5% of flags
- Caching all flags wastes 95% of memory
- Slower cache operations on large datasets

## Solution: Smart Filtering

Filter flags at cache load time based on:
- ✅ Enabled/disabled state
- ✅ Service tags
- ✅ Environment tags
- ✅ Custom metadata

## Running the Example

```bash
cd examples/10-advanced-filtering
go run main.go
```

## Filtering Strategies

### 1. Only Enabled Flags

Cache only enabled flags, skip disabled ones:

```go
vexilla.WithOnlyEnabled(true)
```

**Savings**: 20-40% (assuming 20-40% flags are disabled)

### 2. Service-Specific Tags

Each microservice caches only its flags:

```go
vexilla.WithServiceTag("user-service", true)
```

**Savings**: 90-95% in microservice architectures

**Setup**: Tag flags in Flagr with service names:
- `user-service`
- `payment-service`
- `notification-service`

### 3. Additional Tags (ANY mode)

Cache flags matching ANY of the specified tags:

```go
vexilla.WithAdditionalTags([]string{"user", "premium"}, "any")
```

Matches flags with:
- Tag: `user` OR
- Tag: `premium`

### 4. Additional Tags (ALL mode)

Cache flags matching ALL of the specified tags:

```go
vexilla.WithAdditionalTags([]string{"user", "premium"}, "all")
```

Matches flags with:
- Tag: `user` AND
- Tag: `premium`

### 5. Combined Filtering

Stack multiple filters for maximum optimization:

```go
client, _ := vexilla.New(
    vexilla.WithOnlyEnabled(true),                      // Filter disabled
    vexilla.WithServiceTag("user-service", true),       // Service-specific
    vexilla.WithAdditionalTags([]string{"production"}, "all"), // Environment
)
```

**Savings**: 95-97% in production microservices

## Memory Savings Calculation

### Example: 10,000 flags, 1KB per flag

| Strategy | Flags Cached | Memory Used | Saved | % Saved |
|----------|--------------|-------------|-------|---------|
| No filtering | 10,000 | 9.77 MB | 0 MB | 0% |
| OnlyEnabled | 7,000 | 6.84 MB | 2.93 MB | 30% |
| Service tags | 500 | 0.49 MB | 9.28 MB | 95% |
| Combined | 300 | 0.29 MB | 9.48 MB | 97% |

## Real-World Example

### E-commerce Platform

**Architecture**: 8 microservices, 10,000 total flags

#### Without Filtering
- Each service caches: 10,000 flags
- Total memory: 8 × 10,000 × 1KB = **78.13 MB**

#### With Filtering
- api-gateway: 300 flags = 0.29 MB
- user-service: 450 flags = 0.44 MB
- auth-service: 200 flags = 0.20 MB
- product-service: 600 flags = 0.59 MB
- cart-service: 400 flags = 0.39 MB
- payment-service: 350 flags = 0.34 MB
- notification-service: 250 flags = 0.24 MB
- analytics-service: 500 flags = 0.49 MB

**Total memory**: **2.98 MB**

**Savings**: **75.15 MB (96.2% reduction)**

## Tagging Strategy

### In Flagr UI

Tag your flags appropriately:

```
Flag: "user_registration"
Tags: ["user-service", "registration", "production"]

Flag: "payment_gateway_v2"
Tags: ["payment-service", "gateway", "production"]

Flag: "beta_dark_mode"
Tags: ["user-service", "ui", "beta"]
```

### Naming Conventions

Recommended tag structure:

1. **Service tags**: `<service-name>`
   - Example: `user-service`, `payment-service`

2. **Environment tags**: `production`, `staging`, `development`

3. **Feature tags**: `beta`, `experimental`, `deprecated`

4. **Team tags**: `team-a`, `team-b`

## Configuration Examples

### Monolithic Application

```go
client, _ := vexilla.New(
    vexilla.WithFlagrEndpoint("http://localhost:18000"),
    vexilla.WithOnlyEnabled(true), // 20-40% savings
)
```

### Microservice

```go
client, _ := vexilla.New(
    vexilla.WithFlagrEndpoint("http://localhost:18000"),
    vexilla.WithOnlyEnabled(true),
    vexilla.WithServiceTag("user-service", true), // 90-95% savings
)
```

### Multi-tenant Service

```go
client, _ := vexilla.New(
    vexilla.WithFlagrEndpoint("http://localhost:18000"),
    vexilla.WithOnlyEnabled(true),
    vexilla.WithAdditionalTags([]string{"tenant-acme"}, "any"), // 50-80% savings
)
```

### Environment-specific

```go
client, _ := vexilla.New(
    vexilla.WithFlagrEndpoint("http://localhost:18000"),
    vexilla.WithOnlyEnabled(true),
    vexilla.WithAdditionalTags([]string{"production"}, "all"), // 30-50% savings
)
```

## Validation

### Verify Filtering Effectiveness

```go
metrics := client.Metrics()

// Check how many flags were cached
flagsCached := metrics.Storage.KeysAdded

// Calculate savings
totalFlags := 10000
percentCached := float64(flagsCached) / float64(totalFlags) * 100
percentSaved := 100 - percentCached

fmt.Printf("Flags cached: %d / %d (%.1f%%)\n", flagsCached, totalFlags, percentCached)
fmt.Printf("Memory saved: %.1f%%\n", percentSaved)
```

### Expected Results

| Architecture | Flags Cached | % Saved |
|--------------|--------------|---------|
| Monolithic | 60-80% | 20-40% |
| Microservice | 3-10% | 90-97% |
| Multi-tenant | 20-50% | 50-80% |

## Best Practices

### 1. Tag Organization

- **Consistent naming**: Use lowercase, hyphen-separated
- **Hierarchical**: service-name, team-name, environment
- **Comprehensive**: Tag all flags appropriately

### 2. Filter Configuration

- **Start conservative**: Begin with OnlyEnabled only
- **Measure**: Check metrics before adding filters
- **Iterate**: Add service/environment tags gradually
- **Validate**: Ensure critical flags are cached

### 3. Monitoring

Monitor these metrics:

```go
metrics := client.Metrics()

// Flags cached
if metrics.Storage.KeysAdded < expectedMinimum {
    log.Warn("Too few flags cached, check filters")
}

// Hit ratio
if metrics.Storage.HitRatio < 0.95 {
    log.Warn("Low hit ratio, may need more flags cached")
}
```

### 4. Testing

Test filtering in staging first:

```go
// Staging: More permissive
vexilla.WithAdditionalTags([]string{"staging", "production"}, "any")

// Production: Strict
vexilla.WithAdditionalTags([]string{"production"}, "all")
```

## Troubleshooting

### Problem: Flags Missing

**Symptom**: Expected flag not available

**Solutions**:
1. Check flag tags in Flagr
2. Verify filter configuration
3. Use less restrictive filters temporarily
4. Check logs for filtering statistics

### Problem: Too Many Flags Cached

**Symptom**: High memory usage despite filtering

**Solutions**:
1. Use more restrictive service tags
2. Add environment tag filtering
3. Review tag assignment in Flagr
4. Check for wildcard matches

### Problem: Low Hit Ratio

**Symptom**: Hit ratio <90%

**Solutions**:
1. Flags may be missing due to over-filtering
2. Check if filters are too restrictive
3. Review which flags are being requested
4. Adjust filtering strategy

## Performance Impact

Filtering has minimal overhead:

- **Load time**: +10-50ms (one-time cost)
- **Memory**: Reduced by 50-97%
- **Cache operations**: 2-10x faster (smaller dataset)
- **Refresh time**: 50-95% faster

**Net result**: Faster overall performance with lower memory usage!
