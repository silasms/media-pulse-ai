package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/silasms/media-pulse-ai/internal/adapter/ai"
	"github.com/silasms/media-pulse-ai/internal/adapter/repository"
	"github.com/silasms/media-pulse-ai/internal/domain"
	"github.com/silasms/media-pulse-ai/internal/usecase"
)

func TestEnrichmentWorker_SuccessfulProcess(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryAssetRepository()
	provider := ai.NewMockProvider()
	cb := ai.NewCircuitBreaker(ai.CircuitBreakerConfig{})
	worker := usecase.NewEnrichmentWorker(repo, provider, cb, nil, nil)

	asset := &domain.MediaAsset{
		ID:              "ast_test_100",
		Title:           "Special Feature - Inteligência Artificial",
		Description:     "Reportagens especiais e investigação",
		DurationSeconds: 7200,
		Language:        "pt-BR",
		MediaURL:        "https://cdn.example.com/vod/special-feature.mp4",
		Status:          domain.StatusPending,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	_ = repo.Save(ctx, asset)

	job := &domain.EnrichmentJob{
		ID:      "job_test_1",
		AssetID: asset.ID,
	}

	worker.Process(ctx, job)

	updated, err := repo.FindByID(ctx, asset.ID)
	if err != nil {
		t.Fatalf("failed to find asset: %v", err)
	}
	if updated.Status != domain.StatusCompleted {
		t.Fatalf("expected status COMPLETED, got %s", updated.Status)
	}
	if updated.Metadata == nil {
		t.Fatalf("expected metadata to be populated")
	}
	if len(updated.Metadata.Chapters) == 0 {
		t.Fatalf("expected chapters to be generated")
	}
	if len(updated.Metadata.Embeddings) != 384 {
		t.Fatalf("expected 384-dim embeddings, got %d", len(updated.Metadata.Embeddings))
	}
}

func TestEnrichmentWorker_FailureRecovery(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryAssetRepository()
	provider := ai.NewMockProvider()
	provider.SetFailNext(true)

	cb := ai.NewCircuitBreaker(ai.CircuitBreakerConfig{})
	worker := usecase.NewEnrichmentWorker(repo, provider, cb, nil, nil)

	asset := &domain.MediaAsset{
		ID:              "ast_test_fail",
		Title:           "Programa Teste",
		DurationSeconds: 300,
		Language:        "pt-BR",
		MediaURL:        "https://cdn.example.com/vod/teste.mp4",
		Status:          domain.StatusPending,
	}
	_ = repo.Save(ctx, asset)

	job := &domain.EnrichmentJob{
		ID:      "job_test_fail",
		AssetID: asset.ID,
	}

	worker.Process(ctx, job)

	updated, _ := repo.FindByID(ctx, asset.ID)
	if updated.Status != domain.StatusFailed {
		t.Fatalf("expected status FAILED after provider error, got %s", updated.Status)
	}
}
