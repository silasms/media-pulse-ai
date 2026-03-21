package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/silasms/media-pulse-ai/internal/adapter/repository"
	"github.com/silasms/media-pulse-ai/internal/domain"
)

func TestMemoryAssetRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryAssetRepository()

	asset := &domain.MediaAsset{
		ID:              "asset-001",
		Title:           "Documentário - Vida Selvagem",
		Description:     "Especial sobre a biodiversidade brasileira",
		DurationSeconds: 3600,
		Language:        "pt-BR",
		MediaURL:        "https://cdn.example.com/vod/documentario.mp4",
		Status:          domain.StatusPending,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	if err := repo.Save(ctx, asset); err != nil {
		t.Fatalf("failed to save asset: %v", err)
	}

	found, err := repo.FindByID(ctx, "asset-001")
	if err != nil {
		t.Fatalf("failed to find asset: %v", err)
	}
	if found.Title != asset.Title {
		t.Fatalf("expected title %s, got %s", asset.Title, found.Title)
	}

	if err := repo.UpdateStatus(ctx, "asset-001", domain.StatusProcessing, ""); err != nil {
		t.Fatalf("failed to update status: %v", err)
	}
	updated, _ := repo.FindByID(ctx, "asset-001")
	if updated.Status != domain.StatusProcessing {
		t.Fatalf("expected status %s, got %s", domain.StatusProcessing, updated.Status)
	}

	meta := &domain.EnrichedMetadata{
		Summary:        "Documentário fascinante sobre preservação ambiental e biodiversidade.",
		ParentalRating: domain.RatingGeneral,
		KeyThemes:      []string{"Natureza", "Ciência", "Brasil"},
		ModelUsed:      "mock-llm-v1",
		ProcessedAt:    time.Now().UTC(),
	}
	if err := repo.SetMetadata(ctx, "asset-001", meta); err != nil {
		t.Fatalf("failed to set metadata: %v", err)
	}

	completed, _ := repo.FindByID(ctx, "asset-001")
	if completed.Status != domain.StatusCompleted {
		t.Fatalf("expected status %s, got %s", domain.StatusCompleted, completed.Status)
	}
	if completed.Metadata == nil || completed.Metadata.Summary != meta.Summary {
		t.Fatalf("metadata summary mismatch")
	}

	list, total, err := repo.List(ctx, domain.ListFilter{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("failed to list assets: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected 1 item, got total=%d len=%d", total, len(list))
	}
}

func TestMemoryAssetRepository_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryAssetRepository()

	_, err := repo.FindByID(ctx, "non-existent")
	if err != domain.ErrAssetNotFound {
		t.Fatalf("expected ErrAssetNotFound, got %v", err)
	}
}
