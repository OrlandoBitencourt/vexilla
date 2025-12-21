# HTTP Middleware Example

This example demonstrates how to integrate Vexilla into HTTP APIs for **automatic feature flag evaluation** based on request context.

## Features

- ✅ **Automatic context extraction** from HTTP headers
- ✅ **Feature gating** for API endpoints
- ✅ **A/B testing** for different user segments
- ✅ **Regional features** based on user location
- ✅ **Premium features** gating by subscription tier

## Running the Example

```bash
cd examples/06-http-middleware
go run main.go
```

The server will start on `http://localhost:8080`.

## API Endpoints

### 1. Dashboard Endpoint

Returns appropriate dashboard URL based on user tier:

```bash
curl -H 'X-User-ID: user-123' \
     -H 'X-User-Tier: premium' \
     http://localhost:8080/api/dashboard
```

**Response:**
```json
{
  "user_id": "user-123",
  "tier": "premium",
  "premium_features": true,
  "dashboard_url": "/dashboard/premium"
}
```

### 2. Pricing Endpoint (A/B Testing)

Returns different pricing based on A/B test variant:

```bash
curl -H 'X-User-ID: user-456' \
     -H 'X-Country: BR' \
     http://localhost:8080/api/pricing
```

**Response:**
```json
{
  "user_id": "user-456",
  "country": "BR",
  "layout": "discount",
  "prices": {
    "basic": "$7/month",
    "pro": "$24/month",
    "enterprise": "$79/month"
  }
}
```

### 3. Beta Features Endpoint

Restricted endpoint requiring beta access:

```bash
curl -H 'X-User-ID: user-789' \
     -H 'X-User-Tier: premium' \
     -H 'X-Signup-Date: 2025-01-01' \
     http://localhost:8080/api/beta-features
```

**Response:**
```json
{
  "user_id": "user-789",
  "beta_access": true,
  "beta_features": [
    "advanced_analytics",
    "custom_themes",
    "api_access",
    "priority_support"
  ]
}
```

### 4. Features Endpoint

Returns all enabled features for the user:

```bash
curl -H 'X-User-ID: user-999' \
     -H 'X-Country: US' \
     http://localhost:8080/api/features
```

**Response:**
```json
{
  "user_id": "user-999",
  "country": "US",
  "features": {
    "dark_mode": true,
    "new_ui": false,
    "brazil_launch": false,
    "payment_pix": false,
    "premium_features": false
  }
}
```

### 5. Health Check

Monitor service health:

```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "status": "healthy",
  "circuit_open": false,
  "last_refresh": "2025-01-15T10:30:00Z",
  "keys_cached": 25,
  "hit_ratio": 0.95
}
```

## Integration Patterns

### Pattern 1: Feature Gating

```go
hasPremium := client.Bool(r.Context(), "premium_features", evalCtx)

if !hasPremium {
    http.Error(w, "Premium required", http.StatusForbidden)
    return
}

// Premium feature code here
```

### Pattern 2: A/B Testing

```go
result, _ := client.Evaluate(r.Context(), "pricing_test", evalCtx)
variant := result.GetString("variant", "control")

switch variant {
case "variant_a":
    // Show pricing variant A
case "variant_b":
    // Show pricing variant B
default:
    // Show control pricing
}
```

### Pattern 3: Regional Features

```go
evalCtx := vexilla.NewContext(userID).
    WithAttribute("country", country)

hasBrazilFeature := client.Bool(r.Context(), "brazil_launch", evalCtx)
```

## Benefits

- ✅ **Clean code**: No scattered if/else feature flag checks
- ✅ **Type-safe**: Compile-time checks for flag keys
- ✅ **Fast**: <1ms evaluation for local flags
- ✅ **Testable**: Easy to mock for testing
- ✅ **Observable**: Built-in metrics and health checks

## Production Considerations

1. **Caching**: Use Vexilla's built-in cache (Ristretto)
2. **Headers**: Standardize header names across services
3. **Fallbacks**: Handle missing headers gracefully
4. **Monitoring**: Monitor hit ratios and circuit breaker state
5. **Security**: Validate and sanitize header values
