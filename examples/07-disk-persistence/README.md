# Disk Persistence Example

This example demonstrates **disk-based cache persistence** for fast startup and recovery scenarios.

## Benefits

- ✅ **Fast startup**: Load flags from disk instead of fetching from Flagr
- ✅ **Resilience**: Works when Flagr is temporarily unavailable
- ✅ **Recovery**: Last-known-good state preserved across restarts
- ✅ **No data loss**: Survives application crashes and restarts

## How It Works

```
First Startup (Cold Cache):
    │
    ▼
Fetch from Flagr API
    │
    ├─> Save to memory (Ristretto)
    └─> Save to disk (JSON files)

Second Startup (Warm Cache):
    │
    ▼
Load from disk
    │
    ├─> Populate memory cache
    ├─> Refresh from Flagr (async)
    └─> Ready in <100ms!
```

## Running the Example

```bash
cd examples/07-disk-persistence
go run main.go
```

## Performance Comparison

| Startup Type | Time | Source |
|--------------|------|--------|
| Cold start | 500-2000ms | Flagr API fetch |
| Warm start | 50-100ms | Disk cache load |

**Speedup**: 10-20x faster startup!

## Cache Files

Flags are stored as individual JSON files:

```bash
# Default location (customize with WithDiskCache)
/tmp/vexilla-cache/

# Structure
/tmp/vexilla-cache/
  ├── new_feature.json
  ├── premium_features.json
  └── dark_mode.json
```

## Configuration

### Enable Disk Persistence

```go
client, err := vexilla.New(
    vexilla.WithFlagrEndpoint("http://localhost:18000"),
    vexilla.WithDiskCache("/path/to/cache"), // Enable disk persistence
    vexilla.WithRefreshInterval(5*time.Minute),
)
```

### Customize Cache Directory

```go
// Production example
cacheDir := filepath.Join("/var/cache/myapp", "vexilla")
vexilla.WithDiskCache(cacheDir)
```

## Recovery Scenarios

### Scenario 1: Application Restart

```
Application crashes → Restarts
    │
    ▼
Load flags from disk (instant)
    │
    └─> Continue serving requests
```

### Scenario 2: Flagr Temporarily Down

```
Flagr unavailable
    │
    ▼
Use disk cache (last-known-good)
    │
    ├─> Serve requests with cached flags
    └─> Retry Flagr connection in background
```

### Scenario 3: Network Issues

```
Network partition
    │
    ▼
Cannot reach Flagr
    │
    └─> Fall back to disk cache
        │
        └─> Graceful degradation
```

## Cache Management

### View Cache Contents

```bash
# List cache files
ls -lh /tmp/vexilla-cache/

# Inspect a flag
cat /tmp/vexilla-cache/new_feature.json | jq
```

### Clear Cache

```bash
# Clear all cached flags
rm -rf /tmp/vexilla-cache/

# Clear specific flag
rm /tmp/vexilla-cache/new_feature.json
```

### Monitor Cache Size

```bash
# Check disk usage
du -sh /tmp/vexilla-cache/

# Count files
ls /tmp/vexilla-cache/ | wc -l
```

## Production Considerations

### 1. Cache Location

```go
// Development
cacheDir := filepath.Join(os.TempDir(), "vexilla-cache")

// Production (Linux)
cacheDir := "/var/cache/myapp/vexilla"

// Production (with app name)
cacheDir := filepath.Join("/var/cache", appName, "vexilla")
```

### 2. Permissions

```bash
# Ensure write permissions
mkdir -p /var/cache/myapp/vexilla
chown myapp:myapp /var/cache/myapp/vexilla
chmod 755 /var/cache/myapp/vexilla
```

### 3. Disk Space Monitoring

```go
// Monitor cache size
cacheSize := calculateCacheSize(cacheDir)
if cacheSize > maxCacheSize {
    // Alert or cleanup
}
```

### 4. Stale Data

Disk cache uses TTL and background refresh to prevent stale data:

- TTL: 5 minutes default
- Background refresh: Automatic
- Manual refresh: Admin API endpoint

## Trade-offs

### Pros

- ✅ Fast startup
- ✅ Offline capability
- ✅ Recovery from failures

### Cons

- ⚠️ Disk I/O overhead (async, minimal)
- ⚠️ Stale data risk (mitigated by TTL)
- ⚠️ Disk space usage (~1KB per flag)

## When to Use

Use disk persistence when:

- Fast startup is critical
- You need resilience during Flagr outages
- You want graceful degradation
- Your application frequently restarts

Skip it when:

- Disk I/O is severely constrained
- You always have reliable Flagr connectivity
- Fresh data is absolutely critical
