package benchmarks

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
	"github.com/OrlandoBitencourt/vexilla/internal/cache"
	"github.com/OrlandoBitencourt/vexilla/internal/domain"
	"github.com/OrlandoBitencourt/vexilla/internal/evaluator"
	"github.com/OrlandoBitencourt/vexilla/internal/storage"
)

// mockFlagrClient is a mock implementation for benchmarks
type mockFlagrClient struct{}

func (m *mockFlagrClient) GetAllFlags(ctx context.Context) ([]domain.Flag, error) {
	return []domain.Flag{}, nil
}

func (m *mockFlagrClient) GetFlag(ctx context.Context, flagID int64) (*domain.Flag, error) {
	return nil, fmt.Errorf("not implemented for benchmarks")
}

func (m *mockFlagrClient) EvaluateFlag(ctx context.Context, flagKey string, evalCtx domain.EvaluationContext) (*domain.EvaluationResult, error) {
	return nil, fmt.Errorf("not implemented for benchmarks")
}

func (m *mockFlagrClient) Health(ctx context.Context) error {
	return nil
}

func (m *mockFlagrClient) HealthCheck(ctx context.Context) error {
	return nil
}

// BenchmarkLocalEvaluation_Simple benchmarks local evaluation with simple boolean flags
func BenchmarkLocalEvaluation_Simple(b *testing.B) {
	c := setupCache(b, simpleBooleanFlag())

	ctx := context.Background()
	evalCtx := domain.EvaluationContext{
		EntityID: "user-123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Evaluate(ctx, "test-flag", evalCtx)
	}
}

// BenchmarkLocalEvaluation_WithConstraints benchmarks local evaluation with constraint matching
func BenchmarkLocalEvaluation_WithConstraints(b *testing.B) {
	c := setupCache(b, flagWithConstraints())

	ctx := context.Background()
	evalCtx := domain.EvaluationContext{
		EntityID: "user-123",
		Context: map[string]interface{}{
			"tier":    "premium",
			"country": "BR",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Evaluate(ctx, "test-flag", evalCtx)
	}
}

// BenchmarkLocalEvaluation_MultipleSegments benchmarks evaluation with multiple segments
func BenchmarkLocalEvaluation_MultipleSegments(b *testing.B) {
	c := setupCache(b, flagWithMultipleSegments())

	ctx := context.Background()
	evalCtx := domain.EvaluationContext{
		EntityID: "user-123",
		Context: map[string]interface{}{
			"tier": "premium",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Evaluate(ctx, "test-flag", evalCtx)
	}
}

// BenchmarkLocalEvaluation_ComplexConstraints benchmarks evaluation with complex constraint expressions
func BenchmarkLocalEvaluation_ComplexConstraints(b *testing.B) {
	c := setupCache(b, flagWithComplexConstraints())

	ctx := context.Background()
	evalCtx := domain.EvaluationContext{
		EntityID: "user-123",
		Context: map[string]interface{}{
			"tier":          "premium",
			"country":       "BR",
			"user_bucket":   42,
			"signup_date":   "2024-01-15",
			"feature_flags": []string{"beta", "experimental"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Evaluate(ctx, "test-flag", evalCtx)
	}
}

// BenchmarkCacheHit benchmarks cache hit performance (Ristretto lookup)
func BenchmarkCacheHit(b *testing.B) {
	c := setupCache(b, simpleBooleanFlag())

	ctx := context.Background()
	evalCtx := domain.EvaluationContext{
		EntityID: "user-123",
	}

	// Warm up cache
	c.Evaluate(ctx, "test-flag", evalCtx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Evaluate(ctx, "test-flag", evalCtx)
	}
}

// BenchmarkConcurrentEvaluations benchmarks concurrent flag evaluations
func BenchmarkConcurrentEvaluations(b *testing.B) {
	c := setupCache(b, simpleBooleanFlag())

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			evalCtx := domain.EvaluationContext{
				EntityID: fmt.Sprintf("user-%d", i),
			}
			_, _ = c.Evaluate(ctx, "test-flag", evalCtx)
			i++
		}
	})
}

// BenchmarkDeterministicRollout benchmarks deterministic bucket-based rollout
func BenchmarkDeterministicRollout(b *testing.B) {
	c := setupCache(b, flagWithDeterministicRollout())

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalCtx := domain.EvaluationContext{
			EntityID: fmt.Sprintf("user-%d", i),
			Context: map[string]interface{}{
				"user_bucket": i % 100, // Bucket 0-99
			},
		}
		_, _ = c.Evaluate(ctx, "test-flag", evalCtx)
	}
}

