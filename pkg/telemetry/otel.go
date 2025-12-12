package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	meterName  = "vexilla"
	tracerName = "vexilla"
)

// OTelProvider implements Provider using OpenTelemetry
type OTelProvider struct {
	tracer trace.Tracer
	meter  metric.Meter

	// Metrics
	cacheHits       metric.Int64Counter
	cacheMisses     metric.Int64Counter
	evaluations     metric.Int64Counter
	refreshDuration metric.Float64Histogram
	refreshSuccess  metric.Int64Counter
	refreshFailure  metric.Int64Counter
	circuitState    metric.Int64ObservableGauge

	// Current circuit state (for gauge)
	currentCircuitState string
}

// NewOTel creates a new OpenTelemetry provider
func NewOTel() (*OTelProvider, error) {
	tracer := otel.Tracer(tracerName)
	meter := otel.Meter(meterName)

	provider := &OTelProvider{
		tracer: tracer,
		meter:  meter,
	}

	if err := provider.initMetrics(); err != nil {
		return nil, err
	}

	return provider, nil
}

// initMetrics initializes all metrics
func (o *OTelProvider) initMetrics() error {
	var err error

	// Cache metrics
	o.cacheHits, err = o.meter.Int64Counter(
		"vexilla.cache.hits",
		metric.WithDescription("Number of cache hits"),
	)
	if err != nil {
		return err
	}

	o.cacheMisses, err = o.meter.Int64Counter(
		"vexilla.cache.misses",
		metric.WithDescription("Number of cache misses"),
	)
	if err != nil {
		return err
	}

	// Evaluation metrics
	o.evaluations, err = o.meter.Int64Counter(
		"vexilla.evaluations",
		metric.WithDescription("Number of flag evaluations"),
	)
	if err != nil {
		return err
	}

	// Refresh metrics
	o.refreshDuration, err = o.meter.Float64Histogram(
		"vexilla.refresh.duration",
		metric.WithDescription("Duration of cache refresh operations"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}

	o.refreshSuccess, err = o.meter.Int64Counter(
		"vexilla.refresh.success",
		metric.WithDescription("Number of successful refreshes"),
	)
	if err != nil {
		return err
	}

	o.refreshFailure, err = o.meter.Int64Counter(
		"vexilla.refresh.failure",
		metric.WithDescription("Number of failed refreshes"),
	)
	if err != nil {
		return err
	}

	// Circuit breaker gauge
	o.circuitState, err = o.meter.Int64ObservableGauge(
		"vexilla.circuit.state",
		metric.WithDescription("Circuit breaker state (0=closed, 1=open, 2=half-open)"),
		metric.WithInt64Callback(func(ctx context.Context, observer metric.Int64Observer) error {
			state := o.getCircuitStateValue()
			observer.Observe(state)
			return nil
		}),
	)
	if err != nil {
		return err
	}

	return nil
}

// getCircuitStateValue converts circuit state string to numeric value
func (o *OTelProvider) getCircuitStateValue() int64 {
	switch o.currentCircuitState {
	case "closed":
		return 0
	case "open":
		return 1
	case "half-open":
		return 2
	default:
		return 0
	}
}

// StartSpan creates a new trace span
func (o *OTelProvider) StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	config := &SpanConfig{}
	for _, opt := range opts {
		opt(config)
	}

	// Convert our attributes to OTel attributes
	otelAttrs := make([]attribute.KeyValue, len(config.Attributes))
	for i, attr := range config.Attributes {
		otelAttrs[i] = o.convertAttribute(attr)
	}

	ctx, otelSpan := o.tracer.Start(ctx, name,
		trace.WithAttributes(otelAttrs...))

	return ctx, &OTelSpan{span: otelSpan, provider: o}
}

// convertAttribute converts our Attribute to OTel attribute
func (o *OTelProvider) convertAttribute(attr Attribute) attribute.KeyValue {
	switch v := attr.Value.(type) {
	case string:
		return attribute.String(attr.Key, v)
	case int:
		return attribute.Int(attr.Key, v)
	case int64:
		return attribute.Int64(attr.Key, v)
	case bool:
		return attribute.Bool(attr.Key, v)
	case float64:
		return attribute.Float64(attr.Key, v)
	default:
		return attribute.String(attr.Key, "")
	}
}

// RecordCacheHit records a cache hit
func (o *OTelProvider) RecordCacheHit(ctx context.Context, flagKey string) {
	o.cacheHits.Add(ctx, 1, metric.WithAttributes(
		attribute.String("flag.key", flagKey),
	))
}

// RecordCacheMiss records a cache miss
func (o *OTelProvider) RecordCacheMiss(ctx context.Context, flagKey string) {
	o.cacheMisses.Add(ctx, 1, metric.WithAttributes(
		attribute.String("flag.key", flagKey),
	))
}

// RecordEvaluation records a flag evaluation
func (o *OTelProvider) RecordEvaluation(ctx context.Context, flagKey string, strategy string, duration time.Duration) {
	o.evaluations.Add(ctx, 1, metric.WithAttributes(
		attribute.String("flag.key", flagKey),
		attribute.String("strategy", strategy),
	))
}

// RecordRefresh records a cache refresh operation
func (o *OTelProvider) RecordRefresh(ctx context.Context, success bool, duration time.Duration, flagCount int) {
	// Record duration
	o.refreshDuration.Record(ctx, float64(duration.Milliseconds()),
		metric.WithAttributes(
			attribute.Bool("success", success),
		))

	// Record success/failure
	if success {
		o.refreshSuccess.Add(ctx, 1, metric.WithAttributes(
			attribute.Int("flag.count", flagCount),
		))
	} else {
		o.refreshFailure.Add(ctx, 1)
	}
}

// RecordCircuitState records the circuit breaker state
func (o *OTelProvider) RecordCircuitState(ctx context.Context, state string) {
	o.currentCircuitState = state
}

// Shutdown shuts down the provider
func (o *OTelProvider) Shutdown(ctx context.Context) error {
	// OTel SDK shutdown is handled globally
	return nil
}

// OTelSpan wraps an OpenTelemetry span
type OTelSpan struct {
	span     trace.Span
	provider *OTelProvider
}

// End completes the span
func (s *OTelSpan) End() {
	s.span.End()
}

// SetAttributes sets attributes on the span
func (s *OTelSpan) SetAttributes(attrs ...Attribute) {
	otelAttrs := make([]attribute.KeyValue, len(attrs))
	for i, attr := range attrs {
		otelAttrs[i] = s.provider.convertAttribute(attr)
	}
	s.span.SetAttributes(otelAttrs...)
}

// RecordError records an error on the span
func (s *OTelSpan) RecordError(err error) {
	s.span.RecordError(err)
}

// AddEvent adds an event to the span
func (s *OTelSpan) AddEvent(name string, attrs ...Attribute) {
	otelAttrs := make([]attribute.KeyValue, len(attrs))
	for i, attr := range attrs {
		otelAttrs[i] = s.provider.convertAttribute(attr)
	}
	s.span.AddEvent(name, trace.WithAttributes(otelAttrs...))
}
