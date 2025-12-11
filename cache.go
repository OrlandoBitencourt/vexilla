package vexilla

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	meterName  = "vexilla"
	tracerName = "vexilla"
)

// Flag represents a feature flag with its evaluation rules
type Flag struct {
	ID          int           `json:"id"`
	Key         string        `json:"key"`
	Default     interface{}   `json:"default"`
	Rules       []FlagRule    `json:"rules"`
	Segments    []Segment     `json:"segments"`
	LastUpdated time.Time     `json:"last_updated"`
	CacheExpiry time.Time     `json:"cache_expiry"`
	TTL         time.Duration `json:"ttl,omitempty"`
	IsCritical  bool          `json:"is_critical,omitempty"`

	// Evaluation strategy
	EvaluationType string `json:"evaluation_type"` // "static", "dynamic", "auto"
}

// Segment represents a flag segment with constraints and rollout
type Segment struct {
	ID             string                `json:"id"`
	RolloutPercent int                   `json:"rollout_percent"` // 0-100
	Constraints    []Constraint          `json:"constraints"`
	Distributions  []VariantDistribution `json:"distributions"`
}

// Constraint represents a segment constraint (e.g., country == "US")
type Constraint struct {
	Property string      `json:"property"`
	Operator string      `json:"operator"` // "EQ", "NEQ", "IN", "NOTIN", "MATCHES", etc.
	Value    interface{} `json:"value"`
}

// VariantDistribution represents percentage-based variant distribution
type VariantDistribution struct {
	VariantID  string      `json:"variant_id"`
	VariantKey string      `json:"variant_key"`
	Percentage int         `json:"percentage"` // 0-100
	Attachment interface{} `json:"attachment,omitempty"`
}

// FlagRule represents a single evaluation rule
type FlagRule struct {
	Condition string      `json:"condition"`
	Value     interface{} `json:"value"`
}

// EvaluationContext holds user/request context for flag evaluation
type EvaluationContext struct {
	UserID     string                 `json:"user_id"`
	Attributes map[string]interface{} `json:"attributes"`
}

// Cache is the main feature flag cache manager
type Cache struct {
	config     Config
	cache      *ristretto.Cache
	httpClient *http.Client

	// OpenTelemetry
	tracer         trace.Tracer
	meter          metric.Meter
	cacheHits      metric.Int64Counter
	cacheMisses    metric.Int64Counter
	refreshSuccess metric.Int64Counter
	refreshFailure metric.Int64Counter
	cacheSize      metric.Int64ObservableGauge
	refreshLatency metric.Float64Histogram
	webhookEvents  metric.Int64Counter

	// State management
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	mu               sync.RWMutex
	lastRefresh      time.Time
	consecutiveFails int
	circuitOpen      bool

	// Servers
	webhookServer *http.Server
	adminServer   *http.Server

	// Evaluator
	evaluator *Evaluator
}

// New creates a new feature flag cache instance
func New(config Config) (*Cache, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Initialize Ristretto cache
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: config.CacheNumCounters,
		MaxCost:     config.CacheMaxCost,
		BufferItems: config.CacheBufferItems,
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create ristretto cache: %w", err)
	}

	// Initialize OpenTelemetry
	tracer := otel.Tracer(tracerName)
	meter := otel.Meter(meterName)

	cache := &Cache{
		config: config,
		cache:  ristrettoCache,
		httpClient: &http.Client{
			Timeout: config.HTTPTimeout,
		},
		ctx:       ctx,
		cancel:    cancel,
		tracer:    tracer,
		meter:     meter,
		evaluator: NewEvaluator(),
	}

	// Initialize metrics
	if err := cache.initMetrics(); err != nil {
		cancel()
		ristrettoCache.Close()
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	return cache, nil
}

// Start begins the background refresh process and optional servers
func (c *Cache) Start() error {
	ctx, span := c.tracer.Start(c.ctx, "cache.start")
	defer span.End()

	// Load from disk persistence first
	if c.config.PersistenceEnabled {
		if err := c.loadFromDisk(ctx); err != nil {
			span.AddEvent("disk_load_failed", trace.WithAttributes(
				attribute.String("error", err.Error()),
			))
		}
	}

	// Initial synchronous load from Flagr
	loadCtx, cancel := context.WithTimeout(ctx, c.config.InitialTimeout)
	defer cancel()

	if err := c.refreshFlags(loadCtx); err != nil {
		span.RecordError(err)
		// Don't fail if we have disk cache
		if c.cache.Metrics.KeysAdded() == 0 {
			return fmt.Errorf("initial flag load failed and no disk cache available: %w", err)
		}
	}

	// Start background refresh goroutine
	c.wg.Add(1)
	go c.refreshLoop()

	// Start webhook server if enabled
	if c.config.WebhookEnabled {
		if err := c.startWebhookServer(); err != nil {
			return fmt.Errorf("failed to start webhook server: %w", err)
		}
	}

	// Start admin API if enabled
	if c.config.AdminAPIEnabled {
		if err := c.startAdminAPI(); err != nil {
			return fmt.Errorf("failed to start admin API: %w", err)
		}
	}

	return nil
}