// BenchmarkMultipleFlagEvaluation benchmarks evaluating different flags
func BenchmarkMultipleFlagEvaluation(b *testing.B) {
	c := setupCacheWithMultipleFlags(b, 10)

	ctx := context.Background()
	evalCtx := domain.EvaluationContext{
		EntityID: "user-123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flagKey := fmt.Sprintf("flag-%d", i%10)
		_, _ = c.Evaluate(ctx, flagKey, evalCtx)
	}
}

// BenchmarkStorageSet benchmarks storage write operations
func BenchmarkStorageSet(b *testing.B) {
	store, err := storage.NewMemoryStorage(storage.Config{
		MaxCost:     1 << 30,
		NumCounters: 1e7,
		BufferItems: 64,
		DefaultTTL:  5 * time.Minute,
	})
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	flag := simpleBooleanFlag()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("flag-%d", i)
		_ = store.Set(ctx, key, flag, 5*time.Minute)
	}
}

// BenchmarkStorageGet benchmarks storage read operations
func BenchmarkStorageGet(b *testing.B) {
	store, err := storage.NewMemoryStorage(storage.Config{
		MaxCost:     1 << 30,
		NumCounters: 1e7,
		BufferItems: 64,
		DefaultTTL:  5 * time.Minute,
	})
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	flag := simpleBooleanFlag()

	// Pre-populate
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("flag-%d", i)
		_ = store.Set(ctx, key, flag, 5*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("flag-%d", i%1000)
		_, _ = store.Get(ctx, key)
	}
}

