package circuit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	config := DefaultConfig()
	breaker := New(config)

	assert.NotNil(t, breaker)
	assert.Equal(t, StateClosed, breaker.GetState())
}

func TestBreaker_ClosedState(t *testing.T) {
	breaker := New(DefaultConfig())
	ctx := context.Background()

	// Successful calls should keep circuit closed
	for i := 0; i < 10; i++ {
		err := breaker.Call(ctx, func() error {
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, StateClosed, breaker.GetState())
	}

	stats := breaker.GetStats()
	assert.Equal(t, int64(10), stats.TotalRequests)
	assert.Equal(t, int64(10), stats.TotalSuccesses)
}

func TestBreaker_OpenState(t *testing.T) {
	config := DefaultConfig()
	config.MaxFailures = 3
	breaker := New(config)
	ctx := context.Background()

	testErr := errors.New("test error")

	// Trigger failures to open circuit
	for i := 0; i < 3; i++ {
		err := breaker.Call(ctx, func() error {
			return testErr
		})
		assert.Error(t, err)
	}

	// Circuit should be open
	assert.Equal(t, StateOpen, breaker.GetState())

	// Next call should be rejected
	err := breaker.Call(ctx, func() error {
		return nil
	})

	assert.Error(t, err)
	assert.True(t, IsCircuitOpen(err))

	stats := breaker.GetStats()
	assert.Equal(t, int64(1), stats.TotalRejections)
}

func TestBreaker_HalfOpenState(t *testing.T) {
	config := DefaultConfig()
	config.MaxFailures = 2
	config.Timeout = 100 * time.Millisecond
	breaker := New(config)
	ctx := context.Background()

	// Open the circuit
	for i := 0; i < 2; i++ {
		breaker.Call(ctx, func() error {
			return errors.New("fail")
		})
	}

	assert.Equal(t, StateOpen, breaker.GetState())

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Next call should transition to half-open
	err := breaker.Call(ctx, func() error {
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, StateHalfOpen, breaker.GetState())
}

func TestBreaker_HalfOpenToClose(t *testing.T) {
	config := DefaultConfig()
	config.MaxFailures = 2
	config.Timeout = 50 * time.Millisecond
	breaker := New(config)
	ctx := context.Background()

	// Open the circuit
	for i := 0; i < 2; i++ {
		breaker.Call(ctx, func() error {
			return errors.New("fail")
		})
	}

	// Wait and transition to half-open
	time.Sleep(60 * time.Millisecond)

	// Make 2 successful calls in half-open state
	breaker.Call(ctx, func() error {
		return nil
	})

	breaker.Call(ctx, func() error {
		return nil
	})

	// Should transition back to closed
	assert.Equal(t, StateClosed, breaker.GetState())
}

func TestBreaker_HalfOpenToOpen(t *testing.T) {
	config := DefaultConfig()
	config.MaxFailures = 2
	config.Timeout = 50 * time.Millisecond
	breaker := New(config)
	ctx := context.Background()

	// Open the circuit
	for i := 0; i < 2; i++ {
		breaker.Call(ctx, func() error {
			return errors.New("fail")
		})
	}

	// Wait and transition to half-open
	time.Sleep(60 * time.Millisecond)

	// First call succeeds
	breaker.Call(ctx, func() error {
		return nil
	})

	// Second call fails - should reopen circuit
	breaker.Call(ctx, func() error {
		return errors.New("fail again")
	})

	assert.Equal(t, StateOpen, breaker.GetState())
}

func TestBreaker_Reset(t *testing.T) {
	config := DefaultConfig()
	config.MaxFailures = 2
	breaker := New(config)
	ctx := context.Background()

	// Open the circuit
	for i := 0; i < 2; i++ {
		breaker.Call(ctx, func() error {
			return errors.New("fail")
		})
	}

	assert.Equal(t, StateOpen, breaker.GetState())

	// Reset
	breaker.Reset()

	assert.Equal(t, StateClosed, breaker.GetState())

	stats := breaker.GetStats()
	assert.Equal(t, 0, stats.Failures)
}

func TestBreaker_StateChangeCallback(t *testing.T) {
	stateChanges := make([]string, 0)

	config := DefaultConfig()
	config.MaxFailures = 2
	config.OnStateChange = func(from, to State) {
		stateChanges = append(stateChanges,
			from.String()+" -> "+to.String())
	}

	breaker := New(config)
	ctx := context.Background()

	// Trigger state change
	for i := 0; i < 2; i++ {
		breaker.Call(ctx, func() error {
			return errors.New("fail")
		})
	}

	// Wait for callback
	time.Sleep(10 * time.Millisecond)

	require.Len(t, stateChanges, 1)
	assert.Equal(t, "closed -> open", stateChanges[0])
}

func TestBreaker_ConcurrentCalls(t *testing.T) {
	config := DefaultConfig()
	config.MaxFailures = 10
	breaker := New(config)
	ctx := context.Background()

	// Make concurrent calls
	concurrency := 100
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			breaker.Call(ctx, func() error {
				time.Sleep(1 * time.Millisecond)
				return nil
			})
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}

	stats := breaker.GetStats()
	assert.Equal(t, int64(concurrency), stats.TotalRequests)
	assert.Equal(t, int64(concurrency), stats.TotalSuccesses)
	assert.Equal(t, StateClosed, breaker.GetState())
}

func TestBreaker_Statistics(t *testing.T) {
	breaker := New(DefaultConfig())
	ctx := context.Background()

	// 5 successes
	for i := 0; i < 5; i++ {
		breaker.Call(ctx, func() error {
			return nil
		})
	}

	// 3 failures
	for i := 0; i < 3; i++ {
		breaker.Call(ctx, func() error {
			return errors.New("fail")
		})
	}

	stats := breaker.GetStats()

	assert.Equal(t, int64(8), stats.TotalRequests)
	assert.Equal(t, int64(5), stats.TotalSuccesses)
	assert.Equal(t, int64(3), stats.TotalFailures)
	assert.Equal(t, StateOpen, stats.State)
}

func TestCircuitOpenError(t *testing.T) {
	err := &CircuitOpenError{
		State:           StateOpen,
		Failures:        3,
		LastFailureTime: time.Now(),
	}

	assert.Contains(t, err.Error(), "circuit breaker is open")
	assert.True(t, IsCircuitOpen(err))
	assert.False(t, IsCircuitOpen(errors.New("other error")))
}

func BenchmarkBreaker_SuccessfulCall(b *testing.B) {
	breaker := New(DefaultConfig())
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		breaker.Call(ctx, func() error {
			return nil
		})
	}
}

func BenchmarkBreaker_FailedCall(b *testing.B) {
	breaker := New(DefaultConfig())
	ctx := context.Background()
	testErr := errors.New("test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		breaker.Call(ctx, func() error {
			return testErr
		})
	}
}
