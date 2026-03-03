package domain

import (
	"context"
	"time"
)

type JobStatus string

const (
	JobPending   JobStatus = "PENDING"
	JobRunning   JobStatus = "RUNNING"
	JobCompleted JobStatus = "COMPLETED"
	JobFailed    JobStatus = "FAILED"
)

type EnrichmentJob struct {
	ID          string     `json:"id"`
	AssetID     string     `json:"asset_id"`
	Priority    int        `json:"priority"`
	Status      JobStatus  `json:"status"`
	Attempts    int        `json:"attempts"`
	MaxAttempts int        `json:"max_attempts"`
	LastError   string     `json:"last_error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type JobQueue interface {
	Enqueue(ctx context.Context, job *EnrichmentJob) error
	Jobs() <-chan *EnrichmentJob
}
