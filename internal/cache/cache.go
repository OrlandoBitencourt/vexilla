package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
	"github.com/OrlandoBitencourt/vexilla/internal/evaluator"
	"github.com/OrlandoBitencourt/vexilla/internal/flagr"
	"github.com/OrlandoBitencourt/vexilla/internal/storage"
)

// Cache is the main orchestrator that coordinates all components
type Cache struct {
	// Dependencies (injected)
	flagrClient flagr.Client
	storage     storage.Storage
	evaluator   evaluator.Evaluator

	// Configuration
	config Config

	// State management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Refresh state
	mu               sync.RWMutex
	lastRefresh      time.Time
	consecutiveFails int
	circuitOpen      bool
}

// New creates a new cache with the given options
func New(opts ...Option) (*Cache, error) {
	c := &Cache{
		config: DefaultConfig(),
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	// Validate required dependencies
	if c.flagrClient == nil {
		return nil, fmt.Errorf("flagr client is required")
	}
	if c.storage == nil {
		return nil, fmt.Errorf("storage is required")
	}
	if c.evaluator == nil {
		return nil, fmt.Errorf("evaluator is required")
	}

	// Validate config
	if err := c.config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return c, nil
}

// Start initializes the cache and starts background processes
func (c *Cache) Start(ctx context.Context) error {
	c.ctx, c.cancel = context.WithCancel(ctx)

	loadCtx, cancel := context.WithTimeout(c.ctx, c.config.InitialTimeout)
	defer cancel()

	if err := c.refreshFlags(loadCtx); err != nil {
		if diskStorage, ok := c.storage.(*storage.DiskStorage); ok {
			snapshot, loadErr := diskStorage.LoadSnapshot(loadCtx)
			if loadErr == nil && len(snapshot) > 0 {
				for key, flag := range snapshot {
					c.storage.Set(loadCtx, key, flag, 0)
				}
			} else {
				return fmt.Errorf("initial flag load failed and no disk cache available: %w", err)
			}
		} else {
			return fmt.Errorf("initial flag load failed: %w", err)
		}
	}

	if c.config.RefreshInterval > 0 {
		c.wg.Add(1)
		go c.refreshLoop()
	}

	return nil
}

func (c *Cache) Sync(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return c.refreshFlags(ctx)
}

// Stop gracefully stops the cache
func (c *Cache) Stop() error {
	c.cancel()
	c.wg.Wait()

	// Save snapshot to disk if available
	if diskStorage, ok := c.storage.(*storage.DiskStorage); ok {
		keys, _ := c.storage.List(context.Background())
		snapshot := make(map[string]domain.Flag)

		for _, key := range keys {
			flag, err := c.storage.Get(context.Background(), key)
			if err == nil {
				snapshot[key] = *flag
			}
		}

		diskStorage.SaveSnapshot(context.Background(), snapshot)
	}

	return c.storage.Close()
}

// Evaluate evaluates a flag for the given context
func (c *Cache) Evaluate(ctx context.Context, flagKey string, evalCtx domain.EvaluationContext) (*domain.EvaluationResult, error) {
	// Get flag from storage
	flag, err := c.storage.Get(ctx, flagKey)
	if err != nil {
		if domain.IsNotFound(err) {
			return c.handleMissingFlag(ctx, flagKey, evalCtx)
		}
		return nil, err
	}

	// Determine evaluation strategy
	if c.evaluator.CanEvaluateLocally(*flag) {
		// Evaluate locally
		return c.evaluator.Evaluate(ctx, *flag, evalCtx)
	}

	// Evaluate remotely via Flagr
	return c.evaluateRemote(ctx, flagKey, evalCtx)
}

// EvaluateBool is a convenience method that returns a boolean result
func (c *Cache) EvaluateBool(ctx context.Context, flagKey string, evalCtx domain.EvaluationContext) bool {
	result, err := c.Evaluate(ctx, flagKey, evalCtx)
	if err != nil {
		return false
	}

	if result.VariantAttachment != nil {
		for _, key := range []string{"value", "enabled"} {
			if raw, ok := result.VariantAttachment[key]; ok {
				var b bool
				if err := json.Unmarshal(raw, &b); err == nil {
					if b == true {
						return b
					}
				}
			}
		}
	}

	// Fallback using VariantKey
	return result.VariantKey == "enabled" ||
		result.VariantKey == "on" ||
		result.VariantKey == "true"
}

// EvaluateString is a convenience method that returns a string result
func (c *Cache) EvaluateString(ctx context.Context, flagKey string, evalCtx domain.EvaluationContext, defaultVal string) string {
	result, err := c.Evaluate(ctx, flagKey, evalCtx)
	if err != nil {
		return defaultVal
	}

	if result.VariantAttachment != nil {
		if raw, ok := result.VariantAttachment["value"]; ok {
			var s string
			if err := json.Unmarshal(raw, &s); err == nil {
				return s
			}
		}
	}

	if result.VariantKey != "" {
		return result.VariantKey
	}

	return defaultVal
}

// EvaluateInt is a convenience method that returns an int result
func (c *Cache) EvaluateInt(ctx context.Context, flagKey string, evalCtx domain.EvaluationContext, defaultVal int) int {
	result, err := c.Evaluate(ctx, flagKey, evalCtx)
	if err != nil {
		return defaultVal
	}

	if result.VariantAttachment != nil {
		if raw, ok := result.VariantAttachment["value"]; ok {

			// Primeiro tenta int
			var i int
			if err := json.Unmarshal(raw, &i); err == nil {
				return i
			}

			// Depois tenta float64 (JSON padrão)
			var f float64
			if err := json.Unmarshal(raw, &f); err == nil {
				return int(f)
			}
		}
	}

	return defaultVal
}

// evaluateRemote evaluates flag using Flagr
func (c *Cache) evaluateRemote(ctx context.Context, flagKey string, evalCtx domain.EvaluationContext) (*domain.EvaluationResult, error) {
	// Check circuit breaker
	c.mu.RLock()
	if c.circuitOpen {
		c.mu.RUnlock()
		return nil, domain.NewCircuitOpenError("cannot evaluate remotely")
	}
	c.mu.RUnlock()

	return c.flagrClient.EvaluateFlag(ctx, flagKey, evalCtx)
}

// handleMissingFlag handles the case when a flag is not in cache
func (c *Cache) handleMissingFlag(ctx context.Context, flagKey string, evalCtx domain.EvaluationContext) (*domain.EvaluationResult, error) {
	// Try to fetch from Flagr
	flags, err := c.flagrClient.GetAllFlags(ctx)
	if err != nil {
		// Apply fallback strategy
		return c.applyFallbackStrategy(flagKey)
	}

	// Update cache
	for _, flag := range flags {
		c.storage.Set(ctx, flag.Key, flag, c.config.RefreshInterval)
	}

	// Try again
	flag, err := c.storage.Get(ctx, flagKey)
	if err != nil {
		return c.applyFallbackStrategy(flagKey)
	}

	return c.evaluator.Evaluate(ctx, *flag, evalCtx)
}

// applyFallbackStrategy applies configured fallback strategy
func (c *Cache) applyFallbackStrategy(flagKey string) (*domain.EvaluationResult, error) {
	switch c.config.FallbackStrategy {
	case "fail_open":
		return &domain.EvaluationResult{
			FlagKey:          flagKey,
			EvaluationReason: "fallback: fail_open",
			VariantKey:       "enabled",
			VariantAttachment: map[string]json.RawMessage{
				"enabled": marshalRaw(true),
			},
		}, nil

	case "fail_closed":
		return &domain.EvaluationResult{
			FlagKey:          flagKey,
			EvaluationReason: "fallback: fail_closed",
			VariantKey:       "disabled",
			VariantAttachment: map[string]json.RawMessage{
				"enabled": marshalRaw(false),
			},
		}, nil

	default:
		return nil, domain.NewNotFoundError("flag", flagKey)
	}
}

// refreshLoop runs the periodic refresh in background
func (c *Cache) refreshLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return

		case <-ticker.C:
			ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)

			if err := c.refreshFlags(ctx); err != nil {
				c.handleRefreshError(err)
			} else {
				c.resetCircuitBreaker()
			}

			cancel()
		}
	}
}

