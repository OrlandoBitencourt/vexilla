# Webhook Invalidation Example

This example demonstrates **real-time cache invalidation** using webhooks, enabling instant flag updates without waiting for background refresh.

## How It Works

```
Flagr UI: Flag Updated
    │
    ▼
POST http://localhost:18001/webhook
{
  "event": "flag.updated",
  "flag_keys": ["feature_x"],
  "timestamp": "2025-01-15T10:30:00Z"
}
    │
    ▼
Vexilla Webhook Server
    │
    ├─> Verify HMAC signature
    ├─> Invalidate cache for flag_keys
    └─> Trigger immediate refresh
    │
    ▼
Updated flag available instantly!
```

## Benefits

- ✅ **Sub-second updates**: Flags update in <1s instead of minutes
- ✅ **Event-driven**: No polling overhead
- ✅ **Secure**: HMAC-SHA256 signature verification
- ✅ **Selective invalidation**: Only refreshes changed flags

## Running the Example

### 1. Start Flagr (if not running)

```bash
docker run -d --name flagr -p 18000:18000 checkr/flagr
```

### 2. Configure Flagr Webhooks

Edit Flagr's configuration to send webhooks to Vexilla:

```yaml
webhooks:
  - url: http://localhost:18001/webhook
    secret: my-webhook-secret-key
    events:
      - flag.updated
      - flag.deleted
```

### 3. Run the Example

```bash
cd examples/04-webhook-invalidation
go run main.go
```

### 4. Test It!

1. Open Flagr UI: http://localhost:18000
2. Modify a flag (enable/disable or change constraints)
3. Watch the console output - you'll see the change detected instantly!

## Security

The webhook server uses **HMAC-SHA256** signature verification to ensure requests are authentic:

```go
vexilla.WithWebhookInvalidation(vexilla.WebhookConfig{
    Port:   18001,
    Secret: "my-webhook-secret-key", // Must match Flagr config
})
```

## Events Handled

- `flag.updated` - Invalidates and refreshes the flag
- `flag.deleted` - Removes flag from cache

## Performance Impact

| Refresh Method | Update Latency | Network Overhead |
|----------------|----------------|------------------|
| Background refresh (5min) | Up to 5 minutes | Periodic polling |
| Webhook invalidation | <1 second | Event-driven only |

**Result**: Updates propagate ~300x faster!
