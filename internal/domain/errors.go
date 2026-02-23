package domain

import "errors"

var (
	ErrAssetNotFound      = errors.New("media asset not found")
	ErrInvalidTitle       = errors.New("title cannot be empty")
	ErrInvalidMediaURL    = errors.New("media url cannot be empty")
	ErrInvalidDuration    = errors.New("duration must be greater than zero")
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open: downstream AI service unavailable")
	ErrQueueFull          = errors.New("processing job queue is full: backpressure applied")
	ErrJobTimeout         = errors.New("enrichment job execution timed out")
	ErrEmptyInput         = errors.New("input payload cannot be empty")
)