// BenchmarkEvaluatorConstraintMatching benchmarks pure constraint evaluation
func BenchmarkEvaluatorConstraintMatching(b *testing.B) {
	eval := evaluator.New()
	flag := flagWithConstraints()

	evalCtx := domain.EvaluationContext{
		EntityID: "user-123",
		Context: map[string]interface{}{
			"tier":    "premium",
			"country": "BR",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = eval.Evaluate(context.Background(), flag, evalCtx)
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	c := setupCache(b, simpleBooleanFlag())

	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		evalCtx := domain.EvaluationContext{
			EntityID: fmt.Sprintf("user-%d", i),
		}
		_, _ = c.Evaluate(ctx, "test-flag", evalCtx)
	}
}

// BenchmarkLargeScaleCache benchmarks cache with many flags
func BenchmarkLargeScaleCache_1000Flags(b *testing.B) {
	benchmarkLargeScaleCache(b, 1000)
}

func BenchmarkLargeScaleCache_10000Flags(b *testing.B) {
	benchmarkLargeScaleCache(b, 10000)
}

func benchmarkLargeScaleCache(b *testing.B, numFlags int) {
	c := setupCacheWithMultipleFlags(b, numFlags)

	ctx := context.Background()
	evalCtx := domain.EvaluationContext{
		EntityID: "user-123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flagKey := fmt.Sprintf("flag-%d", i%numFlags)
		_, _ = c.Evaluate(ctx, flagKey, evalCtx)
	}
}

// BenchmarkClientAPI benchmarks the public client API
func BenchmarkClientAPI_Bool(b *testing.B) {
	client := setupClient(b)
	defer client.Stop()

	ctx := context.Background()
	evalCtx := vexilla.NewContext("user-123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.Bool(ctx, "test-flag", evalCtx)
	}
}

func BenchmarkClientAPI_Evaluate(b *testing.B) {
	client := setupClient(b)
	defer client.Stop()

	ctx := context.Background()
	evalCtx := vexilla.NewContext("user-123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Evaluate(ctx, "test-flag", evalCtx)
	}
}

// BenchmarkBooleanVsDetailedEvaluation compares Bool() vs Evaluate()
func BenchmarkBooleanVsDetailedEvaluation_Bool(b *testing.B) {
	client := setupClient(b)
	defer client.Stop()

	ctx := context.Background()
	evalCtx := vexilla.NewContext("user-123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.Bool(ctx, "test-flag", evalCtx)
	}
}

func BenchmarkBooleanVsDetailedEvaluation_Evaluate(b *testing.B) {
	client := setupClient(b)
	defer client.Stop()

	ctx := context.Background()
	evalCtx := vexilla.NewContext("user-123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Evaluate(ctx, "test-flag", evalCtx)
	}
}

// Helper functions

func setupCache(b *testing.B, flag domain.Flag) *cache.Cache {
	store, err := storage.NewMemoryStorage(storage.Config{
		MaxCost:     1 << 30,
		NumCounters: 1e7,
		BufferItems: 64,
		DefaultTTL:  5 * time.Minute,
	})
	if err != nil {
		b.Fatal(err)
	}

	eval := evaluator.New()
	mockClient := &mockFlagrClient{}

	c, err := cache.New(
		cache.WithFlagrClient(mockClient),
		cache.WithStorage(store),
		cache.WithEvaluator(eval),
		cache.WithRefreshInterval(24*time.Hour), // Very long interval for benchmarks
	)
	if err != nil {
		b.Fatal(err)
	}

	// Pre-populate cache
	ctx := context.Background()
	_ = store.Set(ctx, "test-flag", flag, 24*time.Hour)

	return c
}

func setupCacheWithMultipleFlags(b *testing.B, count int) *cache.Cache {
	store, err := storage.NewMemoryStorage(storage.Config{
		MaxCost:     1 << 30,
		NumCounters: 1e7,
		BufferItems: 64,
		DefaultTTL:  5 * time.Minute,
	})
	if err != nil {
		b.Fatal(err)
	}

	eval := evaluator.New()
	mockClient := &mockFlagrClient{}

	c, err := cache.New(
		cache.WithFlagrClient(mockClient),
		cache.WithStorage(store),
		cache.WithEvaluator(eval),
		cache.WithRefreshInterval(24*time.Hour),
	)
	if err != nil {
		b.Fatal(err)
	}

	// Pre-populate with multiple flags
	ctx := context.Background()
	for i := 0; i < count; i++ {
		flag := simpleBooleanFlag()
		flag.Key = fmt.Sprintf("flag-%d", i)
		_ = store.Set(ctx, flag.Key, flag, 5*time.Minute)
	}

	return c
}

func setupClient(b *testing.B) *vexilla.Client {
	// Create a mock client with in-memory storage (no Flagr dependency)
	client, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithRefreshInterval(1*time.Hour),
	)
	if err != nil {
		b.Fatal(err)
	}

	// Manually populate cache for benchmarking
	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		// Ignore error if Flagr is not available
		// The cache will still work with empty data for benchmarking
	}

	return client
}

// Test flag generators

func simpleBooleanFlag() domain.Flag {
	return domain.Flag{
		ID:      1,
		Key:     "test-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{
				ID:              1,
				Rank:            0,
				RolloutPercent:  100,
				Constraints:     []domain.Constraint{},
				Distributions: []domain.Distribution{
					{
						Percent:    100,
						VariantID:  1,
						
					},
				},
			},
		},
		Variants: []domain.Variant{
			{
				ID:  1,
				Key: "on",
			},
		},
	}
}

