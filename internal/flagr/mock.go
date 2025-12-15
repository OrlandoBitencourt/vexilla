package flagr

import (
	"context"
	"sync"

	"github.com/OrlandoBitencourt/vexilla/internal/domain"
)

// MockClient is a mock implementation of Client for testing
type MockClient struct {
	mu sync.RWMutex

	// Stored flags
	flags map[int64]domain.Flag

	// Mock behaviors
	GetAllFlagsFunc  func(ctx context.Context) ([]domain.Flag, error)
	GetFlagFunc      func(ctx context.Context, flagID int64) (*domain.Flag, error)
	EvaluateFlagFunc func(ctx context.Context, flagKey string, evalCtx domain.EvaluationContext) (*domain.EvaluationResult, error)
	HealthCheckFunc  func(ctx context.Context) error

	// Call tracking
	GetAllFlagsCalls  int
	GetFlagCalls      int
	EvaluateFlagCalls int
	HealthCheckCalls  int
}

// NewMockClient creates a new mock client
func NewMockClient() *MockClient {
	return &MockClient{
		flags: make(map[int64]domain.Flag),
	}
}

// AddFlag adds a flag to the mock
func (m *MockClient) AddFlag(flag domain.Flag) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flags[flag.ID] = flag
}

// GetAllFlags returns all flags
func (m *MockClient) GetAllFlags(ctx context.Context) ([]domain.Flag, error) {
	m.mu.Lock()
	m.GetAllFlagsCalls++
	m.mu.Unlock()

	if m.GetAllFlagsFunc != nil {
		return m.GetAllFlagsFunc(ctx)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	flags := make([]domain.Flag, 0, len(m.flags))
	for _, flag := range m.flags {
		flags = append(flags, flag)
	}

	return flags, nil
}

// GetFlag returns a single flag
func (m *MockClient) GetFlag(ctx context.Context, flagID int64) (*domain.Flag, error) {
	m.mu.Lock()
	m.GetFlagCalls++
	m.mu.Unlock()

	if m.GetFlagFunc != nil {
		return m.GetFlagFunc(ctx, flagID)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	flag, ok := m.flags[flagID]
	if !ok {
		return nil, domain.NewNotFoundError("flag", string(rune(flagID)))
	}

	return &flag, nil
}

// EvaluateFlag evaluates a flag
func (m *MockClient) EvaluateFlag(ctx context.Context, flagKey string, evalCtx domain.EvaluationContext) (*domain.EvaluationResult, error) {
	m.mu.Lock()
	m.EvaluateFlagCalls++
	m.mu.Unlock()

	if m.EvaluateFlagFunc != nil {
		return m.EvaluateFlagFunc(ctx, flagKey, evalCtx)
	}

	// Default implementation: find flag and return first variant
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, flag := range m.flags {
		if flag.Key == flagKey {
			if len(flag.Variants) == 0 {
				return nil, domain.NewEvaluationError(flagKey, "no variants", nil)
			}

			result := &domain.EvaluationResult{
				FlagID:            flag.ID,
				FlagKey:           flag.Key,
				VariantID:         flag.Variants[0].ID,
				VariantKey:        flag.Variants[0].Key,
				VariantAttachment: flag.Variants[0].Attachment,
				EvaluationReason:  "mock evaluation",
			}

			return result, nil
		}
	}

	return nil, domain.NewNotFoundError("flag", flagKey)
}

// HealthCheck performs health check
func (m *MockClient) HealthCheck(ctx context.Context) error {
	m.mu.Lock()
	m.HealthCheckCalls++
	m.mu.Unlock()

	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc(ctx)
	}

	return nil
}

// Reset resets the mock state
func (m *MockClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.flags = make(map[int64]domain.Flag)
	m.GetAllFlagsCalls = 0
	m.GetFlagCalls = 0
	m.EvaluateFlagCalls = 0
	m.HealthCheckCalls = 0
}

// AssertCalled asserts methods were called expected times
func (m *MockClient) AssertCalled(t interface{ Errorf(string, ...interface{}) }, method string, expected int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var actual int
	switch method {
	case "GetAllFlags":
		actual = m.GetAllFlagsCalls
	case "GetFlag":
		actual = m.GetFlagCalls
	case "EvaluateFlag":
		actual = m.EvaluateFlagCalls
	case "HealthCheck":
		actual = m.HealthCheckCalls
	default:
		t.Errorf("unknown method: %s", method)
		return
	}

	if actual != expected {
		t.Errorf("%s called %d times, expected %d", method, actual, expected)
	}
}
