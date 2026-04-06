package ai

import (
	"context"

	"github.com/silasms/media-pulse-ai/internal/domain"
)

type EnrichmentInput struct {
	AssetID         string
	Title           string
	Description     string
	DurationSeconds int
	Transcript      string
	Language        string
}

type Provider interface {
	Enrich(ctx context.Context, input EnrichmentInput) (*domain.EnrichedMetadata, error)
	GenerateEmbeddings(ctx context.Context, text string) ([]float32, error)
	Name() string
}