// Stop gracefully stops the cache and background services
func (c *Cache) Stop() error {
	c.cancel()

	// Stop servers
	if c.webhookServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		c.webhookServer.Shutdown(ctx)
	}

	if c.adminServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		c.adminServer.Shutdown(ctx)
	}

	c.wg.Wait()

	// Save to disk before shutdown
	if c.config.PersistenceEnabled {
		if err := c.saveToDisk(context.Background()); err != nil {
			return fmt.Errorf("failed to save cache to disk: %w", err)
		}
	}

	c.cache.Close()
	return nil
}

// Evaluate evaluates a feature flag for the given context
func (c *Cache) Evaluate(ctx context.Context, flagKey string, evalCtx EvaluationContext) (interface{}, error) {
	ctx, span := c.tracer.Start(ctx, "cache.evaluate",
		trace.WithAttributes(
			attribute.String("flag.key", flagKey),
			attribute.String("user.id", evalCtx.UserID),
		),
	)
	defer span.End()

	// Try to get from cache
	value, found := c.cache.Get(flagKey)
	if !found {
		c.cacheMisses.Add(ctx, 1, metric.WithAttributes(
			attribute.String("flag.key", flagKey),
		))
		span.SetAttributes(attribute.Bool("cache.hit", false))
		return c.handleCacheMiss(ctx, flagKey)
	}

	c.cacheHits.Add(ctx, 1, metric.WithAttributes(
		attribute.String("flag.key", flagKey),
	))
	span.SetAttributes(attribute.Bool("cache.hit", true))

	flag := value.(Flag)

	// Check if cache entry is stale
	if time.Now().After(flag.CacheExpiry) {
		span.AddEvent("cache_entry_stale")
	}

	// CRITICAL DECISION: Can we evaluate locally?
	canEvalLocally := c.canEvaluateLocally(flag)

	if !canEvalLocally {
		// ANY percentage-based rollout requires Flagr for consistent bucketing
		span.SetAttributes(
			attribute.String("evaluation.strategy", "dynamic"),
			attribute.String("reason", "partial_rollout_requires_flagr"),
		)
		return c.evaluateWithFlagr(ctx, flagKey, evalCtx)
	}

	// Determine evaluation strategy for 100% rollout flags
	strategy := c.determineEvaluationStrategy(flag)
	span.SetAttributes(attribute.String("evaluation.strategy", strategy))

	switch strategy {
	case "static":
		// Pure rule-based evaluation - no Flagr call needed
		// All segments are 100% rollout with deterministic constraints
		result := c.evaluator.Evaluate(ctx, flag, evalCtx)
		span.SetAttributes(attribute.String("result.value", fmt.Sprintf("%v", result)))
		return result, nil

	case "dynamic":
		// Requires Flagr evaluation for consistent bucketing
		return c.evaluateWithFlagr(ctx, flagKey, evalCtx)

	default:
		// Auto-detect based on flag configuration
		return c.evaluateAuto(ctx, flag, evalCtx)
	}
}

// determineEvaluationStrategy determines if a flag needs Flagr evaluation
func (c *Cache) determineEvaluationStrategy(flag Flag) string {
	// Explicit strategy set
	if flag.EvaluationType != "" && flag.EvaluationType != "auto" {
		return flag.EvaluationType
	}

	// CRITICAL: Check for percentage-based rollouts
	// Any rollout < 100% requires Flagr for consistent bucketing
	for _, segment := range flag.Segments {
		// Rollout percentage < 100% = needs Flagr
		if segment.RolloutPercent > 0 && segment.RolloutPercent < 100 {
			return "dynamic"
		}

		// Multiple variants with different percentages = needs Flagr
		if len(segment.Distributions) > 1 {
			return "dynamic"
		}

		// Single variant with < 100% = needs Flagr
		if len(segment.Distributions) == 1 && segment.Distributions[0].Percentage < 100 {
			return "dynamic"
		}
	}

	// All segments are 100% rollout with deterministic constraints
	// Can evaluate locally
	return "static"
}

