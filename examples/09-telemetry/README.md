# Telemetry Example

This example demonstrates **OpenTelemetry integration** for comprehensive observability of Vexilla operations.

## Metrics Exported

### Counters

- `vexilla.cache.hits` - Cache hit counter
- `vexilla.cache.misses` - Cache miss counter
- `vexilla.evaluations` - Total evaluations (labeled by strategy)
- `vexilla.refresh.success` - Successful refresh counter
- `vexilla.refresh.failure` - Failed refresh counter

### Histograms

- `vexilla.refresh.duration` - Refresh latency distribution
- `vexilla.evaluation.duration` - Evaluation latency distribution

### Gauges

- `vexilla.circuit.state` - Circuit breaker state (0=closed, 1=open)
- `vexilla.cache.size` - Number of cached flags
- `vexilla.cache.hit_ratio` - Cache hit ratio percentage

## Running the Example

```bash
cd examples/09-telemetry
go run main.go
```

## Built-in Metrics Access

Access metrics programmatically:

```go
metrics := client.Metrics()

// Storage metrics
fmt.Printf("Keys cached: %d\n", metrics.Storage.KeysAdded)
fmt.Printf("Hit ratio: %.2f%%\n", metrics.Storage.HitRatio*100)

// Health metrics
fmt.Printf("Circuit open: %v\n", metrics.CircuitOpen)
fmt.Printf("Last refresh: %v\n", metrics.LastRefresh)
```

## OpenTelemetry Integration

### Setup with Prometheus

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/prometheus"
)

// Create Prometheus exporter
exporter, _ := prometheus.New()

// Create meter provider
provider := metric.NewMeterProvider(
    metric.WithReader(exporter),
)
otel.SetMeterProvider(provider)

// Vexilla will automatically use the global meter provider
client, _ := vexilla.New(
    vexilla.WithFlagrEndpoint("http://localhost:18000"),
    vexilla.WithTelemetry(true), // Enable OTel
)
```

### Expose Metrics Endpoint

```go
http.Handle("/metrics", promhttp.Handler())
http.ListenAndServe(":9090", nil)
```

### Grafana Dashboard

Example queries for Grafana:

```promql
# Cache hit ratio
vexilla_cache_hit_ratio

# Evaluation rate
rate(vexilla_evaluations_total[5m])

# Refresh duration p95
histogram_quantile(0.95, vexilla_refresh_duration_bucket)

# Circuit breaker state
vexilla_circuit_state
```

## Recommended Dashboards

### 1. Cache Performance

Panels:
- Cache hit ratio (gauge)
- Cache size trend (line chart)
- Hit/miss rates (stacked area)
- Eviction rate (line chart)

### 2. Evaluation Metrics

Panels:
- Total evaluations per second (counter)
- Local vs remote distribution (pie chart)
- Latency percentiles (heatmap)
- Throughput (line chart)

### 3. Reliability

Panels:
- Circuit breaker state (status panel)
- Refresh success rate (gauge)
- Consecutive failures (number)
- Time since last refresh (stat)

### 4. Resource Usage

Panels:
- Memory usage estimate (gauge)
- Network requests to Flagr (counter)
- Disk I/O operations (counter)

## Alerting Rules

### Critical Alerts

```yaml
# Circuit breaker open for too long
- alert: VexillaCircuitOpen
  expr: vexilla_circuit_state == 1
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Vexilla circuit breaker has been open for 5+ minutes"

# Low cache hit ratio
- alert: VexillaLowHitRatio
  expr: vexilla_cache_hit_ratio < 50
  for: 10m
  labels:
    severity: critical
  annotations:
    summary: "Vexilla cache hit ratio below 50%"

# No successful refresh
- alert: VexillaRefreshStale
  expr: time() - vexilla_last_refresh_timestamp > 900
  labels:
    severity: critical
  annotations:
    summary: "No successful Vexilla refresh in 15 minutes"
```

### Warning Alerts

```yaml
# Circuit opened (transient)
- alert: VexillaCircuitOpened
  expr: vexilla_circuit_state == 1
  for: 1m
  labels:
    severity: warning

# Suboptimal hit ratio
- alert: VexillaSuboptimalHitRatio
  expr: vexilla_cache_hit_ratio < 85
  for: 10m
  labels:
    severity: warning

# Slow refresh
- alert: VexillaSlowRefresh
  expr: histogram_quantile(0.95, vexilla_refresh_duration_bucket) > 5
  labels:
    severity: warning
```

## Key Metrics to Monitor

### Cache Efficiency

- **Hit Ratio**: >95% is excellent, <85% needs investigation
- **Eviction Rate**: High eviction may indicate undersized cache
- **Cache Size**: Should be stable after warmup

### Performance

- **Evaluation Latency**: <1ms for local, 50-200ms for remote
- **Throughput**: Should scale linearly with requests
- **p99 Latency**: Monitor for outliers

### Reliability

- **Circuit State**: Should be closed >99.9% of time
- **Refresh Success Rate**: Target 100%
- **Consecutive Failures**: Should be 0 in steady state

## Integration Examples

### Prometheus

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'vexilla'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: '/metrics'
```

### Grafana

Import the Vexilla dashboard:
1. Create new dashboard
2. Add panels using PromQL queries above
3. Set appropriate refresh intervals
4. Configure alerts

### Datadog

```go
import "github.com/DataDog/datadog-go/statsd"

// Custom metrics export
statsd, _ := statsd.New("127.0.0.1:8125")
defer statsd.Close()

// Export Vexilla metrics
metrics := client.Metrics()
statsd.Gauge("vexilla.cache.hit_ratio", metrics.Storage.HitRatio*100, nil, 1)
statsd.Count("vexilla.cache.keys", metrics.Storage.KeysAdded, nil, 1)
```

## Production Best Practices

1. **Always monitor in production**
2. **Set up critical alerts** (circuit breaker, hit ratio)
3. **Track trends** over time to detect degradation
4. **Correlate with application metrics** for full picture
5. **Review dashboards** during incidents

## Performance Benchmarks

Expected performance characteristics:

```
Local Evaluation:
  - Latency: <1ms (p99)
  - Throughput: 200,000+ eval/sec
  - Cache hit ratio: >95%

Remote Evaluation:
  - Latency: 50-200ms (p99)
  - Throughput: Limited by Flagr capacity
  - Network requests: 1 per eval

Background Refresh:
  - Frequency: Every 5 minutes
  - Duration: <5 seconds for 1000 flags
  - Success rate: 100%
```
