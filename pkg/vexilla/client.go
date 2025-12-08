package vexilla

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/OrlandoBitencourt/vexilla/pkg/circuit"
	"github.com/OrlandoBitencourt/vexilla/pkg/client"
	"github.com/OrlandoBitencourt/vexilla/pkg/evaluator"
	"github.com/OrlandoBitencourt/vexilla/pkg/server"
	"github.com/OrlandoBitencourt/vexilla/pkg/storage"
	"github.com/OrlandoBitencourt/vexilla/pkg/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Client is the main Vexilla client for feature flag evaluation
type Client struct {
	config Config

	// Core components
	memoryStore *storage.MemoryStore
	diskStore   *storage.DiskStore
	flagrClient *client.FlagrClient
	evaluator   *evaluator.Evaluator
	strategy    *evaluator.Determiner
	breaker     *circuit.Breaker

	// Telemetry
	metrics *telemetry.Metrics
	tracer  *telemetry.Tracer

	// Servers
	webhookServer *server.WebhookServer
	adminServer   *server.AdminServer

	// State
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	mu               sync.RWMutex
	lastRefresh      time.Time
	consecutiveFails int
}

// New creates a new Vexilla client
func New(config Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Initialize memory store
	memoryStore, err := storage.NewMemoryStore(
		config.CacheMaxCost,
		config.CacheNumCounters,
		config.CacheBufferItems,
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create memory store: %w", err)
	}

	// Initialize disk store if enabled
	var diskStore *storage.DiskStore
	if config.PersistenceEnabled {
		diskStore, err = storage.NewDiskStore(config.PersistencePath)
		if err != nil {
			cancel()
			memoryStore.Close()
			return nil, fmt.Errorf("failed to create disk store: %w", err)
		}
	}

	// Initialize Flagr client
	flagrClient := client.NewFlagrClient(
		config.FlagrEndpoint,
		config.FlagrAPIKey,
		config.HTTPTimeout,
	)

	// Initialize telemetry
	metrics, err := telemetry.NewMetrics("vexilla")
	if err != nil {
		cancel()
		memoryStore.Close()
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	tracer := telemetry.NewTracer("vexilla")

	// Initialize circuit breaker
	breaker := circuit.NewBreaker(config.RetryAttempts, 30*time.Second)

	vexillaClient := &Client{
		config:      config,
		memoryStore: memoryStore,
		diskStore:   diskStore,
		flagrClient: flagrClient,
		evaluator:   evaluator.NewEvaluator(),
		strategy:    evaluator.NewDeterminer(),
		breaker:     breaker,
		metrics:     metrics,
		tracer:      tracer,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Setup cache size gauge
	metrics.SetCacheSizeGauge("vexilla", func(ctx context.Context, observer metric.Int64Observer) error {
		m := memoryStore.Metrics()
		observer.Observe(int64(m.KeysAdded() - m.KeysEvicted()))
		return nil
	})

	return vexillaClient, nil
}

// Start initializes and starts all background services
func (c *Client) Start() error {
	ctx, span := c.tracer.Tracer().Start(c.ctx, "vexilla.start")
	defer span.End()

	// Load from disk if enabled
	if c.config.PersistenceEnabled {
		if err := c.loadFromDisk(ctx); err != nil {
			span.AddEvent("disk_load_failed", attribute.String("error", err.Error()))
		}
	}

	// Initial synchronous load from Flagr
	loadCtx, cancel := context.WithTimeout(ctx, c.config.InitialTimeout)
	defer cancel()

	if err := c.refresh(loadCtx); err != nil {
		span.RecordError(err)
		// Don't fail if we have disk cache
		if c.diskStore != nil {
			span.AddEvent("using_disk_cache")
		} else {
			return fmt.Errorf("initial load failed and no disk cache available: %w", err)
		}
	}

	// Start background refresh
	c.wg.Add(1)
	go c.refreshLoop()

	// Start webhook server if enabled
	if c.config.WebhookEnabled {
		webhookServer, err := server.NewWebhookServer(
			c.config.WebhookPort,
			c.config.WebhookPath,
			c.config.WebhookSecret,
			c,
		)
		if err != nil {
			return fmt.Errorf("failed to create webhook server: %w", err)
		}
		c.webhookServer = webhookServer
		if err := c.webhookServer.Start(ctx); err != nil {
			return fmt.Errorf("failed to start webhook server: %w", err)
		}
	}

	// Start admin API if enabled
	if c.config.AdminAPIEnabled {
		adminServer := server.NewAdminServer(
			c.config.AdminAPIPort,
			c.config.AdminAPIPath,
			c,
		)
		c.adminServer = adminServer
		if err := c.adminServer.Start(ctx); err != nil {
			return fmt.Errorf("failed to start admin server: %w", err)
		}
	}

	return nil
}

// Stop gracefully stops all services
func (c *Client) Stop() error {
	c.cancel()

	// Stop servers
	if c.webhookServer != nil {
		c.webhookServer.Stop(context.Background())
	}
	if c.adminServer != nil {
		c.adminServer.Stop(context.Background())
	}

	c.wg.Wait()

	// Save to disk before shutdown
	if c.config.PersistenceEnabled {
		if err := c.saveToDisk(context.Background()); err != nil {
			return fmt.Errorf("failed to save to disk: %w", err)
		}
	}

	c.memoryStore.Close()
	return nil
}

// Evaluate evaluates a feature flag
func (c *Client) Evaluate(ctx context.Context, flagKey string, evalCtx EvaluationContext) (*EvaluationResult, error) {
	ctx, span := c.tracer.Tracer().Start(ctx, "vexilla.evaluate",
		attribute.String("flag.key", flagKey),
		attribute.String("entity.id", evalCtx.EntityID),
	)
	defer span.End()

	start := time.Now()

	// Get flag from cache
	flag, found := c.memoryStore.Get(ctx, flagKey)
	if !found {
		c.metrics.CacheMisses.Add(ctx, 1, metric.WithAttributes(
			attribute.String("flag.key", flagKey),
		))
		span.SetAttributes(attribute.Bool("cache.hit", false))
		return nil, ErrFlagNotFound{FlagKey: flagKey}
	}

	c.metrics.CacheHits.Add(ctx, 1, metric.WithAttributes(
		attribute.String("flag.key", flagKey),
	))
	span.SetAttributes(attribute.Bool("cache.hit", true))

	// Determine evaluation strategy
	if c.strategy.CanEvaluateLocally(*flag) {
		// Local evaluation
		result, err := c.evaluator.Evaluate(ctx, *flag, evalCtx)
		if err != nil {
			return nil, err
		}
		result.EvaluationTime = time.Since(start)

		c.metrics.LocalEvals.Add(ctx, 1, metric.WithAttributes(
			attribute.String("flag.key", flagKey),
		))
		span.SetAttributes(attribute.String("evaluation.strategy", "local"))

		return result, nil
	}

	// Remote evaluation (requires Flagr)
	result, err := c.flagrClient.PostEvaluation(ctx, flagKey, evalCtx)
	if err != nil {
		return nil, err
	}
	result.EvaluationTime = time.Since(start)

	c.metrics.RemoteEvals.Add(ctx, 1, metric.WithAttributes(
		attribute.String("flag.key", flagKey),
	))
	span.SetAttributes(attribute.String("evaluation.strategy", "remote"))

	return result, nil
}

// EvaluateBool is a convenience method that returns a boolean result
func (c *Client) EvaluateBool(ctx context.Context, flagKey string, evalCtx EvaluationContext) bool {
	result, err := c.Evaluate(ctx, flagKey, evalCtx)
	if err != nil {
		return false
	}

	if result.VariantKey == "enabled" || result.VariantKey == "on" || result.VariantKey == "true" {
		return true
	}

	return false
}

// EvaluateString is a convenience method that returns a string result
func (c *Client) EvaluateString(ctx context.Context, flagKey string, evalCtx EvaluationContext, defaultVal string) string {
	result, err := c.Evaluate(ctx, flagKey, evalCtx)
	if err != nil {
		return defaultVal
	}
	return result.VariantKey
}

// refresh fetches flags from Flagr and updates cache
func (c *Client) refresh(ctx context.Context) error {
	ctx, span := c.tracer.Tracer().Start(ctx, "vexilla.refresh")
	defer span.End()

	start := time.Now()
	defer func() {
		c.metrics.RefreshLatency.Record(ctx, float64(time.Since(start).Milliseconds()))
	}()

	// Use circuit breaker
	err := c.breaker.Call(func() error {
		flags, err := c.flagrClient.GetFlags(ctx)
		if err != nil {
			return err
		}

		// Update memory store
		for _, flag := range flags {
			c.memoryStore.Set(ctx, flag.Key, flag)
		}
		c.memoryStore.Wait()

		c.mu.Lock()
		c.lastRefresh = time.Now()
		c.consecutiveFails = 0
		c.mu.Unlock()

		span.SetAttributes(attribute.Int("flags.count", len(flags)))
		return nil
	})

	if err != nil {
		c.metrics.RefreshFailure.Add(ctx, 1)
		c.mu.Lock()
		c.consecutiveFails++
		c.mu.Unlock()
		return err
	}

	c.metrics.RefreshSuccess.Add(ctx, 1)
	return nil
}

// refreshLoop runs periodic background refresh
func (c *Client) refreshLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if err := c.refresh(c.ctx); err != nil {
				// Errors are logged via telemetry
			} else if c.config.PersistenceEnabled {
				c.saveToDisk(c.ctx)
			}
		}
	}
}

