package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// Metrics holds all OpenTelemetry metrics
type Metrics struct {
	CacheHits      metric.Int64Counter
	CacheMisses    metric.Int64Counter
	RefreshSuccess metric.Int64Counter
	RefreshFailure metric.Int64Counter
	RefreshLatency metric.Float64Histogram
	CacheSize      metric.Int64ObservableGauge
	WebhookEvents  metric.Int64Counter
	LocalEvals     metric.Int64Counter
	RemoteEvals    metric.Int64Counter
}

// NewMetrics creates a new metrics instance
func NewMetrics(meterName string) (*Metrics, error) {
	meter := otel.Meter(meterName)

	cacheHits, err := meter.Int64Counter(
		"vexilla.cache.hits",
		metric.WithDescription("Number of cache hits"),
		metric.WithUnit("{hit}"),
	)
	if err != nil {
		return nil, err
	}

	cacheMisses, err := meter.Int64Counter(
		"vexilla.cache.misses",
		metric.WithDescription("Number of cache misses"),
		metric.WithUnit("{miss}"),
	)
	if err != nil {
		return nil, err
	}

	refreshSuccess, err := meter.Int64Counter(
		"vexilla.refresh.success",
		metric.WithDescription("Number of successful refresh operations"),
		metric.WithUnit("{refresh}"),
	)
	if err != nil {
		return nil, err
	}

	refreshFailure, err := meter.Int64Counter(
		"vexilla.refresh.failure",
		metric.WithDescription("Number of failed refresh operations"),
		metric.WithUnit("{refresh}"),
	)
	if err != nil {
		return nil, err
	}

	refreshLatency, err := meter.Float64Histogram(
		"vexilla.refresh.duration",
		metric.WithDescription("Duration of refresh operations"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, err
	}

	webhookEvents, err := meter.Int64Counter(
		"vexilla.webhook.events",
		metric.WithDescription("Number of webhook events received"),
		metric.WithUnit("{event}"),
	)
	if err != nil {
		return nil, err
	}

	localEvals, err := meter.Int64Counter(
		"vexilla.evaluation.local",
		metric.WithDescription("Number of local evaluations"),
		metric.WithUnit("{evaluation}"),
	)
	if err != nil {
		return nil, err
	}

	remoteEvals, err := meter.Int64Counter(
		"vexilla.evaluation.remote",
		metric.WithDescription("Number of remote evaluations"),
		metric.WithUnit("{evaluation}"),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		CacheHits:      cacheHits,
		CacheMisses:    cacheMisses,
		RefreshSuccess: refreshSuccess,
		RefreshFailure: refreshFailure,
		RefreshLatency: refreshLatency,
		WebhookEvents:  webhookEvents,
		LocalEvals:     localEvals,
		RemoteEvals:    remoteEvals,
	}, nil
}

// SetCacheSizeGauge sets up the cache size observable gauge
func (m *Metrics) SetCacheSizeGauge(meterName string, callback func(context.Context, metric.Int64Observer) error) error {
	meter := otel.Meter(meterName)

	gauge, err := meter.Int64ObservableGauge(
		"vexilla.cache.size",
		metric.WithDescription("Number of flags in cache"),
		metric.WithUnit("{flag}"),
		metric.WithInt64Callback(callback),
	)
	if err != nil {
		return err
	}

	m.CacheSize = gauge
	return nil
}