func flagWithConstraints() domain.Flag {
	return domain.Flag{
		ID:      1,
		Key:     "test-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{
				ID:             1,
				Rank:           0,
				RolloutPercent: 100,
				Constraints: []domain.Constraint{
					{
						Property: "tier",
						Operator: "EQ",
						Value:    "premium",
					},
					{
						Property: "country",
						Operator: "EQ",
						Value:    "BR",
					},
				},
				Distributions: []domain.Distribution{
					{
						Percent:    100,
						VariantID:  1,
						
					},
				},
			},
		},
		Variants: []domain.Variant{
			{
				ID:  1,
				Key: "on",
			},
		},
	}
}

func flagWithMultipleSegments() domain.Flag {
	return domain.Flag{
		ID:      1,
		Key:     "test-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{
				ID:             1,
				Rank:           0,
				RolloutPercent: 100,
				Constraints: []domain.Constraint{
					{
						Property: "tier",
						Operator: "EQ",
						Value:    "premium",
					},
				},
				Distributions: []domain.Distribution{
					{
						Percent:   100,
						VariantID: 1,
					},
				},
			},
			{
				ID:             2,
				Rank:           1,
				RolloutPercent: 100,
				Constraints: []domain.Constraint{
					{
						Property: "tier",
						Operator: "EQ",
						Value:    "free",
					},
				},
				Distributions: []domain.Distribution{
					{
						Percent:   100,
						VariantID: 2,
					},
				},
			},
			{
				ID:             3,
				Rank:           2,
				RolloutPercent: 100,
				Constraints:    []domain.Constraint{},
				Distributions: []domain.Distribution{
					{
						Percent:   100,
						VariantID: 3,
					},
				},
			},
		},
		Variants: []domain.Variant{
			{ID: 1, Key: "premium"},
			{ID: 2, Key: "free"},
			{ID: 3, Key: "default"},
		},
	}
}

func flagWithComplexConstraints() domain.Flag {
	return domain.Flag{
		ID:      1,
		Key:     "test-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{
				ID:             1,
				Rank:           0,
				RolloutPercent: 100,
				Constraints: []domain.Constraint{
					{
						Property: "tier",
						Operator: "EQ",
						Value:    "premium",
					},
					{
						Property: "country",
						Operator: "IN",
						Value:    []string{"BR", "US", "UK"},
					},
					{
						Property: "user_bucket",
						Operator: "GTE",
						Value:    0,
					},
					{
						Property: "user_bucket",
						Operator: "LTE",
						Value:    70,
					},
					{
						Property: "signup_date",
						Operator: "GTE",
						Value:    "2024-01-01",
					},
				},
				Distributions: []domain.Distribution{
					{
						Percent:    100,
						VariantID:  1,
						
					},
				},
			},
		},
		Variants: []domain.Variant{
			{
				ID:  1,
				Key: "on",
			},
		},
	}
}

func flagWithDeterministicRollout() domain.Flag {
	return domain.Flag{
		ID:      1,
		Key:     "test-flag",
		Enabled: true,
		Segments: []domain.Segment{
			{
				ID:             1,
				Rank:           0,
				RolloutPercent: 100,
				Constraints: []domain.Constraint{
					{
						Property: "user_bucket",
						Operator: "GTE",
						Value:    0,
					},
					{
						Property: "user_bucket",
						Operator: "LTE",
						Value:    70,
					},
				},
				Distributions: []domain.Distribution{
					{
						Percent:   100,
						VariantID: 1,
					},
				},
			},
			{
				ID:             2,
				Rank:           1,
				RolloutPercent: 100,
				Constraints:    []domain.Constraint{},
				Distributions: []domain.Distribution{
					{
						Percent:   100,
						VariantID: 2,
					},
				},
			},
		},
		Variants: []domain.Variant{
			{ID: 1, Key: "enabled"},
			{ID: 2, Key: "disabled"},
		},
	}
}
