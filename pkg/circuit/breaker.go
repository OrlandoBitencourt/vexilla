package circuit

import (
	"fmt"
	"sync"
	"time"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed   State = iota // Normal operation
	StateOpen                  // Circuit open, blocking requests
	StateHalfOpen              // Testing if service recovered
)

// Breaker implements a circuit breaker pattern
type Breaker struct {
	mu              sync.RWMutex
	state           State
	failures        int
	threshold       int           // Failures needed to open
	timeout         time.Duration // Time before attempting recovery
	lastFailureTime time.Time
	lastSuccessTime time.Time
}

// NewBreaker creates a new circuit breaker
func NewBreaker(threshold int, timeout time.Duration) *Breaker {
	return &Breaker{
		state:     StateClosed,
		threshold: threshold,
		timeout:   timeout,
	}
}

// Call executes the function if circuit allows
func (b *Breaker) Call(fn func() error) error {
	if !b.Allow() {
		return ErrCircuitOpen
	}

	err := fn()
	if err != nil {
		b.RecordFailure()
		return err
	}

	b.RecordSuccess()
	return nil
}

// Allow checks if the circuit allows requests
func (b *Breaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if timeout elapsed
		if time.Since(b.lastFailureTime) > b.timeout {
			b.state = StateHalfOpen
			return true
		}
		return false
	case StateHalfOpen:
		return true
	}

	return false
}

// RecordSuccess records a successful call
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.failures = 0
	b.lastSuccessTime = time.Now()

	if b.state == StateHalfOpen {
		b.state = StateClosed
	}
}

// RecordFailure records a failed call
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.failures++
	b.lastFailureTime = time.Now()

	if b.failures >= b.threshold {
		b.state = StateOpen
	}
}

// Reset manually resets the circuit breaker
func (b *Breaker) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.state = StateClosed
	b.failures = 0
}

// State returns the current state
func (b *Breaker) State() State {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.state
}

// Failures returns the current failure count
func (b *Breaker) Failures() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.failures
}

// ErrCircuitOpen is returned when circuit is open
var ErrCircuitOpen = fmt.Errorf("circuit breaker is open")
