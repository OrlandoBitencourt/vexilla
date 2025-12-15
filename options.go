package vexilla

import (
	"fmt"
	"time"

	"github.com/OrlandoBitencourt/vexilla/internal/cache"
	"github.com/OrlandoBitencourt/vexilla/internal/evaluator"
	"github.com/OrlandoBitencourt/vexilla/internal/flagr"
	"github.com/OrlandoBitencourt/vexilla/internal/storage"
)

// Option configures a Vexilla client.
type Option func(*clientConfig) error

// clientConfig holds internal configuration.
type clientConfig struct {
	flagrEndpoint   string
	flagrAPIKey     string
	flagrTimeout    time.Duration
	flagrMaxRetries int

	refreshInterval  time.Duration
	initialTimeout   time.Duration
	fallbackStrategy string

	circuitThreshold int
	circuitTimeout   time.Duration

	onlyEnabled       bool
	serviceName       string
	requireServiceTag bool
	additionalTags    []string
	tagMatchMode      string
}

// toCacheOptions converts clientConfig to cache options.
func (c *clientConfig) toCacheOptions() []cache.Option {
	opts := []cache.Option{}

	// Create Flagr client
	if c.flagrEndpoint != "" {
		flagrClient := flagr.NewHTTPClient(flagr.Config{
			Endpoint:   c.flagrEndpoint,
			APIKey:     c.flagrAPIKey,
			Timeout:    c.flagrTimeout,
			MaxRetries: c.flagrMaxRetries,
		})
		opts = append(opts, cache.WithFlagrClient(flagrClient))
	}

	// Create storage
	memStorage, _ := storage.NewMemoryStorage(storage.DefaultConfig())
	opts = append(opts, cache.WithStorage(memStorage))

	// Create evaluator
	eval := evaluator.New()
	opts = append(opts, cache.WithEvaluator(eval))

	// Apply configuration
	if c.refreshInterval > 0 {
		opts = append(opts, cache.WithRefreshInterval(c.refreshInterval))
	}

	if c.initialTimeout > 0 {
		opts = append(opts, cache.WithInitialTimeout(c.initialTimeout))
	}

	if c.fallbackStrategy != "" {
		opts = append(opts, cache.WithFallbackStrategy(c.fallbackStrategy))
	}

	if c.circuitThreshold > 0 {
		opts = append(opts, cache.WithCircuitBreaker(c.circuitThreshold, c.circuitTimeout))
	}

	// Filtering options
	if c.onlyEnabled {
		opts = append(opts, cache.WithOnlyEnabled(true))
	}

	if c.serviceName != "" && c.requireServiceTag {
		opts = append(opts, cache.WithServiceTag(c.serviceName, true))
	}

	if len(c.additionalTags) > 0 {
		matchMode := c.tagMatchMode
		if matchMode == "" {
			matchMode = "any"
		}
		opts = append(opts, cache.WithAdditionalTags(c.additionalTags, matchMode))
	}

	return opts
}

// WithFlagrEndpoint sets the Flagr server endpoint.
// This is required.
//
// Example: vexilla.WithFlagrEndpoint("http://localhost:18000")
func WithFlagrEndpoint(endpoint string) Option {
	return func(c *clientConfig) error {
		if endpoint == "" {
			return fmt.Errorf("flagr endpoint cannot be empty")
		}
		c.flagrEndpoint = endpoint
		return nil
	}
}

// WithFlagrAPIKey sets the Flagr API key for authentication.
func WithFlagrAPIKey(apiKey string) Option {
	return func(c *clientConfig) error {
		c.flagrAPIKey = apiKey
		return nil
	}
}

// WithFlagrTimeout sets the HTTP timeout for Flagr requests.
func WithFlagrTimeout(timeout time.Duration) Option {
	return func(c *clientConfig) error {
		c.flagrTimeout = timeout
		return nil
	}
}

// WithFlagrMaxRetries sets the maximum number of retries for failed Flagr requests.
func WithFlagrMaxRetries(maxRetries int) Option {
	return func(c *clientConfig) error {
		c.flagrMaxRetries = maxRetries
		return nil
	}
}

