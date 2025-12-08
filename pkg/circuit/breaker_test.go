package circuit

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBreaker_SuccessfulCalls(t *testing.T) {
	breaker := NewBreaker(3, 1*time.Second)

	// Successful calls should keep circuit closed
	for i := 0; i < 10; i++ {
		err := breaker.Call(func() error {
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, StateClosed, breaker.State())
		assert.Equal(t, 0, breaker.Failures())
	}
}

func TestBreaker_OpensAfterThreshold(t *testing.T) {
	breaker := NewBreaker(3, 1*time.Second)
	testErr := errors.New("test error")

	// First 2 failures should keep circuit closed
	for i := 0; i < 2; i++ {
		err := breaker.Call(func() error {
			return testErr
		})
		assert.Error(t, err)
		assert.Equal(t, StateClosed, breaker.State())
	}

	// Third failure should open circuit
	err := breaker.Call(func() error {
		return testErr
	})
	assert.Error(t, err)
	assert.Equal(t, StateOpen, breaker.State())
	assert.Equal(t, 3, breaker.Failures())

	// Subsequent calls should be blocked
	err = breaker.Call(func() error {
		return nil
	})
	assert.Equal(t, ErrCircuitOpen, err)
}

func TestBreaker_HalfOpenAndRecovery(t *testing.T) {
	breaker := NewBreaker(2, 100*time.Millisecond)
	testErr := errors.New("test error")

	// Open the circuit
	breaker.Call(func() error { return testErr })
	breaker.Call(func() error { return testErr })
	assert.Equal(t, StateOpen, breaker.State())

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Next call should enter half-open
	err := breaker.Call(func() error {
		return nil // Success
	})
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, breaker.State())
	assert.Equal(t, 0, breaker.Failures())
}

func TestBreaker_Reset(t *testing.T) {
	breaker := NewBreaker(2, 1*time.Second)
	testErr := errors.New("test error")

	// Open the circuit
	breaker.Call(func() error { return testErr })
	breaker.Call(func() error { return testErr })
	assert.Equal(t, StateOpen, breaker.State())

	// Manual reset
	breaker.Reset()
	assert.Equal(t, StateClosed, breaker.State())
	assert.Equal(t, 0, breaker.Failures())
}