// Implement server.WebhookHandler interface
func (c *Client) OnFlagUpdated(ctx context.Context, flagKeys []string) error {
	for _, key := range flagKeys {
		c.memoryStore.Delete(ctx, key)
	}
	return c.refresh(ctx)
}

func (c *Client) OnFlagDeleted(ctx context.Context, flagKeys []string) error {
	for _, key := range flagKeys {
		c.memoryStore.Delete(ctx, key)
	}
	return nil
}

// Implement server.AdminHandler interface
func (c *Client) GetStats(ctx context.Context) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	m := c.memoryStore.Metrics()

	return Stats{
		KeysAdded:        m.KeysAdded(),
		KeysUpdated:      m.KeysUpdated(),
		KeysEvicted:      m.KeysEvicted(),
		CostAdded:        m.CostAdded(),
		CostEvicted:      m.CostEvicted(),
		SetsDropped:      m.SetsDropped(),
		SetsRejected:     m.SetsRejected(),
		GetsKept:         m.GetsKept(),
		GetsDropped:      m.GetsDropped(),
		HitRatio:         m.Ratio(),
		LastRefresh:      c.lastRefresh,
		ConsecutiveFails: c.consecutiveFails,
		CircuitOpen:      c.breaker.State() == circuit.StateOpen,
	}, nil
}

