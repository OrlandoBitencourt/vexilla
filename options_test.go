package vexilla

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWithFlagrEndpoint tests endpoint configuration
func TestWithFlagrEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		wantErr  bool
	}{
		{
			name:     "valid endpoint",
			endpoint: "http://localhost:18000",
			wantErr:  false,
		},
		{
			name:     "empty endpoint",
			endpoint: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &clientConfig{}
			opt := WithFlagrEndpoint(tt.endpoint)
			err := opt(cfg)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.endpoint, cfg.flagrEndpoint)
			}
		})
	}
}

// TestWithFallbackStrategy tests fallback strategy validation
func TestWithFallbackStrategy(t *testing.T) {
	tests := []struct {
		name     string
		strategy string
		wantErr  bool
	}{
		{
			name:     "fail_open strategy",
			strategy: "fail_open",
			wantErr:  false,
		},
		{
			name:     "fail_closed strategy",
			strategy: "fail_closed",
			wantErr:  false,
		},
		{
			name:     "error strategy",
			strategy: "error",
			wantErr:  false,
		},
		{
			name:     "invalid strategy",
			strategy: "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &clientConfig{}
			opt := WithFallbackStrategy(tt.strategy)
			err := opt(cfg)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.strategy, cfg.fallbackStrategy)
			}
		})
	}
}

// TestWithAdditionalTags tests tag configuration
func TestWithAdditionalTags(t *testing.T) {
	tests := []struct {
		name      string
		tags      []string
		matchMode string
		wantErr   bool
	}{
		{
			name:      "valid 'any' mode",
			tags:      []string{"production", "api"},
			matchMode: "any",
			wantErr:   false,
		},
		{
			name:      "valid 'all' mode",
			tags:      []string{"production", "api"},
			matchMode: "all",
			wantErr:   false,
		},
		{
			name:      "invalid match mode",
			tags:      []string{"production"},
			matchMode: "invalid",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &clientConfig{}
			opt := WithAdditionalTags(tt.tags, tt.matchMode)
			err := opt(cfg)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.tags, cfg.additionalTags)
				assert.Equal(t, tt.matchMode, cfg.tagMatchMode)
			}
		})
	}
}

// TestToCacheOptions tests converting clientConfig to cache options
func TestToCacheOptions(t *testing.T) {
	cfg := &clientConfig{
		flagrEndpoint:     "http://localhost:18000",
		refreshInterval:   5 * time.Minute,
		initialTimeout:    10 * time.Second,
		fallbackStrategy:  "fail_closed",
		circuitThreshold:  3,
		circuitTimeout:    30 * time.Second,
		onlyEnabled:       true,
		serviceName:       "test-service",
		requireServiceTag: true,
		additionalTags:    []string{"production"},
		tagMatchMode:      "any",
	}

	opts := cfg.toCacheOptions()
	assert.NotEmpty(t, opts)
}

// TestToCacheOptions_WithoutServiceTag tests conversion without service tag requirement
func TestToCacheOptions_WithoutServiceTag(t *testing.T) {
	cfg := &clientConfig{
		flagrEndpoint:     "http://localhost:18000",
		serviceName:       "test-service",
		requireServiceTag: false,
	}

	opts := cfg.toCacheOptions()
	assert.NotEmpty(t, opts)
}

// TestToCacheOptions_EmptyTagMatchMode tests conversion with empty tag match mode
func TestToCacheOptions_EmptyTagMatchMode(t *testing.T) {
	cfg := &clientConfig{
		flagrEndpoint:  "http://localhost:18000",
		additionalTags: []string{"production", "api"},
		tagMatchMode:   "",
	}

	opts := cfg.toCacheOptions()
	assert.NotEmpty(t, opts)
}

