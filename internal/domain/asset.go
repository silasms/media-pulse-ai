package domain

import (
	"context"
	"time"
)

type AssetStatus string

const (
	StatusPending    AssetStatus = "PENDING"
	StatusProcessing AssetStatus = "PROCESSING"
	StatusCompleted  AssetStatus = "COMPLETED"
	StatusFailed     AssetStatus = "FAILED"
)

type ParentalRating string

const (
	RatingGeneral ParentalRating = "L"
	Rating10      ParentalRating = "10"
	Rating12      ParentalRating = "12"
	Rating14      ParentalRating = "14"
	Rating16      ParentalRating = "16"
	Rating18      ParentalRating = "18"
)

type Chapter struct {
	StartSecond int    `json:"start_second"`
	EndSecond   int    `json:"end_second"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
}

type EnrichedMetadata struct {
	Summary         string         `json:"summary"`
	KeyThemes       []string       `json:"key_themes"`
	Chapters        []Chapter      `json:"chapters"`
	ParentalRating  ParentalRating `json:"parental_rating"`
	ContentWarnings []string       `json:"content_warnings"`
	Tags            []string       `json:"tags"`
	Sentiment       string         `json:"sentiment"`
	Embeddings      []float32      `json:"embeddings,omitempty"`
	ModelUsed       string         `json:"model_used"`
	ProcessedAt     time.Time      `json:"processed_at"`
}

type MediaAsset struct {
	ID              string            `json:"id"`
	Title           string            `json:"title"`
	Description     string            `json:"description"`
	DurationSeconds int               `json:"duration_seconds"`
	Language        string            `json:"language"`
	MediaURL        string            `json:"media_url"`
	Transcript      string            `json:"transcript,omitempty"`
	Status          AssetStatus       `json:"status"`
	Metadata        *EnrichedMetadata `json:"metadata,omitempty"`
	ErrorMessage    string            `json:"error_message,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

func (a *MediaAsset) Validate() error {
	if a.Title == "" {
		return ErrInvalidTitle
	}
	if a.MediaURL == "" {
		return ErrInvalidMediaURL
	}
	if a.DurationSeconds <= 0 {
		return ErrInvalidDuration
	}
	if a.Language == "" {
		a.Language = "pt-BR"
	}
	return nil
}

type ListFilter struct {
	Status AssetStatus
	Limit  int
	Offset int
}

type AssetRepository interface {
	Save(ctx context.Context, asset *MediaAsset) error
	FindByID(ctx context.Context, id string) (*MediaAsset, error)
	List(ctx context.Context, filter ListFilter) ([]*MediaAsset, int, error)
	Update(ctx context.Context, asset *MediaAsset) error
	UpdateStatus(ctx context.Context, id string, status AssetStatus, errMsg string) error
	SetMetadata(ctx context.Context, id string, metadata *EnrichedMetadata) error
}
