package vexilla

import (
	"time"
)

// Config holds all configuration for a Vexilla client.
type Config struct {
	// Flagr configuration
	Flagr FlagrConfig

	// Cache configuration
	Cache CacheConfig

	// Circuit breaker configuration
	CircuitBreaker CircuitBreakerConfig

	// Fallback strategy when flag evaluation fails
	// Options: "fail_open", "fail_closed", "error"
	FallbackStrategy string
}

// FlagrConfig configures the connection to Flagr.
type FlagrConfig struct {
	// Endpoint is the base URL of the Flagr server
	// Example: "http://localhost:18000"
	Endpoint string

	// APIKey is an optional authentication token
	APIKey string

	// Timeout for HTTP requests to Flagr
	Timeout time.Duration

	// MaxRetries for failed requests
	MaxRetries int
}

// CacheConfig configures cache behavior.
type CacheConfig struct {
	// RefreshInterval determines how often to fetch flags from Flagr
	RefreshInterval time.Duration

	// InitialTimeout is the timeout for the initial flag load
	InitialTimeout time.Duration

	// Filter configures which flags to cache (resource optimization)
	Filter FilterConfig
}

// FilterConfig configures flag filtering for resource optimization.
// This is especially useful in microservice architectures where each
// service only needs a subset of flags.
type FilterConfig struct {
	// OnlyEnabled filters out disabled flags
	// When true, only stores flags with Enabled=true
	OnlyEnabled bool

	// ServiceName is the current service identifier
	// Used to filter flags by tags
	ServiceName string

	// RequireServiceTag when true, only caches flags that have
	// the ServiceName in their tags list.
	// This dramatically reduces memory footprint.
	RequireServiceTag bool

	// AdditionalTags allows filtering by additional tag values
	// Useful for environment-specific flags (e.g., "production", "staging")
	AdditionalTags []string

	// TagMatchMode determines how tags are matched
	// "any": flag must have ANY of the tags
	// "all": flag must have ALL of the tags
	TagMatchMode string
}

// CircuitBreakerConfig configures the circuit breaker.
type CircuitBreakerConfig struct {
	// Threshold is the number of consecutive failures before opening
	Threshold int

	// Timeout is how long to wait before attempting recovery
	Timeout time.Duration
}

// DefaultConfig returns recommended default configuration.
func DefaultConfig() Config {
	return Config{
		Flagr: FlagrConfig{
			Timeout:    5 * time.Second,
			MaxRetries: 3,
		},
		Cache: CacheConfig{
			RefreshInterval: 5 * time.Minute,
			InitialTimeout:  10 * time.Second,
			Filter: FilterConfig{
				OnlyEnabled:  true,
				TagMatchMode: "any",
			},
		},
		CircuitBreaker: CircuitBreakerConfig{
			Threshold: 3,
			Timeout:   30 * time.Second,
		},
		FallbackStrategy: "fail_closed",
	}
}
