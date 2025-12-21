# Admin Server Example

This example demonstrates the **Admin Server** for operational management of the Vexilla cache.

## Features

The Admin Server provides a REST API for:

- ✅ **Health checks** - Monitor service health
- ✅ **Cache statistics** - View performance metrics
- ✅ **Manual invalidation** - Clear specific flags
- ✅ **Bulk invalidation** - Clear entire cache
- ✅ **Force refresh** - Trigger immediate sync with Flagr

## Running the Example

```bash
cd examples/05-admin-server
go run main.go
```

## API Endpoints

### Health Check

```bash
GET http://localhost:19000/health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

### Cache Statistics

```bash
GET http://localhost:19000/admin/stats
```

**Response:**
```json
{
  "storage": {
    "keys_added": 42,
    "keys_evicted": 3,
    "hit_ratio": 0.95
  },
  "last_refresh": "2025-01-15T10:29:00Z",
  "circuit_open": false,
  "consecutive_fails": 0
}
```

### Invalidate Specific Flag

```bash
POST http://localhost:19000/admin/invalidate
Content-Type: application/json

{
  "flag_key": "new_feature"
}
```

### Invalidate All Flags

```bash
POST http://localhost:19000/admin/invalidate-all
```

### Force Refresh

```bash
POST http://localhost:19000/admin/refresh
```

## Use Cases

### 1. Emergency Flag Update

```bash
# Flag needs immediate update
curl -X POST http://localhost:19000/admin/invalidate \
  -H "Content-Type: application/json" \
  -d '{"flag_key": "critical_feature"}'

# Force refresh from Flagr
curl -X POST http://localhost:19000/admin/refresh
```

### 2. Cache Debugging

```bash
# Check cache health
curl http://localhost:19000/health

# View detailed statistics
curl http://localhost:19000/admin/stats | jq
```

### 3. Deployment Reset

```bash
# Clear cache during deployment
curl -X POST http://localhost:19000/admin/invalidate-all

# Refresh with new configuration
curl -X POST http://localhost:19000/admin/refresh
```

## Security Considerations

⚠️ **Important**: The Admin Server has no authentication in this example. In production:

1. Add authentication middleware
2. Use HTTPS
3. Restrict to internal network
4. Enable audit logging

## Monitoring Integration

Integrate with monitoring tools:

```bash
# Prometheus-style metrics endpoint (future)
curl http://localhost:19000/metrics

# Health check for load balancer
curl http://localhost:19000/health
```
