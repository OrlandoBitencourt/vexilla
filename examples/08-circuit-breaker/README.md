# Circuit Breaker Example

This example demonstrates **circuit breaker protection** to prevent cascade failures when Flagr is unavailable.

## Problem Without Circuit Breaker

When Flagr is down:
- ❌ Every evaluation waits for timeout (30+ seconds)
- ❌ Cascading failures across the application
- ❌ High latency even when failure is obvious
- ❌ Resource exhaustion (connections, threads)
- ❌ Poor user experience

## Solution: Circuit Breaker

The circuit breaker **fails fast** when Flagr is unavailable:
- ✅ Instant failure (<1ms vs 30s)
- ✅ Prevents cascade failures
- ✅ Automatic recovery testing
- ✅ Protects both application and Flagr

## How It Works

### State Machine

```
┌─────────┐
│ CLOSED  │ ← Normal operation
└────┬────┘
     │
     │ failures >= 3
     ▼
┌─────────┐
│  OPEN   │ ← Failing fast
└────┬────┘
     │
     │ after 30s
     ▼
┌──────────┐
│ HALF-OPEN│ ← Testing recovery
└────┬─────┘
     │
     ├─> Success → CLOSED
     └─> Failure → OPEN
```

### States Explained

**CLOSED** (Normal):
- All requests pass through
- Failures are counted
- Transitions to OPEN after max failures

**OPEN** (Protecting):
- Requests fail immediately
- No calls to Flagr (fail fast)
- Transitions to HALF-OPEN after timeout

**HALF-OPEN** (Testing):
- Limited requests allowed
- Testing if Flagr recovered
- Success → CLOSED, Failure → OPEN

## Running the Example

```bash
cd examples/08-circuit-breaker
go run main.go
```

## Manual Testing

### Test Circuit Opening

1. Start Flagr:
```bash
docker run -d --name flagr -p 18000:18000 checkr/flagr
```

2. Run the example (circuit CLOSED)

3. Stop Flagr:
```bash
docker stop flagr
```

4. Watch circuit OPEN after 3 failures

### Test Circuit Recovery

1. With circuit OPEN, start Flagr:
```bash
docker start flagr
```

2. Watch circuit transition:
   - OPEN → HALF-OPEN (testing)
   - HALF-OPEN → CLOSED (recovered)

## Configuration

### Default Settings

```go
// Default circuit breaker config
Config{
    MaxFailures:     3,                // Open after 3 failures
    Timeout:         30 * time.Second, // Stay open for 30s
    HalfOpenTimeout: 10 * time.Second, // Test recovery after 10s
}
```

### Tuning Guidelines

**High-traffic services:**
```go
// Open quickly, recover quickly
Config{
    MaxFailures:     2,                // Lower threshold
    Timeout:         15 * time.Second, // Shorter timeout
    HalfOpenTimeout: 5 * time.Second,  // Faster recovery
}
```

**Low-traffic services:**
```go
// More conservative
Config{
    MaxFailures:     5,                // Higher threshold
    Timeout:         60 * time.Second, // Longer timeout
    HalfOpenTimeout: 15 * time.Second, // Slower recovery
}
```

## Performance Impact

| Scenario | Without Circuit Breaker | With Circuit Breaker |
|----------|------------------------|---------------------|
| Flagr healthy | 50-200ms | 50-200ms |
| Flagr down (per eval) | 30,000ms (timeout) | <1ms (fail fast) |
| 100 evals when down | 50 minutes | <100ms |

**Impact**: ~30,000x faster failure handling!

## Monitoring

### Key Metrics

Monitor via `client.Metrics()`:

```go
metrics := client.Metrics()

// Circuit state
if metrics.CircuitOpen {
    alert("Circuit breaker OPEN!")
}

// Failure tracking
if metrics.ConsecutiveFails >= 2 {
    warn("Approaching circuit threshold")
}

// Recovery monitoring
lastRefresh := time.Since(metrics.LastRefresh)
if lastRefresh > 10*time.Minute {
    alert("No successful refresh in 10 minutes")
}
```

### Alerting Rules

**Critical:**
- Circuit OPEN for > 5 minutes
- Consecutive failures > max threshold
- No successful refresh in 15 minutes

**Warning:**
- Circuit opened (transient is okay)
- Failures approaching threshold
- Recovery taking longer than expected

## Production Best Practices

### 1. Graceful Degradation

```go
enabled := client.Bool(ctx, "feature", evalCtx)
if circuitOpen() {
    // Use safe default
    enabled = false
}
```

### 2. Fallback Strategies

```go
if metrics.CircuitOpen {
    // Use disk cache
    // Use hardcoded defaults
    // Disable optional features
}
```

### 3. User Communication

```go
if metrics.CircuitOpen {
    log.Warn("Feature flags degraded, using defaults")
    metrics.IncrementDegradedMode()
}
```

### 4. Automatic Recovery Validation

The circuit breaker automatically tests recovery, but you can add additional validation:

```go
// After circuit closes, validate critical flags
if justRecovered {
    validateCriticalFlags()
}
```

## When to Use

**Always use circuit breaker** in production for:
- Protection against Flagr outages
- Preventing cascade failures
- Improving resilience
- Better user experience during failures

The circuit breaker is **enabled by default** in Vexilla - no configuration needed!