func (c *Cache) refreshFlags(ctx context.Context) error {
	// Circuit breaker check
	c.mu.RLock()
	if c.circuitOpen {
		c.mu.RUnlock()
		return domain.NewCircuitOpenError("refresh blocked")
	}
	c.mu.RUnlock()

	flags, err := c.flagrClient.GetAllFlags(ctx)
	if err != nil {
		// erro → incrementa falhas
		c.mu.Lock()
		c.consecutiveFails++

		if c.consecutiveFails >= c.config.CircuitBreakerThreshold {
			c.circuitOpen = true
		}

		c.mu.Unlock()
		return fmt.Errorf("failed to fetch flags: %w", err)
	}

	// sucesso → reset
	c.mu.Lock()
	c.consecutiveFails = 0
	c.circuitOpen = false
	c.mu.Unlock()

	// Atualiza o cache
	for _, flag := range flags {
		meta := FlagMetadata{
			Key:     flag.Key,
			Enabled: flag.Enabled,
			Tags:    extractTagValues(flag.Tags),
		}

		if !c.config.FilterConfig.ShouldCacheFlag(meta) {
			continue
		}

		if err := c.storage.Set(ctx, flag.Key, flag, c.config.RefreshInterval); err != nil {
			return fmt.Errorf("failed to cache flag %s: %w", flag.Key, err)
		}
	}

	// Atualiza lastRefresh
	c.mu.Lock()
	c.lastRefresh = time.Now()
	c.mu.Unlock()

	return nil
}

