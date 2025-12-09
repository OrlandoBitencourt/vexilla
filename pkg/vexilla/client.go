// pkg/vexilla/vexilla.go
package vexilla

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/OrlandoBitencourt/vexilla/pkg/cache"
	"github.com/OrlandoBitencourt/vexilla/pkg/client"
	"github.com/OrlandoBitencourt/vexilla/pkg/evaluator"
	"github.com/OrlandoBitencourt/vexilla/pkg/types"
)

// Client is the main Vexilla client
type Client struct {
	config    Config
	flagr     types.FlagrClient
	cache     types.FlagCache
	evaluator types.FlagEvaluator

	refreshTicker *time.Ticker
	stopCh        chan struct{}
	wg            sync.WaitGroup
	mu            sync.RWMutex
	started       bool
}

// Config holds Vexilla configuration
type Config struct {
	// Flagr connection
	FlagrEndpoint string
	FlagrAPIKey   string

	// Cache settings
	CacheMaxCost     int64
	CacheNumCounters int64

	// Persistence
	PersistenceEnabled bool
	PersistencePath    string

	// Refresh behavior
	RefreshInterval time.Duration
	InitialTimeout  time.Duration

	// HTTP settings
	HTTPTimeout   time.Duration
	RetryAttempts int

	// Fallback behavior
	FallbackStrategy string // "fail_closed", "fail_open", "last_known_good"
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		FlagrEndpoint:      "http://localhost:18000",
		CacheMaxCost:       1 << 30, // 1GB
		CacheNumCounters:   1e7,     // 10M
		PersistenceEnabled: false,
		PersistencePath:    "/tmp/vexilla",
		RefreshInterval:    5 * time.Minute,
		InitialTimeout:     10 * time.Second,
		HTTPTimeout:        5 * time.Second,
		RetryAttempts:      3,
		FallbackStrategy:   "fail_closed",
	}
}

// New creates a new Vexilla client
func New(config Config) (*Client, error) {
	// Create Flagr client
	flagrClient := client.NewClient(config.FlagrEndpoint, config.FlagrAPIKey)
	flagrClient.SetTimeout(config.HTTPTimeout)
	flagrClient.SetRetries(config.RetryAttempts)

	// Create cache
	cacheConfig := cache.Config{
		MaxCost:            config.CacheMaxCost,
		NumCounters:        config.CacheNumCounters,
		PersistenceEnabled: config.PersistenceEnabled,
		PersistencePath:    config.PersistencePath,
	}
	flagCache, err := cache.NewCache(cacheConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	// Create evaluator
	flagEvaluator := evaluator.NewEvaluator()

	return &Client{
		config:    config,
		flagr:     flagrClient,
		cache:     flagCache,
		evaluator: flagEvaluator,
		stopCh:    make(chan struct{}),
	}, nil
}

// Start begins background flag refresh
func (c *Client) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return fmt.Errorf("client already started")
	}

	// Initial flag load
	ctx, cancel := context.WithTimeout(context.Background(), c.config.InitialTimeout)
	defer cancel()

	if err := c.refreshFlags(ctx); err != nil {
		return fmt.Errorf("failed initial flag load: %w", err)
	}

	// Start background refresh
	c.refreshTicker = time.NewTicker(c.config.RefreshInterval)
	c.wg.Add(1)
	go c.refreshLoop()

	c.started = true
	return nil
}

// Stop gracefully stops the client
func (c *Client) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.started {
		return nil
	}

	close(c.stopCh)
	if c.refreshTicker != nil {
		c.refreshTicker.Stop()
	}
	c.wg.Wait()

	c.started = false
	return nil
}

// EvaluationContext contains context for flag evaluation
type EvaluationContext struct {
	EntityID string
	Context  map[string]interface{}
}

// toTypesContext converts to internal types
func (e EvaluationContext) toTypesContext() types.EvaluationContext {
	return types.EvaluationContext{
		EntityID: e.EntityID,
		Context:  e.Context,
	}
}

// EvaluationResult contains the result of flag evaluation
type EvaluationResult struct {
	FlagKey          string
	VariantKey       string
	Enabled          bool
	EvaluatedLocally bool
	EvaluationTime   time.Duration
	SegmentID        int64
	Error            error
}

// EvaluateBool evaluates a flag and returns a boolean
func (c *Client) EvaluateBool(ctx context.Context, flagKey string, evalCtx EvaluationContext) bool {
	result, err := c.Evaluate(ctx, flagKey, evalCtx)
	if err != nil {
		// Apply fallback strategy
		return c.applyFallback(flagKey, false)
	}
	return result.Enabled
}

// EvaluateString evaluates a flag and returns a string variant
func (c *Client) EvaluateString(ctx context.Context, flagKey string, evalCtx EvaluationContext, defaultValue string) string {
	result, err := c.Evaluate(ctx, flagKey, evalCtx)
	if err != nil || result.VariantKey == "" {
		return defaultValue
	}
	return result.VariantKey
}