// WithRefreshInterval sets how often to refresh flags from Flagr.
// Default: 5 minutes
//
// Example: vexilla.WithRefreshInterval(5 * time.Minute)
func WithRefreshInterval(interval time.Duration) Option {
	return func(c *clientConfig) error {
		c.refreshInterval = interval
		return nil
	}
}

// WithInitialTimeout sets the timeout for the initial flag load.
// Default: 10 seconds
func WithInitialTimeout(timeout time.Duration) Option {
	return func(c *clientConfig) error {
		c.initialTimeout = timeout
		return nil
	}
}

// WithFallbackStrategy sets the strategy for handling missing flags.
// Options: "fail_open", "fail_closed", "error"
// Default: "fail_closed"
func WithFallbackStrategy(strategy string) Option {
	return func(c *clientConfig) error {
		validStrategies := map[string]bool{
			"fail_open":   true,
			"fail_closed": true,
			"error":       true,
		}
		if !validStrategies[strategy] {
			return fmt.Errorf("invalid fallback strategy: %s", strategy)
		}
		c.fallbackStrategy = strategy
		return nil
	}
}

// WithCircuitBreaker configures the circuit breaker.
//
// Example: vexilla.WithCircuitBreaker(3, 30*time.Second)
func WithCircuitBreaker(threshold int, timeout time.Duration) Option {
	return func(c *clientConfig) error {
		c.circuitThreshold = threshold
		c.circuitTimeout = timeout
		return nil
	}
}

// WithOnlyEnabled filters out disabled flags.
// When true, only enabled flags are cached, reducing memory usage.
//
// Example: vexilla.WithOnlyEnabled(true)
func WithOnlyEnabled(enabled bool) Option {
	return func(c *clientConfig) error {
		c.onlyEnabled = enabled
		return nil
	}
}

// WithServiceTag filters flags by service name tag.
// Only flags tagged with the given service name will be cached.
// This is highly recommended for microservice architectures.
//
// Example: vexilla.WithServiceTag("user-service")
func WithServiceTag(serviceName string) Option {
	return func(c *clientConfig) error {
		c.serviceName = serviceName
		c.requireServiceTag = true
		return nil
	}
}

// WithAdditionalTags filters flags by additional tags.
// Useful for environment-specific flags (e.g., "production", "staging").
//
// matchMode can be "any" or "all":
//   - "any": flag must have ANY of the tags
//   - "all": flag must have ALL of the tags
//
// Example: vexilla.WithAdditionalTags([]string{"production"}, "any")
func WithAdditionalTags(tags []string, matchMode string) Option {
	return func(c *clientConfig) error {
		if matchMode != "any" && matchMode != "all" {
			return fmt.Errorf("tag match mode must be 'any' or 'all'")
		}
		c.additionalTags = tags
		c.tagMatchMode = matchMode
		return nil
	}
}

// WithConfig applies a full Config struct.
// This is an alternative to using individual options.
func WithConfig(cfg Config) Option {
	return func(c *clientConfig) error {
		c.flagrEndpoint = cfg.Flagr.Endpoint
		c.flagrAPIKey = cfg.Flagr.APIKey
		c.flagrTimeout = cfg.Flagr.Timeout
		c.flagrMaxRetries = cfg.Flagr.MaxRetries

		c.refreshInterval = cfg.Cache.RefreshInterval
		c.initialTimeout = cfg.Cache.InitialTimeout

		c.onlyEnabled = cfg.Cache.Filter.OnlyEnabled
		c.serviceName = cfg.Cache.Filter.ServiceName
		c.requireServiceTag = cfg.Cache.Filter.RequireServiceTag
		c.additionalTags = cfg.Cache.Filter.AdditionalTags
		c.tagMatchMode = cfg.Cache.Filter.TagMatchMode

		c.circuitThreshold = cfg.CircuitBreaker.Threshold
		c.circuitTimeout = cfg.CircuitBreaker.Timeout

		c.fallbackStrategy = cfg.FallbackStrategy

		return nil
	}
}
