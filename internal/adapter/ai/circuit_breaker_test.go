package ai_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/silasms/media-pulse-ai/internal/adapter/ai"
	"github.com/silasms/media-pulse-ai/internal/domain"
)

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	ctx := context.Background()
	cb := ai.NewCircuitBreaker(ai.CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          50 * time.Millisecond,
	})

	if cb.State() != ai.StateClosed {
		t.Fatalf("expected initial state CLOSED, got %s", cb.State())
	}

	errDummy := errors.New("ai provider 503")

	_ = cb.Execute(ctx, func(ctx context.Context) error { return errDummy })
	if cb.State() != ai.StateClosed {
		t.Fatalf("expected state CLOSED after 1 failure, got %s", cb.State())
	}

	_ = cb.Execute(ctx, func(ctx context.Context) error { return errDummy })
	if cb.State() != ai.StateOpen {
		t.Fatalf("expected state OPEN after 2 failures, got %s", cb.State())
	}

	err := cb.Execute(ctx, func(ctx context.Context) error { return nil })
	if err != domain.ErrCircuitBreakerOpen {
		t.Fatalf("expected ErrCircuitBreakerOpen, got %v", err)
	}

	time.Sleep(60 * time.Millisecond)
	if cb.State() != ai.StateHalfOpen {
		t.Fatalf("expected state HALF_OPEN after timeout, got %s", cb.State())
	}

	_ = cb.Execute(ctx, func(ctx context.Context) error { return nil })
	_ = cb.Execute(ctx, func(ctx context.Context) error { return nil })

	if cb.State() != ai.StateClosed {
		t.Fatalf("expected state CLOSED after recovery, got %s", cb.State())
	}
}