// Evaluate performs full flag evaluation
func (c *Client) Evaluate(ctx context.Context, flagKey string, evalCtx EvaluationContext) (*EvaluationResult, error) {
	startTime := time.Now()
	typesCtx := evalCtx.toTypesContext()

	result := &EvaluationResult{
		FlagKey: flagKey,
	}

	// Try to get flag from cache
	flag, found := c.cache.GetFlag(flagKey)
	if !found {
		// Flag not in cache, fetch from Flagr
		var err error
		flag, err = c.flagr.GetFlag(ctx, flagKey)
		if err != nil {
			result.Error = fmt.Errorf("failed to fetch flag: %w", err)
			result.Enabled = c.applyFallback(flagKey, false)
			result.EvaluationTime = time.Since(startTime)
			return result, result.Error
		}

		// Cache the flag
		if err := c.cache.SetFlag(flag); err != nil {
			// Log warning but continue
		}
	}

	// Check if we can evaluate locally
	if c.evaluator.CanEvaluateLocally(flag) {
		enabled, err := c.evaluator.Evaluate(flag, typesCtx)
		if err != nil {
			result.Error = fmt.Errorf("local evaluation failed: %w", err)
			result.Enabled = c.applyFallback(flagKey, false)
		} else {
			result.Enabled = enabled
			result.EvaluatedLocally = true

			// Get variant key from flag
			if enabled && len(flag.Segments) > 0 {
				for _, seg := range flag.Segments {
					if len(seg.Distributions) > 0 {
						result.VariantKey = seg.Distributions[0].VariantKey
						result.SegmentID = seg.ID
						break
					}
				}
			}
		}
	} else {
		// Need to evaluate via Flagr
		enabled, err := c.flagr.EvaluateFlag(ctx, flagKey, typesCtx)
		if err != nil {
			result.Error = fmt.Errorf("flagr evaluation failed: %w", err)
			result.Enabled = c.applyFallback(flagKey, false)
		} else {
			result.Enabled = enabled
			result.EvaluatedLocally = false

			// For remote evaluation, we'd need the full response
			// For now, just set a basic variant
			if enabled {
				result.VariantKey = "enabled"
			}
		}
	}

	result.EvaluationTime = time.Since(startTime)
	return result, result.Error
}

// GetFlag retrieves a flag definition
func (c *Client) GetFlag(ctx context.Context, flagKey string) (*types.Flag, error) {
	// Try cache first
	flag, found := c.cache.GetFlag(flagKey)
	if found {
		return flag, nil
	}

	// Fetch from Flagr
	flag, err := c.flagr.GetFlag(ctx, flagKey)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch flag: %w", err)
	}

	// Cache it
	c.cache.SetFlag(flag)
	return flag, nil
}

// GetCacheStats returns cache statistics
func (c *Client) GetCacheStats() cache.Stats {
	return c.cache.GetStats()
}

// InvalidateFlag removes a flag from cache
func (c *Client) InvalidateFlag(flagKey string) {
	c.cache.InvalidateFlag(flagKey)
}

// RefreshFlags manually triggers a flag refresh
func (c *Client) RefreshFlags(ctx context.Context) error {
	return c.refreshFlags(ctx)
}

// refreshFlags fetches all flags from Flagr and updates cache
func (c *Client) refreshFlags(ctx context.Context) error {
	flags, err := c.flagr.ListFlags(ctx)
	if err != nil {
		return fmt.Errorf("failed to list flags: %w", err)
	}

	// Update cache
	if err := c.cache.SetFlags(flags); err != nil {
		return fmt.Errorf("failed to update cache: %w", err)
	}

	return nil
}

// refreshLoop runs periodic flag refresh
func (c *Client) refreshLoop() {
	defer c.wg.Done()

	for {
		select {
		case <-c.refreshTicker.C:
			ctx, cancel := context.WithTimeout(context.Background(), c.config.InitialTimeout)
			if err := c.refreshFlags(ctx); err != nil {
				// Log error but continue
				fmt.Printf("Warning: flag refresh failed: %v\n", err)
			}
			cancel()

		case <-c.stopCh:
			return
		}
	}
}

// applyFallback applies the configured fallback strategy
func (c *Client) applyFallback(flagKey string, defaultValue bool) bool {
	switch c.config.FallbackStrategy {
	case "fail_open":
		return true
	case "fail_closed":
		return false
	case "last_known_good":
		// Try to get from cache
		flag, found := c.cache.GetFlag(flagKey)
		if found && flag.Enabled {
			return true
		}
		return defaultValue
	default:
		return defaultValue
	}
}

// Health returns the health status of the client
func (c *Client) Health() map[string]interface{} {
	stats := c.cache.GetStats()

	return map[string]interface{}{
		"started": c.started,
		"cache": map[string]interface{}{
			"hits":        stats.HitRatio,
			"misses":      stats.ConsecutiveFails,
			"hit_rate":    c.cache.HitRate(),
			"last_update": stats.LastRefresh,
		},
	}
}
