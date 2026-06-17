package usecase_test

import (
	"context"
	"testing"

	"github.com/silasms/media-pulse-ai/internal/adapter/repository"
	"github.com/silasms/media-pulse-ai/internal/domain"
	"github.com/silasms/media-pulse-ai/internal/usecase"
)

type mockDispatcher struct {
	dispatched []*domain.EnrichmentJob
}

func (m *mockDispatcher) Dispatch(job *domain.EnrichmentJob) bool {
	m.dispatched = append(m.dispatched, job)
	return true
}

func TestAssetService_IngestAndRetrieve(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryAssetRepository()
	disp := &mockDispatcher{}
	svc := usecase.NewAssetService(repo, disp)

	input := usecase.IngestAssetInput{
		Title:           "Sports Highlights - Melhores Momentos",
		Description:     "Gols da rodada do Campeonato Brasileiro",
		DurationSeconds: 1200,
		MediaURL:        "https://cdn.example.com/vod/highlights.mp4",
		Language:        "pt-BR",
	}

	asset, err := svc.IngestAsset(ctx, input)
	if err != nil {
		t.Fatalf("unexpected ingest error: %v", err)
	}
	if asset.ID == "" {
		t.Fatalf("expected non-empty asset ID")
	}
	if asset.Status != domain.StatusPending {
		t.Fatalf("expected status PENDING, got %s", asset.Status)
	}
	if len(disp.dispatched) != 1 {
		t.Fatalf("expected 1 job dispatched, got %d", len(disp.dispatched))
	}

	retrieved, err := svc.GetAsset(ctx, asset.ID)
	if err != nil {
		t.Fatalf("failed to retrieve asset: %v", err)
	}
	if retrieved.Title != input.Title {
		t.Fatalf("expected title %s, got %s", input.Title, retrieved.Title)
	}
}

func TestAssetService_ValidationFailure(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryAssetRepository()
	svc := usecase.NewAssetService(repo, nil)

	input := usecase.IngestAssetInput{
		Title:           "",
		DurationSeconds: 100,
		MediaURL:        "https://valid.url",
	}

	_, err := svc.IngestAsset(ctx, input)
	if err != domain.ErrInvalidTitle {
		t.Fatalf("expected ErrInvalidTitle, got %v", err)
	}
}
