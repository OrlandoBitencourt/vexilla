package cache

import (
	"time"

	"github.com/OrlandoBitencourt/vexilla/pkg/evaluator"
	"github.com/OrlandoBitencourt/vexilla/pkg/flagr"
	"github.com/OrlandoBitencourt/vexilla/pkg/storage"
)

// Option is a functional option for configuring Cache
type Option func(*Cache)

// WithFlagrClient sets the Flagr client
func WithFlagrClient(client flagr.Client) Option {
	return func(c *Cache) {
		c.flagrClient = client
	}
}

// WithStorage sets the storage implementation
func WithStorage(s storage.Storage) Option {
	return func(c *Cache) {
		c.storage = s
	}
}

// WithEvaluator sets the evaluator implementation
func WithEvaluator(e evaluator.Evaluator) Option {
	return func(c *Cache) {
		c.evaluator = e
	}
}

// WithConfig sets the configuration
func WithConfig(config Config) Option {
	return func(c *Cache) {
		c.config = config
	}
}

// WithRefreshInterval sets the refresh interval
func WithRefreshInterval(interval time.Duration) Option {
	return func(c *Cache) {
		c.config.RefreshInterval = interval
	}
}

// WithInitialTimeout sets the initial load timeout
func WithInitialTimeout(timeout time.Duration) Option {
	return func(c *Cache) {
		c.config.InitialTimeout = timeout
	}
}

// WithFallbackStrategy sets the fallback strategy
func WithFallbackStrategy(strategy string) Option {
	return func(c *Cache) {
		c.config.FallbackStrategy = strategy
	}
}

// WithCircuitBreaker configures circuit breaker
func WithCircuitBreaker(threshold int, timeout time.Duration) Option {
	return func(c *Cache) {
		c.config.CircuitBreakerThreshold = threshold
		c.config.CircuitBreakerTimeout = timeout
	}
}

// 🔥 NEW: Filtering Options

// WithOnlyEnabled configures cache to only store enabled flags
func WithOnlyEnabled(enabled bool) Option {
	return func(c *Cache) {
		c.config.FilterConfig.OnlyEnabled = enabled
	}
}

// WithServiceTag configures cache to only store flags tagged with service name
func WithServiceTag(serviceName string, require bool) Option {
	return func(c *Cache) {
		c.config.FilterConfig.ServiceName = serviceName
		c.config.FilterConfig.RequireServiceTag = require
	}
}

// WithAdditionalTags configures additional tag filtering
func WithAdditionalTags(tags []string, matchMode string) Option {
	return func(c *Cache) {
		c.config.FilterConfig.AdditionalTags = tags
		c.config.FilterConfig.TagMatchMode = matchMode
	}
}

// WithFilterConfig sets the complete filter configuration
func WithFilterConfig(config FilterConfig) Option {
	return func(c *Cache) {
		c.config.FilterConfig = config
	}
}

// SimpleConfig provides a simple configuration structure for v1 compatibility
type SimpleConfig struct {
	FlagrEndpoint   string
	FlagrAPIKey     string
	RefreshInterval time.Duration
	CachePath       string

	// 🔥 NEW: Filtering options
	ServiceName    string   // Filter by service name
	OnlyEnabled    bool     // Only cache enabled flags
	AdditionalTags []string // Additional tags to filter by
}

// NewSimple creates a cache with simple configuration (v1 compatible)
func NewSimple(config SimpleConfig) (*Cache, error) {
	// Create Flagr client
	flagrClient := flagr.NewHTTPClient(flagr.Config{
		Endpoint:   config.FlagrEndpoint,
		APIKey:     config.FlagrAPIKey,
		Timeout:    5 * time.Second,
		MaxRetries: 3,
	})

	// Create memory storage
	memStorage, err := storage.NewMemoryStorage(storage.DefaultConfig())
	if err != nil {
		return nil, err
	}

	// Create evaluator
	eval := evaluator.New()

	// Build options
	opts := []Option{
		WithFlagrClient(flagrClient),
		WithStorage(memStorage),
		WithEvaluator(eval),
	}

	if config.RefreshInterval > 0 {
		opts = append(opts, WithRefreshInterval(config.RefreshInterval))
	}

	// 🔥 NEW: Apply filtering options
	if config.OnlyEnabled {
		opts = append(opts, WithOnlyEnabled(true))
	}

	if config.ServiceName != "" {
		opts = append(opts, WithServiceTag(config.ServiceName, true))
	}

	if len(config.AdditionalTags) > 0 {
		opts = append(opts, WithAdditionalTags(config.AdditionalTags, "any"))
	}

	return New(opts...)
}