// canEvaluateLocally checks if a flag can be safely evaluated without Flagr
func (c *Cache) canEvaluateLocally(flag Flag) bool {
	// No segments = simple default value = safe
	if len(flag.Segments) == 0 {
		return true
	}

	for _, segment := range flag.Segments {
		// Any partial rollout requires Flagr
		if segment.RolloutPercent > 0 && segment.RolloutPercent < 100 {
			return false
		}

		// A/B testing with multiple variants requires Flagr
		if len(segment.Distributions) > 1 {
			return false
		}

		// Single variant with partial percentage requires Flagr
		if len(segment.Distributions) == 1 {
			if segment.Distributions[0].Percentage < 100 {
				return false
			}
		}
	}

	// All segments are 100% rollout = deterministic = safe for local eval
	return true
}

// evaluateAuto automatically determines and executes evaluation strategy
func (c *Cache) evaluateAuto(ctx context.Context, flag Flag, evalCtx EvaluationContext) (interface{}, error) {
	strategy := c.determineEvaluationStrategy(flag)

	if strategy == "dynamic" {
		return c.evaluateWithFlagr(ctx, flag.Key, evalCtx)
	}

	return c.evaluator.Evaluate(ctx, flag, evalCtx), nil
}

// evaluateWithFlagr makes a direct call to Flagr for evaluation
func (c *Cache) evaluateWithFlagr(ctx context.Context, flagKey string, evalCtx EvaluationContext) (interface{}, error) {
	ctx, span := c.tracer.Start(ctx, "cache.evaluate_with_flagr",
		trace.WithAttributes(
			attribute.String("flag.key", flagKey),
		),
	)
	defer span.End()

	url := fmt.Sprintf("%s/api/v1/evaluation", c.config.FlagrEndpoint)

	// Build Flagr evaluation request
	reqBody := map[string]interface{}{
		"flagKey": flagKey,
		"entityContext": map[string]interface{}{
			"entityID":   evalCtx.UserID,
			"entityType": "user",
		},
		"context": evalCtx.Attributes,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.config.FlagrAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.FlagrAPIKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("flagr evaluation request failed: %w", err)
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("flagr returned status %d: %s", resp.StatusCode, string(body))
	}

	var evalResult struct {
		FlagKey     string                 `json:"flagKey"`
		VariantKey  string                 `json:"variantKey"`
		VariantID   int64                  `json:"variantID"`
		Attachment  interface{}            `json:"variantAttachment"`
		EvalContext map[string]interface{} `json:"evalContext"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&evalResult); err != nil {
		return nil, fmt.Errorf("failed to decode evaluation result: %w", err)
	}

	span.SetAttributes(
		attribute.String("variant.key", evalResult.VariantKey),
		attribute.Int64("variant.id", evalResult.VariantID),
	)

	// Return the variant attachment if available, otherwise variant key
	if evalResult.Attachment != nil {
		return evalResult.Attachment, nil
	}

	return evalResult.VariantKey, nil
}

// EvaluateBool is a convenience method that returns a boolean result
func (c *Cache) EvaluateBool(ctx context.Context, flagKey string, evalCtx EvaluationContext) bool {
	result, err := c.Evaluate(ctx, flagKey, evalCtx)
	if err != nil {
		return false
	}
	if b, ok := result.(bool); ok {
		return b
	}
	return false
}

// EvaluateString is a convenience method that returns a string result
func (c *Cache) EvaluateString(ctx context.Context, flagKey string, evalCtx EvaluationContext, defaultVal string) string {
	result, err := c.Evaluate(ctx, flagKey, evalCtx)
	if err != nil {
		return defaultVal
	}
	if s, ok := result.(string); ok {
		return s
	}
	return defaultVal
}

// EvaluateInt is a convenience method that returns an int result
func (c *Cache) EvaluateInt(ctx context.Context, flagKey string, evalCtx EvaluationContext, defaultVal int) int {
	result, err := c.Evaluate(ctx, flagKey, evalCtx)
	if err != nil {
		return defaultVal
	}
	if i, ok := result.(int); ok {
		return i
	}
	if f, ok := result.(float64); ok {
		return int(f)
	}
	return defaultVal
}

// InvalidateFlag manually invalidates a specific flag
func (c *Cache) InvalidateFlag(ctx context.Context, flagKey string) {
	_, span := c.tracer.Start(ctx, "cache.invalidate_flag",
		trace.WithAttributes(attribute.String("flag.key", flagKey)),
	)
	defer span.End()

	c.cache.Del(flagKey)
}

// InvalidateAll clears the entire cache
func (c *Cache) InvalidateAll(ctx context.Context) {
	_, span := c.tracer.Start(ctx, "cache.invalidate_all")
	defer span.End()

	c.cache.Clear()
}

// GetCacheStats returns current cache statistics
func (c *Cache) GetCacheStats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	metrics := c.cache.Metrics

	return Stats{
		KeysAdded:        metrics.KeysAdded(),
		KeysUpdated:      metrics.KeysUpdated(),
		KeysEvicted:      metrics.KeysEvicted(),
		CostAdded:        metrics.CostAdded(),
		CostEvicted:      metrics.CostEvicted(),
		SetsDropped:      metrics.SetsDropped(),
		SetsRejected:     metrics.SetsRejected(),
		GetsKept:         metrics.GetsKept(),
		GetsDropped:      metrics.GetsDropped(),
		HitRatio:         metrics.Ratio(),
		LastRefresh:      c.lastRefresh,
		ConsecutiveFails: c.consecutiveFails,
		CircuitOpen:      c.circuitOpen,
	}
}

// Stats represents cache statistics
type Stats struct {
	KeysAdded        uint64    `json:"keys_added"`
	KeysUpdated      uint64    `json:"keys_updated"`
	KeysEvicted      uint64    `json:"keys_evicted"`
	CostAdded        uint64    `json:"cost_added"`
	CostEvicted      uint64    `json:"cost_evicted"`
	SetsDropped      uint64    `json:"sets_dropped"`
	SetsRejected     uint64    `json:"sets_rejected"`
	GetsKept         uint64    `json:"gets_kept"`
	GetsDropped      uint64    `json:"gets_dropped"`
	HitRatio         float64   `json:"hit_ratio"`
	LastRefresh      time.Time `json:"last_refresh"`
	ConsecutiveFails int       `json:"consecutive_fails"`
	CircuitOpen      bool      `json:"circuit_open"`
}

// refreshLoop runs the periodic refresh in the background
func (c *Cache) refreshLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			ctx, span := c.tracer.Start(c.ctx, "cache.refresh_tick")

			if err := c.refreshFlags(ctx); err != nil {
				span.RecordError(err)
				c.handleRefreshError(err)
			} else {
				c.resetCircuitBreaker()
				if c.config.PersistenceEnabled {
					if err := c.saveToDisk(ctx); err != nil {
						span.AddEvent("disk_save_failed", trace.WithAttributes(
							attribute.String("error", err.Error()),
						))
					}
				}
			}

			span.End()
		}
	}
}

// refreshFlags fetches all flags from Flagr and updates the cache
func (c *Cache) refreshFlags(ctx context.Context) error {
	ctx, span := c.tracer.Start(ctx, "cache.refresh_flags")
	defer span.End()

	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		c.refreshLatency.Record(ctx, float64(duration))
	}()

	// Check circuit breaker
	c.mu.RLock()
	if c.circuitOpen {
		c.mu.RUnlock()
		span.SetAttributes(attribute.Bool("circuit_breaker.open", true))
		return fmt.Errorf("circuit breaker is open")
	}
	c.mu.RUnlock()

	// Fetch flags with retries
	var flags []Flag
	var lastErr error

	for attempt := 0; attempt < c.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * time.Second
			time.Sleep(backoff)
		}

		flags, lastErr = c.fetchFlagsFromFlagr(ctx)
		if lastErr == nil {
			break
		}

		span.AddEvent("retry", trace.WithAttributes(
			attribute.Int("attempt", attempt+1),
			attribute.String("error", lastErr.Error()),
		))
	}

	if lastErr != nil {
		c.refreshFailure.Add(ctx, 1)
		return lastErr
	}

	// Update cache
	now := time.Now()
	for _, flag := range flags {
		flag.LastUpdated = now
		flag.CacheExpiry = now.Add(c.getFlagTTL(flag))
		c.cache.Set(flag.Key, flag, 1)
	}

	c.cache.Wait()

	c.mu.Lock()
	c.lastRefresh = now
	c.mu.Unlock()

	c.refreshSuccess.Add(ctx, 1)
	span.SetAttributes(
		attribute.Int("flags.count", len(flags)),
		attribute.String("refresh.timestamp", now.Format(time.RFC3339)),
	)

	return nil
}

// fetchFlagsFromFlagr makes the HTTP request to Flagr API
func (c *Cache) fetchFlagsFromFlagr(ctx context.Context) ([]Flag, error) {
	ctx, span := c.tracer.Start(ctx, "cache.fetch_from_flagr")
	defer span.End()

	url := fmt.Sprintf("%s/api/v1/flags?expand=segments,variants,distributions", c.config.FlagrEndpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.config.FlagrAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.FlagrAPIKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("flagr request failed: %w", err)
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("flagr returned status %d: %s", resp.StatusCode, string(body))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var flagrFlags []FlagrResponse
	if err := json.Unmarshal(bodyBytes, &flagrFlags); err != nil {
		return nil, fmt.Errorf("failed to decode flags: %w", err)
	}

	var completeFlagrFlags []FlagrResponse
	for _, f := range flagrFlags {
		url := fmt.Sprintf("%s/api/v1/flags/%d", c.config.FlagrEndpoint, f.ID)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}

		if c.config.FlagrAPIKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.config.FlagrAPIKey)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("flagr request failed: %w", err)
		}
		defer resp.Body.Close()

		span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("flagr returned status %d: %s", resp.StatusCode, string(body))
		}

		detailBodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var r FlagrResponse
		if err := json.Unmarshal(detailBodyBytes, &r); err != nil {
			return nil, fmt.Errorf("failed to decode flags: %w", err)
		}
		completeFlagrFlags = append(completeFlagrFlags, r)
		fmt.Printf("📦 Detail Raw Response: %s\n", string(detailBodyBytes))
	}

	vexillaFlags := parseFlagrResponse(completeFlagrFlags)

	span.SetAttributes(attribute.Int("flags.count", len(vexillaFlags)))

	// Logo após ler bodyBytes
	fmt.Printf("📦 Raw Response: %s\n", string(bodyBytes))
	fmt.Printf("📦 Flagr Flags Parsed: %d\n", len(flagrFlags))
	fmt.Printf("📦 Flagr Complete Flags Converted: %d\n", len(completeFlagrFlags))
	fmt.Printf("📦 Vexilla Flags Converted: %d\n", len(vexillaFlags))

	return vexillaFlags, nil
}

// handleCacheMiss handles the case when a flag is not in cache
func (c *Cache) handleCacheMiss(ctx context.Context, flagKey string) (interface{}, error) {
	switch c.config.FallbackStrategy {
	case "fail_open":
		return true, nil
	case "fail_closed":
		return false, nil
	case "last_known_good":
		return c.loadFlagFromDisk(ctx, flagKey)
	default:
		return nil, fmt.Errorf("flag not found: %s", flagKey)
	}
}

// getFlagTTL returns the TTL for a flag based on its criticality
func (c *Cache) getFlagTTL(flag Flag) time.Duration {
	if flag.TTL > 0 {
		return flag.TTL
	}
	if flag.IsCritical {
		return 30 * time.Second
	}
	return c.config.RefreshInterval
}

// handleRefreshError handles errors during refresh
func (c *Cache) handleRefreshError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.consecutiveFails++

	if c.consecutiveFails >= 3 {
		c.circuitOpen = true
		time.AfterFunc(30*time.Second, func() {
			c.mu.Lock()
			c.circuitOpen = false
			c.mu.Unlock()
		})
	}
}

// resetCircuitBreaker resets the circuit breaker on successful refresh
func (c *Cache) resetCircuitBreaker() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.consecutiveFails = 0
	c.circuitOpen = false
}

func (c *Cache) initMetrics() error {
	var err error

	c.cacheHits, err = c.meter.Int64Counter("flagr.cache.hits")
	if err != nil {
		return err
	}

	c.cacheMisses, err = c.meter.Int64Counter("flagr.cache.misses")
	if err != nil {
		return err
	}

	c.refreshSuccess, err = c.meter.Int64Counter("flagr.refresh.success")
	if err != nil {
		return err
	}

	c.refreshFailure, err = c.meter.Int64Counter("flagr.refresh.failure")
	if err != nil {
		return err
	}

	c.refreshLatency, err = c.meter.Float64Histogram("flagr.refresh.duration")
	if err != nil {
		return err
	}

	c.webhookEvents, err = c.meter.Int64Counter("flagr.webhook.events")
	if err != nil {
		return err
	}

	c.cacheSize, err = c.meter.Int64ObservableGauge("flagr.cache.size",
		metric.WithInt64Callback(func(ctx context.Context, observer metric.Int64Observer) error {
			metrics := c.cache.Metrics
			observer.Observe(int64(metrics.KeysAdded() - metrics.KeysEvicted()))
			return nil
		}),
	)
	return err
}
