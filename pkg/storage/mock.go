// pkg/storage/mock.go
package storage

import (
	"context"
	"sync"
	"time"

	"github.com/OrlandoBitencourt/vexilla/pkg/domain"
)

// MockStorage is a mock implementation of Storage for testing
type MockStorage struct {
	mu    sync.RWMutex
	flags map[string]domain.Flag

	// Mock behaviors
	GetFunc     func(ctx context.Context, key string) (*domain.Flag, error)
	SetFunc     func(ctx context.Context, key string, flag domain.Flag, ttl time.Duration) error
	DeleteFunc  func(ctx context.Context, key string) error
	ClearFunc   func(ctx context.Context) error
	ListFunc    func(ctx context.Context) ([]string, error)
	MetricsFunc func() Metrics
	CloseFunc   func() error

	// Call tracking
	GetCalls    int
	SetCalls    int
	DeleteCalls int
	ClearCalls  int
	ListCalls   int
	CloseCalls  int
}

// NewMockStorage creates a new mock storage
func NewMockStorage() *MockStorage {
	return &MockStorage{
		flags: make(map[string]domain.Flag),
	}
}

// Get retrieves a flag by key
func (m *MockStorage) Get(ctx context.Context, key string) (*domain.Flag, error) {
	m.mu.Lock()
	m.GetCalls++
	m.mu.Unlock()

	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	flag, ok := m.flags[key]
	if !ok {
		return nil, domain.NewNotFoundError("flag", key)
	}

	return &flag, nil
}

// Set stores a flag
func (m *MockStorage) Set(ctx context.Context, key string, flag domain.Flag, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SetCalls++

	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, flag, ttl)
	}

	m.flags[key] = flag
	return nil
}

// Delete removes a flag
func (m *MockStorage) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DeleteCalls++

	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, key)
	}

	delete(m.flags, key)
	return nil
}

// Clear removes all flags
func (m *MockStorage) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ClearCalls++

	if m.ClearFunc != nil {
		return m.ClearFunc(ctx)
	}

	m.flags = make(map[string]domain.Flag)
	return nil
}

// List returns all flag keys
func (m *MockStorage) List(ctx context.Context) ([]string, error) {
	m.mu.Lock()
	m.ListCalls++
	m.mu.Unlock()

	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.flags))
	for key := range m.flags {
		keys = append(keys, key)
	}

	return keys, nil
}

// Metrics returns storage metrics
func (m *MockStorage) Metrics() Metrics {
	if m.MetricsFunc != nil {
		return m.MetricsFunc()
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	return Metrics{
		Size: int64(len(m.flags)),
	}
}

// Close closes the storage
func (m *MockStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CloseCalls++

	if m.CloseFunc != nil {
		return m.CloseFunc()
	}

	return nil
}

// Reset resets the mock state
func (m *MockStorage) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.flags = make(map[string]domain.Flag)
	m.GetCalls = 0
	m.SetCalls = 0
	m.DeleteCalls = 0
	m.ClearCalls = 0
	m.ListCalls = 0
	m.CloseCalls = 0
}

// AddFlag adds a flag to the mock
func (m *MockStorage) AddFlag(flag domain.Flag) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.flags[flag.Key] = flag
}

// GetFlag gets a flag from the mock (without tracking)
func (m *MockStorage) GetFlag(key string) *domain.Flag {
	m.mu.RLock()
	defer m.mu.RUnlock()

	flag, ok := m.flags[key]
	if !ok {
		return nil
	}

	return &flag
}

// AssertCalled asserts methods were called expected times
func (m *MockStorage) AssertCalled(t interface{ Errorf(string, ...interface{}) }, method string, expected int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var actual int
	switch method {
	case "Get":
		actual = m.GetCalls
	case "Set":
		actual = m.SetCalls
	case "Delete":
		actual = m.DeleteCalls
	case "Clear":
		actual = m.ClearCalls
	case "List":
		actual = m.ListCalls
	case "Close":
		actual = m.CloseCalls
	default:
		t.Errorf("unknown method: %s", method)
		return
	}

	if actual != expected {
		t.Errorf("%s called %d times, expected %d", method, actual, expected)
	}
}