// handleRefreshError handles refresh failures
func (c *Cache) handleRefreshError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.consecutiveFails++

	// Open circuit breaker after 3 consecutive failures
	if c.consecutiveFails >= 3 {
		c.circuitOpen = true

		// Auto-reset after 30 seconds
		time.AfterFunc(30*time.Second, func() {
			c.mu.Lock()
			c.circuitOpen = false
			c.mu.Unlock()
		})
	}
}

// resetCircuitBreaker resets circuit breaker state
func (c *Cache) resetCircuitBreaker() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.consecutiveFails = 0
	c.circuitOpen = false
}

// InvalidateFlag removes a flag from cache
func (c *Cache) InvalidateFlag(ctx context.Context, flagKey string) error {
	return c.storage.Delete(ctx, flagKey)
}

// InvalidateAll clears the entire cache
func (c *Cache) InvalidateAll(ctx context.Context) error {
	return c.storage.Clear(ctx)
}

// GetMetrics returns cache metrics
func (c *Cache) GetMetrics() Metrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	storageMetrics := c.storage.Metrics()

	return Metrics{
		Storage:          storageMetrics,
		LastRefresh:      c.lastRefresh,
		ConsecutiveFails: c.consecutiveFails,
		CircuitOpen:      c.circuitOpen,
	}
}

// Metrics represents cache metrics
type Metrics struct {
	Storage          storage.Metrics
	LastRefresh      time.Time
	ConsecutiveFails int
	CircuitOpen      bool
}

// raw marshals a value into json.RawMessage and ignores marshal errors on purpose
func marshalRaw(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return json.RawMessage(b)
}
func extractTagValues(tags []domain.Tag) []string {
	out := make([]string, len(tags))
	for i, t := range tags {
		out[i] = t.Value
	}
	return out
}
