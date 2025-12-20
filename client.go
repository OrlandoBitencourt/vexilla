// Package vexilla provides a high-performance caching layer for Flagr
// with intelligent local/remote evaluation routing.
package vexilla

import (
	"context"
	"errors"
	"fmt"

	"github.com/OrlandoBitencourt/vexilla/internal/cache"
	"github.com/OrlandoBitencourt/vexilla/internal/domain"
)

// Client is the main entry point for Vexilla.
// It provides flag evaluation with intelligent caching and routing.
type Client struct {
	cache *cache.Cache

	// Server configurations
	webhookEnabled bool
	webhookPort    int
	webhookSecret  string
	adminEnabled   bool
	adminPort      int
}

// New creates a new Vexilla client with the given options.
//
// Example:
//
//	client, err := vexilla.New(
//	    vexilla.WithFlagrEndpoint("http://localhost:18000"),
//	    vexilla.WithRefreshInterval(5 * time.Minute),
//	    vexilla.WithOnlyEnabled(true),
//	)
func New(opts ...Option) (*Client, error) {
	cfg := &clientConfig{}

	// Apply options
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	// Build cache with options
	cacheOpts := cfg.toCacheOptions()
	c, err := cache.New(cacheOpts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		cache:          c,
		webhookEnabled: cfg.webhookEnabled,
		webhookPort:    cfg.webhookPort,
		webhookSecret:  cfg.webhookSecret,
		adminEnabled:   cfg.adminEnabled,
		adminPort:      cfg.adminPort,
	}, nil
}

// Start initializes the client and begins background processes.
// This method blocks until the initial flag synchronization is complete.
//
// The initial sync has a timeout (default 10 seconds) configured via WithInitialTimeout().
// If the sync fails and no disk cache is available, Start returns an error.
//
// After the initial sync completes, background refresh begins automatically
// based on the configured refresh interval.
//
// This must be called before evaluating flags.
func (c *Client) Start(ctx context.Context) error {
	err := c.cache.Start(ctx)
	if err != nil {
		return err
	}

	if err := c.cache.Sync(ctx); err != nil {
		return err
	}

	// Start optional servers
	if c.webhookEnabled {
		if err := c.startWebhookServer(ctx, c.webhookPort, c.webhookSecret); err != nil {
			return fmt.Errorf("failed to start webhook server: %w", err)
		}
	}

	if c.adminEnabled {
		if err := c.startAdminServer(ctx, c.adminPort); err != nil {
			return fmt.Errorf("failed to start admin server: %w", err)
		}
	}

	return nil
}

func (c *Client) Sync(ctx context.Context) error {
	if c.cache == nil {
		return errors.New("vexilla client sync - cache not initialized")
	}
	return c.cache.Sync(ctx)
}

// Stop gracefully shuts down the client and its background processes.
func (c *Client) Stop() error {
	return c.cache.Stop()
}

// Bool evaluates a flag and returns a boolean result.
// Returns false if the flag is not found or evaluation fails.
//
// Example:
//
//	enabled := client.Bool(ctx, "new-feature", vexilla.Context{
//	    EntityID: "user-123",
//	    Attributes: map[string]any{"country": "BR"},
//	})
func (c *Client) Bool(ctx context.Context, flagKey string, evalCtx Context) bool {
	return c.cache.EvaluateBool(ctx, flagKey, toDomainContext(evalCtx))
}

// String evaluates a flag and returns a string result.
// Returns the default value if the flag is not found or evaluation fails.
func (c *Client) String(ctx context.Context, flagKey string, evalCtx Context, defaultVal string) string {
	return c.cache.EvaluateString(ctx, flagKey, toDomainContext(evalCtx), defaultVal)
}

// Int evaluates a flag and returns an integer result.
// Returns the default value if the flag is not found or evaluation fails.
func (c *Client) Int(ctx context.Context, flagKey string, evalCtx Context, defaultVal int) int {
	return c.cache.EvaluateInt(ctx, flagKey, toDomainContext(evalCtx), defaultVal)
}

// Evaluate performs a full flag evaluation and returns detailed results.
// Use this when you need access to variant attachments or evaluation metadata.
func (c *Client) Evaluate(ctx context.Context, flagKey string, evalCtx Context) (*Result, error) {
	result, err := c.cache.Evaluate(ctx, flagKey, toDomainContext(evalCtx))
	if err != nil {
		return nil, err
	}
	return toResult(result), nil
}

// InvalidateFlag removes a specific flag from the cache.
// The flag will be re-fetched on the next evaluation or refresh.
func (c *Client) InvalidateFlag(ctx context.Context, flagKey string) error {
	return c.cache.InvalidateFlag(ctx, flagKey)
}

// InvalidateAll clears all flags from the cache.
// All flags will be re-fetched on the next evaluation or refresh.
func (c *Client) InvalidateAll(ctx context.Context) error {
	return c.cache.InvalidateAll(ctx)
}

// Metrics returns current cache performance metrics.
func (c *Client) Metrics() Metrics {
	cacheMetrics := c.cache.GetMetrics()
	return Metrics{
		Storage: StorageMetrics{
			KeysAdded:   cacheMetrics.Storage.KeysAdded,
			KeysEvicted: cacheMetrics.Storage.KeysEvicted,
			HitRatio:    cacheMetrics.Storage.HitRatio,
		},
		LastRefresh:      cacheMetrics.LastRefresh,
		ConsecutiveFails: cacheMetrics.ConsecutiveFails,
		CircuitOpen:      cacheMetrics.CircuitOpen,
	}
}

// Internal conversion helpers

func toDomainContext(ctx Context) domain.EvaluationContext {
	return domain.EvaluationContext{
		EntityID:   ctx.EntityID,
		EntityType: ctx.EntityType,
		Context:    ctx.Attributes,
	}
}

func toResult(r *domain.EvaluationResult) *Result {
	return &Result{
		FlagKey:           r.FlagKey,
		VariantKey:        r.VariantKey,
		VariantAttachment: r.VariantAttachment,
		EvaluationReason:  r.EvaluationReason,
	}
}
