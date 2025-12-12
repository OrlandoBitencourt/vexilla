package cache

import (
	"fmt"
	"time"
)

// Config holds cache configuration
type Config struct {
	// Refresh behavior
	RefreshInterval time.Duration
	InitialTimeout  time.Duration

	// Fallback strategy when flag not found
	// Options: "fail_open", "fail_closed", "error"
	FallbackStrategy string

	// Circuit breaker
	CircuitBreakerThreshold int
	CircuitBreakerTimeout   time.Duration

	// 🔥 NEW: Flag Filtering (Resource Optimization)
	FilterConfig FilterConfig
}

// FilterConfig defines filtering rules for flags
type FilterConfig struct {
	// OnlyEnabled filters out disabled flags
	// When true, only stores flags with Enabled=true
	// This reduces memory usage and improves cache hit ratio
	OnlyEnabled bool

	// ServiceName is the current service identifier
	// Used to filter flags by tags
	ServiceName string

	// RequireServiceTag when true, only caches flags that have
	// the ServiceName in their tags list
	// This dramatically reduces memory footprint in microservice environments
	RequireServiceTag bool

	// AdditionalTags allows filtering by additional tag values
	// Useful for environment-specific flags (e.g., "production", "staging")
	AdditionalTags []string

	// TagMatchMode determines how tags are matched
	// "any": flag must have ANY of the tags
	// "all": flag must have ALL of the tags
	TagMatchMode string
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		RefreshInterval:         5 * time.Minute,
		InitialTimeout:          10 * time.Second,
		FallbackStrategy:        "fail_closed",
		CircuitBreakerThreshold: 3,
		CircuitBreakerTimeout:   30 * time.Second,
		FilterConfig: FilterConfig{
			OnlyEnabled:       true, // 🔥 NEW: Default to enabled only
			ServiceName:       "",
			RequireServiceTag: false,
			TagMatchMode:      "any",
		},
	}
}

// Validate validates the configuration
func (c Config) Validate() error {
	if c.RefreshInterval <= 0 {
		return fmt.Errorf("refresh interval must be positive")
	}

	if c.InitialTimeout <= 0 {
		return fmt.Errorf("initial timeout must be positive")
	}

	validStrategies := map[string]bool{
		"fail_open":   true,
		"fail_closed": true,
		"error":       true,
	}

	if !validStrategies[c.FallbackStrategy] {
		return fmt.Errorf("invalid fallback strategy: %s (must be 'fail_open', 'fail_closed', or 'error')", c.FallbackStrategy)
	}

	if c.CircuitBreakerThreshold < 1 {
		return fmt.Errorf("circuit breaker threshold must be at least 1")
	}

	// Validate filter config
	if err := c.FilterConfig.Validate(); err != nil {
		return fmt.Errorf("invalid filter config: %w", err)
	}

	return nil
}

// Validate validates the filter configuration
func (f FilterConfig) Validate() error {
	if f.RequireServiceTag && f.ServiceName == "" {
		return fmt.Errorf("service_name must be set when require_service_tag is true")
	}

	if f.TagMatchMode != "" && f.TagMatchMode != "any" && f.TagMatchMode != "all" {
		return fmt.Errorf("tag_match_mode must be 'any' or 'all'")
	}

	return nil
}

// ShouldCacheFlag determines if a flag should be cached based on filter rules
func (f FilterConfig) ShouldCacheFlag(flag FlagMetadata) bool {
	// Rule 1: OnlyEnabled filter
	if f.OnlyEnabled && !flag.Enabled {
		return false
	}

	// Rule 2: Service tag filter
	if f.RequireServiceTag {
		if !f.hasServiceTag(flag.Tags) {
			return false
		}
	}

	// Rule 3: Additional tags filter
	if len(f.AdditionalTags) > 0 {
		if !f.matchesAdditionalTags(flag.Tags) {
			return false
		}
	}

	return true
}

// hasServiceTag checks if flag has the service name in tags
func (f FilterConfig) hasServiceTag(tags []string) bool {
	for _, tag := range tags {
		if tag == f.ServiceName {
			return true
		}
	}
	return false
}

// matchesAdditionalTags checks if flag matches additional tag rules
func (f FilterConfig) matchesAdditionalTags(tags []string) bool {
	if len(f.AdditionalTags) == 0 {
		return true
	}

	tagMap := make(map[string]bool)
	for _, tag := range tags {
		tagMap[tag] = true
	}

	switch f.TagMatchMode {
	case "all":
		// Must have ALL additional tags
		for _, requiredTag := range f.AdditionalTags {
			if !tagMap[requiredTag] {
				return false
			}
		}
		return true

	case "any":
		// Must have ANY additional tag
		for _, requiredTag := range f.AdditionalTags {
			if tagMap[requiredTag] {
				return true
			}
		}
		return false

	default:
		return true
	}
}

// FlagMetadata contains minimal flag information for filtering
type FlagMetadata struct {
	Key     string
	Enabled bool
	Tags    []string
}

// String returns a human-readable description of the filter config
func (f FilterConfig) String() string {
	if !f.OnlyEnabled && !f.RequireServiceTag && len(f.AdditionalTags) == 0 {
		return "no filtering (all flags cached)"
	}

	filters := []string{}

	if f.OnlyEnabled {
		filters = append(filters, "enabled=true")
	}

	if f.RequireServiceTag {
		filters = append(filters, fmt.Sprintf("service=%s", f.ServiceName))
	}

	if len(f.AdditionalTags) > 0 {
		filters = append(filters, fmt.Sprintf("tags=%v (%s)", f.AdditionalTags, f.TagMatchMode))
	}

	return fmt.Sprintf("filtering: %v", filters)
}

// EstimateMemorySavings estimates memory savings from filtering
func (f FilterConfig) EstimateMemorySavings(totalFlags, filteredFlags int) MemorySavings {
	savedFlags := totalFlags - filteredFlags
	percentSaved := 0.0
	if totalFlags > 0 {
		percentSaved = float64(savedFlags) / float64(totalFlags) * 100
	}

	// Rough estimate: 1KB per flag
	const bytesPerFlag = 1024
	savedBytes := int64(savedFlags * bytesPerFlag)

	return MemorySavings{
		TotalFlags:      totalFlags,
		CachedFlags:     filteredFlags,
		FilteredFlags:   savedFlags,
		PercentFiltered: percentSaved,
		BytesSaved:      savedBytes,
	}
}

// MemorySavings represents the memory saved by filtering
type MemorySavings struct {
	TotalFlags      int
	CachedFlags     int
	FilteredFlags   int
	PercentFiltered float64
	BytesSaved      int64
}

// String returns a human-readable description of memory savings
func (m MemorySavings) String() string {
	mb := float64(m.BytesSaved) / 1024 / 1024
	return fmt.Sprintf("Cached %d/%d flags (%.1f%% filtered, ~%.2f MB saved)",
		m.CachedFlags, m.TotalFlags, m.PercentFiltered, mb)
}
