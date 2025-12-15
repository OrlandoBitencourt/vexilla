package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		want Config
	}{
		{
			name: "test default config",
			want: Config{
				MaxCost:        1 << 30, // 1GB
				NumCounters:    1e7,     // 10M counters
				BufferItems:    64,
				DefaultTTL:     5 * time.Minute,
				MetricsEnabled: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultConfig()
			if !assert.Equal(t, got, tt.want) {
				t.Errorf("DefaultConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
