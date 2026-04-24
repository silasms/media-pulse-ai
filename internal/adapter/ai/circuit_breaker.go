package ai

import (
	"context"
	"sync"
	"time"

	"github.com/silasms/media-pulse-ai/internal/domain"
)

type State string

const (
	StateClosed   State = "CLOSED"
	StateOpen     State = "OPEN"
	StateHalfOpen State = "HALF_OPEN"
)

type CircuitBreakerConfig struct {
	FailureThreshold int
	SuccessThreshold int
	Timeout          time.Duration
}

type CircuitBreaker struct {
	mu               sync.Mutex
	config           CircuitBreakerConfig
	state            State
	failureCount     int
	successCount     int
	lastStateChanged time.Time
}

func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 3
	}
	if cfg.SuccessThreshold <= 0 {
		cfg.SuccessThreshold = 2
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}

	return &CircuitBreaker{
		config:           cfg,
		state:            StateClosed,
		lastStateChanged: time.Now(),
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.evaluateStateLocked()
	return cb.state
}

func (cb *CircuitBreaker) evaluateStateLocked() {
	if cb.state == StateOpen && time.Since(cb.lastStateChanged) > cb.config.Timeout {
		cb.state = StateHalfOpen
		cb.failureCount = 0
		cb.successCount = 0
		cb.lastStateChanged = time.Now()
	}
}

func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	cb.mu.Lock()
	cb.evaluateStateLocked()

	if cb.state == StateOpen {
		cb.mu.Unlock()
		return domain.ErrCircuitBreakerOpen
	}
	cb.mu.Unlock()

	err := fn(ctx)

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailureLocked()
		return err
	}

	cb.onSuccessLocked()
	return nil
}

func (cb *CircuitBreaker) onFailureLocked() {
	cb.failureCount++
	if cb.state == StateClosed && cb.failureCount >= cb.config.FailureThreshold {
		cb.state = StateOpen
		cb.lastStateChanged = time.Now()
	} else if cb.state == StateHalfOpen {
		cb.state = StateOpen
		cb.lastStateChanged = time.Now()
	}
}

func (cb *CircuitBreaker) onSuccessLocked() {
	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.config.SuccessThreshold {
			cb.state = StateClosed
			cb.failureCount = 0
			cb.successCount = 0
			cb.lastStateChanged = time.Now()
		}
	} else if cb.state == StateClosed {
		cb.failureCount = 0
	}
}
