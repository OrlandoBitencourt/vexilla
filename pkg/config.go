package vexilla

import (
	"fmt"
	"time"
)

// Config holds the configuration for the flag cache
type Config struct {
	// Flagr connection settings
	FlagrEndpoint string
	FlagrAPIKey   string // Optional API key for authentication

	// Cache behavior
	RefreshInterval  time.Duration
	InitialTimeout   time.Duration
	HTTPTimeout      time.Duration
	RetryAttempts    int
	FallbackStrategy string // "fail_closed", "fail_open", "last_known_good"

	// Persistence
	PersistenceEnabled bool
	PersistencePath    string

	// Webhook settings
	WebhookEnabled bool
	WebhookPort    int
	WebhookPath    string // Default: "/webhook"
	WebhookSecret  string

	// Admin API settings
	AdminAPIEnabled bool
	AdminAPIPort    int
	AdminAPIPath    string // Base path for admin endpoints

	// Ristretto cache settings
	CacheMaxCost     int64 // Maximum cache size in bytes
	CacheNumCounters int64 // Number of counters for admission policy
	CacheBufferItems int64 // Number of keys per Get buffer
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() Config {
	return Config{
		RefreshInterval:    5 * time.Minute,
		InitialTimeout:     10 * time.Second,
		HTTPTimeout:        5 * time.Second,
		RetryAttempts:      3,
		FallbackStrategy:   "fail_closed",
		PersistenceEnabled: true,
		PersistencePath:    "/tmp/vexilla-cache",
		WebhookEnabled:     false,
		WebhookPort:        8081,
		WebhookPath:        "/webhook",
		AdminAPIEnabled:    false,
		AdminAPIPort:       8082,
		AdminAPIPath:       "/admin",
		CacheMaxCost:       1 << 30, // 1GB
		CacheNumCounters:   1e7,     // 10M
		CacheBufferItems:   64,
	}
}

// Validate checks if the configuration is valid
func (c Config) Validate() error {
	if c.FlagrEndpoint == "" {
		return ErrInvalidConfig{Field: "FlagrEndpoint", Reason: "cannot be empty"}
	}
	if c.RefreshInterval <= 0 {
		return ErrInvalidConfig{Field: "RefreshInterval", Reason: "must be positive"}
	}
	if c.RetryAttempts < 0 {
		return ErrInvalidConfig{Field: "RetryAttempts", Reason: "cannot be negative"}
	}
	if c.FallbackStrategy != "fail_closed" && c.FallbackStrategy != "fail_open" && c.FallbackStrategy != "last_known_good" {
		return ErrInvalidConfig{Field: "FallbackStrategy", Reason: "must be 'fail_closed', 'fail_open', or 'last_known_good'"}
	}
	return nil
}

// ErrInvalidConfig represents a configuration validation error
type ErrInvalidConfig struct {
	Field  string
	Reason string
}

func (e ErrInvalidConfig) Error() string {
	return fmt.Sprintf("invalid config field '%s': %s", e.Field, e.Reason)
}
