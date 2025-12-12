package circuit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// State represents the circuit breaker state
type State int

const (
	// StateClosed - circuit is closed, requests pass through
	StateClosed State = iota
	// StateOpen - circuit is open, requests fail fast
	StateOpen
	// StateHalfOpen - circuit is testing if service recovered
	StateHalfOpen
)

// String returns string representation of state
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Breaker implements the circuit breaker pattern
type Breaker struct {
	mu sync.RWMutex

	// Configuration
	maxFailures     int
	timeout         time.Duration
	halfOpenTimeout time.Duration

	// State
	state           State
	failures        int
	successes       int
	lastFailureTime time.Time
	lastStateChange time.Time

	// Statistics
	totalRequests   int64
	totalSuccesses  int64
	totalFailures   int64
	totalRejections int64

	// Callbacks
	onStateChange func(from, to State)
}

// Config holds circuit breaker configuration
type Config struct {
	// MaxFailures is the number of consecutive failures before opening
	MaxFailures int

	// Timeout is how long to wait before attempting recovery (half-open)
	Timeout time.Duration

	// HalfOpenTimeout is how long to stay in half-open before closing
	HalfOpenTimeout time.Duration

	// OnStateChange is called when state changes
	OnStateChange func(from, to State)
}

// DefaultConfig returns default circuit breaker configuration
func DefaultConfig() Config {
	return Config{
		MaxFailures:     3,
		Timeout:         30 * time.Second,
		HalfOpenTimeout: 10 * time.Second,
		OnStateChange:   nil,
	}
}

// New creates a new circuit breaker
func New(config Config) *Breaker {
	if config.MaxFailures <= 0 {
		config.MaxFailures = 3
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}
	if config.HalfOpenTimeout <= 0 {
		config.HalfOpenTimeout = 10 * time.Second
	}

	return &Breaker{
		maxFailures:     config.MaxFailures,
		timeout:         config.Timeout,
		halfOpenTimeout: config.HalfOpenTimeout,
		state:           StateClosed,
		lastStateChange: time.Now(),
		onStateChange:   config.OnStateChange,
	}
}

// Call executes the function with circuit breaker protection
func (b *Breaker) Call(ctx context.Context, fn func() error) error {
	// Check if circuit allows the call
	if err := b.beforeCall(); err != nil {
		return err
	}

	// Execute the function
	err := fn()

	// Record the result
	b.afterCall(err)

	return err
}

// beforeCall checks if the circuit breaker allows the call
func (b *Breaker) beforeCall() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.totalRequests++

	switch b.state {
	case StateClosed:
		// Closed state - allow request
		return nil

	case StateOpen:
		// Check if timeout has elapsed
		if time.Since(b.lastStateChange) >= b.timeout {
			// Transition to half-open
			b.setState(StateHalfOpen)
			return nil
		}

		// Still open - reject request
		b.totalRejections++
		return &CircuitOpenError{
			State:           b.state,
			Failures:        b.failures,
			LastFailureTime: b.lastFailureTime,
		}

	case StateHalfOpen:
		// Half-open state - allow limited requests
		return nil

	default:
		return fmt.Errorf("unknown circuit breaker state: %d", b.state)
	}
}

// afterCall records the result of the call
func (b *Breaker) afterCall(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err != nil {
		b.onFailure()
	} else {
		b.onSuccess()
	}
}

// onSuccess handles successful calls
func (b *Breaker) onSuccess() {
	b.totalSuccesses++
	b.failures = 0

	switch b.state {
	case StateClosed:
		// Remain closed

	case StateHalfOpen:
		b.successes++

		// If we've had enough successes in half-open, close the circuit
		if b.successes >= 2 {
			b.setState(StateClosed)
			b.successes = 0
		}

	case StateOpen:
		// This shouldn't happen, but reset if it does
		b.setState(StateClosed)
	}
}

// onFailure handles failed calls
func (b *Breaker) onFailure() {
	b.totalFailures++
	b.failures++
	b.lastFailureTime = time.Now()

	switch b.state {
	case StateClosed:
		// Check if we've exceeded max failures
		if b.failures >= b.maxFailures {
			b.setState(StateOpen)
		}

	case StateHalfOpen:
		// Any failure in half-open immediately opens the circuit
		b.setState(StateOpen)
		b.successes = 0

	case StateOpen:
		// Already open, update failure time
		b.lastStateChange = time.Now()
	}
}

// setState transitions to a new state
func (b *Breaker) setState(newState State) {
	oldState := b.state

	if oldState == newState {
		return
	}

	b.state = newState
	b.lastStateChange = time.Now()

	// Call state change callback
	if b.onStateChange != nil {
		go b.onStateChange(oldState, newState)
	}
}

// GetState returns the current state
func (b *Breaker) GetState() State {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.state
}

// Reset resets the circuit breaker to closed state
func (b *Breaker) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.setState(StateClosed)
	b.failures = 0
	b.successes = 0
}

// GetStats returns circuit breaker statistics
func (b *Breaker) GetStats() Stats {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return Stats{
		State:           b.state,
		Failures:        b.failures,
		Successes:       b.successes,
		TotalRequests:   b.totalRequests,
		TotalSuccesses:  b.totalSuccesses,
		TotalFailures:   b.totalFailures,
		TotalRejections: b.totalRejections,
		LastFailureTime: b.lastFailureTime,
		LastStateChange: b.lastStateChange,
	}
}

// Stats represents circuit breaker statistics
type Stats struct {
	State           State
	Failures        int
	Successes       int
	TotalRequests   int64
	TotalSuccesses  int64
	TotalFailures   int64
	TotalRejections int64
	LastFailureTime time.Time
	LastStateChange time.Time
}

// CircuitOpenError is returned when the circuit is open
type CircuitOpenError struct {
	State           State
	Failures        int
	LastFailureTime time.Time
}

// Error implements the error interface
func (e *CircuitOpenError) Error() string {
	return fmt.Sprintf("circuit breaker is %s (failures: %d, last failure: %s)",
		e.State.String(), e.Failures, e.LastFailureTime.Format(time.RFC3339))
}

// IsCircuitOpen checks if error is a circuit open error
func IsCircuitOpen(err error) bool {
	_, ok := err.(*CircuitOpenError)
	return ok
}