// TestWithConfig tests applying full configuration
func TestWithConfig(t *testing.T) {
	fullCfg := Config{
		Flagr: FlagrConfig{
			Endpoint:   "http://localhost:18000",
			APIKey:     "test-key",
			Timeout:    5 * time.Second,
			MaxRetries: 3,
		},
		Cache: CacheConfig{
			RefreshInterval: 10 * time.Minute,
			InitialTimeout:  20 * time.Second,
			Filter: FilterConfig{
				OnlyEnabled:       true,
				ServiceName:       "my-service",
				RequireServiceTag: true,
				AdditionalTags:    []string{"prod"},
				TagMatchMode:      "all",
			},
		},
		CircuitBreaker: CircuitBreakerConfig{
			Threshold: 5,
			Timeout:   60 * time.Second,
		},
		FallbackStrategy: "error",
	}

	cfg := &clientConfig{}
	opt := WithConfig(fullCfg)
	err := opt(cfg)

	require.NoError(t, err)
	assert.Equal(t, fullCfg.Flagr.Endpoint, cfg.flagrEndpoint)
	assert.Equal(t, fullCfg.Flagr.APIKey, cfg.flagrAPIKey)
	assert.Equal(t, fullCfg.Flagr.Timeout, cfg.flagrTimeout)
	assert.Equal(t, fullCfg.Flagr.MaxRetries, cfg.flagrMaxRetries)
	assert.Equal(t, fullCfg.Cache.RefreshInterval, cfg.refreshInterval)
	assert.Equal(t, fullCfg.Cache.InitialTimeout, cfg.initialTimeout)
	assert.Equal(t, fullCfg.Cache.Filter.OnlyEnabled, cfg.onlyEnabled)
	assert.Equal(t, fullCfg.Cache.Filter.ServiceName, cfg.serviceName)
	assert.Equal(t, fullCfg.Cache.Filter.RequireServiceTag, cfg.requireServiceTag)
	assert.Equal(t, fullCfg.Cache.Filter.AdditionalTags, cfg.additionalTags)
	assert.Equal(t, fullCfg.Cache.Filter.TagMatchMode, cfg.tagMatchMode)
	assert.Equal(t, fullCfg.CircuitBreaker.Threshold, cfg.circuitThreshold)
	assert.Equal(t, fullCfg.CircuitBreaker.Timeout, cfg.circuitTimeout)
	assert.Equal(t, fullCfg.FallbackStrategy, cfg.fallbackStrategy)
}

// TestWithFlagrAPIKey tests API key configuration
func TestWithFlagrAPIKey(t *testing.T) {
	cfg := &clientConfig{}
	opt := WithFlagrAPIKey("test-api-key")
	err := opt(cfg)

	require.NoError(t, err)
	assert.Equal(t, "test-api-key", cfg.flagrAPIKey)
}

// TestWithFlagrTimeout tests timeout configuration
func TestWithFlagrTimeout(t *testing.T) {
	cfg := &clientConfig{}
	opt := WithFlagrTimeout(15 * time.Second)
	err := opt(cfg)

	require.NoError(t, err)
	assert.Equal(t, 15*time.Second, cfg.flagrTimeout)
}

// TestWithFlagrMaxRetries tests max retries configuration
func TestWithFlagrMaxRetries(t *testing.T) {
	cfg := &clientConfig{}
	opt := WithFlagrMaxRetries(5)
	err := opt(cfg)

	require.NoError(t, err)
	assert.Equal(t, 5, cfg.flagrMaxRetries)
}

// TestWithRefreshInterval tests refresh interval configuration
func TestWithRefreshInterval(t *testing.T) {
	cfg := &clientConfig{}
	opt := WithRefreshInterval(3 * time.Minute)
	err := opt(cfg)

	require.NoError(t, err)
	assert.Equal(t, 3*time.Minute, cfg.refreshInterval)
}

// TestWithInitialTimeout tests initial timeout configuration
func TestWithInitialTimeout(t *testing.T) {
	cfg := &clientConfig{}
	opt := WithInitialTimeout(30 * time.Second)
	err := opt(cfg)

	require.NoError(t, err)
	assert.Equal(t, 30*time.Second, cfg.initialTimeout)
}

// TestWithCircuitBreaker tests circuit breaker configuration
func TestWithCircuitBreaker(t *testing.T) {
	cfg := &clientConfig{}
	opt := WithCircuitBreaker(5, 60*time.Second)
	err := opt(cfg)

	require.NoError(t, err)
	assert.Equal(t, 5, cfg.circuitThreshold)
	assert.Equal(t, 60*time.Second, cfg.circuitTimeout)
}

// TestWithOnlyEnabled tests only enabled flag filter
func TestWithOnlyEnabled(t *testing.T) {
	cfg := &clientConfig{}
	opt := WithOnlyEnabled(true)
	err := opt(cfg)

	require.NoError(t, err)
	assert.True(t, cfg.onlyEnabled)
}

// TestWithServiceTag tests service tag configuration
func TestWithServiceTag(t *testing.T) {
	cfg := &clientConfig{}
	opt := WithServiceTag("user-service")
	err := opt(cfg)

	require.NoError(t, err)
	assert.Equal(t, "user-service", cfg.serviceName)
	assert.True(t, cfg.requireServiceTag)
}
