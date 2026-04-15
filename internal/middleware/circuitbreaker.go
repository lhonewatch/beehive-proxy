package middleware

import (
	"net/http"
	"sync"
	"time"
)

type circuitState int

const (
	stateClosed circuitState = iota
	stateOpen
	stateHalfOpen
)

// CircuitBreaker tracks failures per upstream and opens the circuit
// after a configurable threshold, rejecting requests until a cooldown
// period has elapsed.
type CircuitBreaker struct {
	mu           sync.Mutex
	failures     int
	threshold    int
	cooldown     time.Duration
	openedAt     time.Time
	state        circuitState
}

// NewCircuitBreaker returns a CircuitBreaker that opens after threshold
// consecutive failures and resets after cooldown.
func NewCircuitBreaker(threshold int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
		cooldown:  cooldown,
		state:     stateClosed,
	}
}

// Allow reports whether the circuit is closed (or half-open for a probe).
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case stateOpen:
		if time.Since(cb.openedAt) >= cb.cooldown {
			cb.state = stateHalfOpen
			return true
		}
		return false
	default:
		return true
	}
}

// RecordSuccess resets the breaker to closed.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = stateClosed
}

// RecordFailure increments the failure counter and opens the circuit if
// the threshold is reached.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.failures >= cb.threshold {
		cb.state = stateOpen
		cb.openedAt = time.Now()
	}
}

// Middleware wraps an http.Handler with circuit-breaker logic.
func (cb *CircuitBreaker) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cb.Allow() {
			http.Error(w, "service unavailable — circuit open", http.StatusServiceUnavailable)
			return
		}
		rec := NewResponseRecorder(w)
		next.ServeHTTP(rec, r)
		if rec.Status >= 500 {
			cb.RecordFailure()
		} else {
			cb.RecordSuccess()
		}
	})
}
