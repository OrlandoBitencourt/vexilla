package vexilla

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, 5*time.Minute, config.RefreshInterval)
	assert.Equal(t, 10*time.Second, config.InitialTimeout)
	assert.Equal(t, "fail_closed", config.FallbackStrategy)
	assert.True(t, config.PersistenceEnabled)
	assert.Equal(t, int64(1<<30), config.CacheMaxCost)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		shouldErr bool
		errField  string
	}{
		{
			name: "valid config",
			config: Config{
				FlagrEndpoint:    "http://flagr.local",
				RefreshInterval:  5 * time.Minute,
				RetryAttempts:    3,
				FallbackStrategy: "fail_closed",
			},
			shouldErr: false,
		},
		{
			name: "missing endpoint",
			config: Config{
				RefreshInterval:  5 * time.Minute,
				FallbackStrategy: "fail_closed",
			},
			shouldErr: true,
			errField:  "FlagrEndpoint",
		},
		{
			name: "invalid refresh interval",
			config: Config{
				FlagrEndpoint:    "http://flagr.local",
				RefreshInterval:  0,
				FallbackStrategy: "fail_closed",
			},
			shouldErr: true,
			errField:  "RefreshInterval",
		},
		{
			name: "invalid fallback strategy",
			config: Config{
				FlagrEndpoint:    "http://flagr.local",
				RefreshInterval:  5 * time.Minute,
				FallbackStrategy: "invalid",
			},
			shouldErr: true,
			errField:  "FallbackStrategy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.shouldErr {
				assert.Error(t, err)
				if tt.errField != "" {
					configErr, ok := err.(ErrInvalidConfig)
					assert.True(t, ok)
					assert.Equal(t, tt.errField, configErr.Field)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