func (c *Client) InvalidateFlags(ctx context.Context, flagKeys []string) error {
	for _, key := range flagKeys {
		c.memoryStore.Delete(ctx, key)
	}
	return nil
}

func (c *Client) InvalidateAll(ctx context.Context) error {
	c.memoryStore.Clear(ctx)
	return nil
}

func (c *Client) ForceRefresh(ctx context.Context) error {
	return c.refresh(ctx)
}

func (c *Client) HealthCheck(ctx context.Context) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := "healthy"
	if c.breaker.State() == circuit.StateOpen {
		status = "degraded"
	}

	return map[string]interface{}{
		"status":            status,
		"circuit_open":      c.breaker.State() == circuit.StateOpen,
		"last_refresh":      c.lastRefresh.Format(time.RFC3339),
		"consecutive_fails": c.consecutiveFails,
	}, nil
}

// loadFromDisk loads flags from disk storage
func (c *Client) loadFromDisk(ctx context.Context) error {
	if c.diskStore == nil {
		return nil
	}

	flags, err := c.diskStore.Load(ctx)
	if err != nil {
		return err
	}

	for key, flag := range flags {
		c.memoryStore.Set(ctx, key, flag)
	}
	c.memoryStore.Wait()

	return nil
}

// saveToDisk saves flags to disk storage
func (c *Client) saveToDisk(ctx context.Context) error {
	if c.diskStore == nil {
		return nil
	}

	flags := make(map[string]Flag)
	return c.diskStore.Save(ctx, flags)
}
