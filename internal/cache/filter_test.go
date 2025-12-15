package cache

import (
	"context"
	"testing"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
	"github.com/OrlandoBitencourt/vexilla/internal/evaluator"
	"github.com/OrlandoBitencourt/vexilla/internal/flagr"
	"github.com/OrlandoBitencourt/vexilla/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterConfig_ShouldCacheFlag_OnlyEnabled(t *testing.T) {
	tests := []struct {
		name     string
		filter   FilterConfig
		flag     FlagMetadata
		expected bool
	}{
		{
			name: "enabled flag - should cache",
			filter: FilterConfig{
				OnlyEnabled: true,
			},
			flag: FlagMetadata{
				Key:     "test-flag",
				Enabled: true,
				Tags:    []string{},
			},
			expected: true,
		},
		{
			name: "disabled flag - should not cache",
			filter: FilterConfig{
				OnlyEnabled: true,
			},
			flag: FlagMetadata{
				Key:     "test-flag",
				Enabled: false,
				Tags:    []string{},
			},
			expected: false,
		},
		{
			name: "disabled flag - no filter - should cache",
			filter: FilterConfig{
				OnlyEnabled: false,
			},
			flag: FlagMetadata{
				Key:     "test-flag",
				Enabled: false,
				Tags:    []string{},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.filter.ShouldCacheFlag(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterConfig_ShouldCacheFlag_ServiceTag(t *testing.T) {
	tests := []struct {
		name     string
		filter   FilterConfig
		flag     FlagMetadata
		expected bool
	}{
		{
			name: "has service tag - should cache",
			filter: FilterConfig{
				ServiceName:       "user-service",
				RequireServiceTag: true,
			},
			flag: FlagMetadata{
				Key:     "test-flag",
				Enabled: true,
				Tags:    []string{"user-service", "production"},
			},
			expected: true,
		},
		{
			name: "missing service tag - should not cache",
			filter: FilterConfig{
				ServiceName:       "user-service",
				RequireServiceTag: true,
			},
			flag: FlagMetadata{
				Key:     "test-flag",
				Enabled: true,
				Tags:    []string{"payment-service", "production"},
			},
			expected: false,
		},
		{
			name: "no service tag required - should cache",
			filter: FilterConfig{
				ServiceName:       "user-service",
				RequireServiceTag: false,
			},
			flag: FlagMetadata{
				Key:     "test-flag",
				Enabled: true,
				Tags:    []string{"payment-service"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.filter.ShouldCacheFlag(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterConfig_ShouldCacheFlag_AdditionalTags(t *testing.T) {
	tests := []struct {
		name     string
		filter   FilterConfig
		flag     FlagMetadata
		expected bool
	}{
		{
			name: "has any additional tag - should cache",
			filter: FilterConfig{
				AdditionalTags: []string{"production", "staging"},
				TagMatchMode:   "any",
			},
			flag: FlagMetadata{
				Key:     "test-flag",
				Enabled: true,
				Tags:    []string{"user-service", "production"},
			},
			expected: true,
		},
		{
			name: "missing all additional tags (any mode) - should not cache",
			filter: FilterConfig{
				AdditionalTags: []string{"production", "staging"},
				TagMatchMode:   "any",
			},
			flag: FlagMetadata{
				Key:     "test-flag",
				Enabled: true,
				Tags:    []string{"user-service", "development"},
			},
			expected: false,
		},
		{
			name: "has all additional tags - should cache",
			filter: FilterConfig{
				AdditionalTags: []string{"production", "critical"},
				TagMatchMode:   "all",
			},
			flag: FlagMetadata{
				Key:     "test-flag",
				Enabled: true,
				Tags:    []string{"user-service", "production", "critical"},
			},
			expected: true,
		},
		{
			name: "missing one additional tag (all mode) - should not cache",
			filter: FilterConfig{
				AdditionalTags: []string{"production", "critical"},
				TagMatchMode:   "all",
			},
			flag: FlagMetadata{
				Key:     "test-flag",
				Enabled: true,
				Tags:    []string{"user-service", "production"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.filter.ShouldCacheFlag(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterConfig_ShouldCacheFlag_Combined(t *testing.T) {
	tests := []struct {
		name     string
		filter   FilterConfig
		flag     FlagMetadata
		expected bool
	}{
		{
			name: "all conditions met - should cache",
			filter: FilterConfig{
				OnlyEnabled:       true,
				ServiceName:       "user-service",
				RequireServiceTag: true,
				AdditionalTags:    []string{"production"},
				TagMatchMode:      "any",
			},
			flag: FlagMetadata{
				Key:     "test-flag",
				Enabled: true,
				Tags:    []string{"user-service", "production"},
			},
			expected: true,
		},
		{
			name: "disabled flag - should not cache",
			filter: FilterConfig{
				OnlyEnabled:       true,
				ServiceName:       "user-service",
				RequireServiceTag: true,
			},
			flag: FlagMetadata{
				Key:     "test-flag",
				Enabled: false,
				Tags:    []string{"user-service", "production"},
			},
			expected: false,
		},
		{
			name: "missing service tag - should not cache",
			filter: FilterConfig{
				OnlyEnabled:       true,
				ServiceName:       "user-service",
				RequireServiceTag: true,
			},
			flag: FlagMetadata{
				Key:     "test-flag",
				Enabled: true,
				Tags:    []string{"payment-service", "production"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.filter.ShouldCacheFlag(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCache_RefreshFlags_WithFiltering(t *testing.T) {
	mockFlagr := flagr.NewMockClient()
	mockStorage := storage.NewMockStorage()
	eval := evaluator.New()

	// Add multiple flags to mock Flagr
	flags := []domain.Flag{
		{
			ID:      1,
			Key:     "user-feature",
			Enabled: true,
			Tags:    []domain.Tag{{Value: "user-service"}, {Value: "production"}},
			Segments: []domain.Segment{
				{RolloutPercent: 100, Distributions: []domain.Distribution{{VariantID: 1, Percent: 100}}},
			},
			Variants: []domain.Variant{{ID: 1, Key: "on"}},
		},
		{
			ID:      2,
			Key:     "payment-feature",
			Enabled: true,
			Tags:    []domain.Tag{{Value: "payment-service"}, {Value: "production"}},
			Segments: []domain.Segment{
				{RolloutPercent: 100, Distributions: []domain.Distribution{{VariantID: 1, Percent: 100}}},
			},
			Variants: []domain.Variant{{ID: 1, Key: "on"}},
		},
		{
			ID:      3,
			Key:     "disabled-feature",
			Enabled: false,
			Tags:    []domain.Tag{{Value: "user-service"}},
			Segments: []domain.Segment{
				{RolloutPercent: 100, Distributions: []domain.Distribution{{VariantID: 1, Percent: 100}}},
			},
			Variants: []domain.Variant{{ID: 1, Key: "on"}},
		},
	}

	for _, flag := range flags {
		mockFlagr.AddFlag(flag)
	}

	// Create cache with filtering
	c, err := New(
		WithFlagrClient(mockFlagr),
		WithStorage(mockStorage),
		WithEvaluator(eval),
		WithOnlyEnabled(true),
		WithServiceTag("user-service", true),
	)
	require.NoError(t, err)

	// Refresh flags
	ctx := context.Background()
	err = c.refreshFlags(ctx)
	require.NoError(t, err)

	// Verify correct flag was cached
	mockStorage.AssertCalled(t, "Set", 1)

	cachedFlag := mockStorage.GetFlag("user-feature")
	assert.NotNil(t, cachedFlag, "user-feature should be cached")

	assert.Nil(t, mockStorage.GetFlag("payment-feature"), "payment-feature should not be cached")
	assert.Nil(t, mockStorage.GetFlag("disabled-feature"), "disabled-feature should not be cached")
}

func TestCache_MemorySavings(t *testing.T) {
	filter := FilterConfig{
		OnlyEnabled:       true,
		ServiceName:       "user-service",
		RequireServiceTag: true,
	}

	savings := filter.EstimateMemorySavings(100, 10)

	assert.Equal(t, 100, savings.TotalFlags)
	assert.Equal(t, 10, savings.CachedFlags)
	assert.Equal(t, 90, savings.FilteredFlags)
	assert.Equal(t, 90.0, savings.PercentFiltered)
	assert.Equal(t, int64(90*1024), savings.BytesSaved)

	str := savings.String()
	assert.Contains(t, str, "10/100")
	assert.Contains(t, str, "90.0%")
}

func TestCache_Start_WithFiltering(t *testing.T) {
	mockFlagr := flagr.NewMockClient()
	mockStorage := storage.NewMockStorage()
	eval := evaluator.New()

	// Add test flags
	mockFlagr.AddFlag(domain.Flag{
		ID:      1,
		Key:     "enabled-flag",
		Enabled: true,
		Tags:    []domain.Tag{{Value: "test-service"}},
		Segments: []domain.Segment{
			{RolloutPercent: 100, Distributions: []domain.Distribution{{VariantID: 1, Percent: 100}}},
		},
		Variants: []domain.Variant{{ID: 1, Key: "on"}},
	})

	mockFlagr.AddFlag(domain.Flag{
		ID:      2,
		Key:     "disabled-flag",
		Enabled: false,
		Tags:    []domain.Tag{{Value: "test-service"}},
	})

	c, err := New(
		WithFlagrClient(mockFlagr),
		WithStorage(mockStorage),
		WithEvaluator(eval),
		WithOnlyEnabled(true),
	)
	require.NoError(t, err)

	ctx := context.Background()
	err = c.Start(ctx)
	require.NoError(t, err)

	// Should only cache enabled flag
	mockStorage.AssertCalled(t, "Set", 1)

	c.Stop()
}

func TestFilterConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		filter    FilterConfig
		expectErr bool
	}{
		{
			name: "valid config",
			filter: FilterConfig{
				OnlyEnabled:       true,
				ServiceName:       "test-service",
				RequireServiceTag: true,
				TagMatchMode:      "any",
			},
			expectErr: false,
		},
		{
			name: "require service tag without service name",
			filter: FilterConfig{
				RequireServiceTag: true,
				ServiceName:       "",
			},
			expectErr: true,
		},
		{
			name: "invalid tag match mode",
			filter: FilterConfig{
				TagMatchMode: "invalid",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func BenchmarkFilterConfig_ShouldCacheFlag(b *testing.B) {
	filter := FilterConfig{
		OnlyEnabled:       true,
		ServiceName:       "user-service",
		RequireServiceTag: true,
		AdditionalTags:    []string{"production", "critical"},
		TagMatchMode:      "any",
	}

	flag := FlagMetadata{
		Key:     "test-flag",
		Enabled: true,
		Tags:    []string{"user-service", "production", "feature-x"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.ShouldCacheFlag(flag)
	}
}
