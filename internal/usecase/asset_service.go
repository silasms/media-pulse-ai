package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/silasms/media-pulse-ai/internal/domain"
)

type IngestAssetInput struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	DurationSeconds int    `json:"duration_seconds"`
	Language        string `json:"language"`
	MediaURL        string `json:"media_url"`
	Transcript      string `json:"transcript,omitempty"`
}

type QueueDispatcher interface {
	Dispatch(job *domain.EnrichmentJob) bool
}

type AssetService struct {
	repo       domain.AssetRepository
	dispatcher QueueDispatcher
}

func NewAssetService(repo domain.AssetRepository, dispatcher QueueDispatcher) *AssetService {
	return &AssetService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

func (s *AssetService) IngestAsset(ctx context.Context, input IngestAssetInput) (*domain.MediaAsset, error) {
	id, err := generateID("ast")
	if err != nil {
		return nil, fmt.Errorf("failed to generate asset id: %w", err)
	}

	asset := &domain.MediaAsset{
		ID:              id,
		Title:           input.Title,
		Description:     input.Description,
		DurationSeconds: input.DurationSeconds,
		Language:        input.Language,
		MediaURL:        input.MediaURL,
		Transcript:      input.Transcript,
		Status:          domain.StatusPending,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	if err := asset.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, asset); err != nil {
		return nil, fmt.Errorf("failed to save asset: %w", err)
	}

	if s.dispatcher != nil {
		jobID, _ := generateID("job")
		job := &domain.EnrichmentJob{
			ID:          jobID,
			AssetID:     asset.ID,
			Priority:    1,
			Status:      domain.JobPending,
			MaxAttempts: 3,
			CreatedAt:   time.Now().UTC(),
		}
		if !s.dispatcher.Dispatch(job) {
			_ = s.repo.UpdateStatus(ctx, asset.ID, domain.StatusPending, "queued_for_retry")
		}
	}

	return asset, nil
}

func (s *AssetService) GetAsset(ctx context.Context, id string) (*domain.MediaAsset, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *AssetService) ListAssets(ctx context.Context, filter domain.ListFilter) ([]*domain.MediaAsset, int, error) {
	return s.repo.List(ctx, filter)
}

func (s *AssetService) TriggerEnrichment(ctx context.Context, assetID string) error {
	asset, err := s.repo.FindByID(ctx, assetID)
	if err != nil {
		return err
	}

	if s.dispatcher == nil {
		return nil
	}

	jobID, _ := generateID("job")
	job := &domain.EnrichmentJob{
		ID:          jobID,
		AssetID:     asset.ID,
		Priority:    2,
		Status:      domain.JobPending,
		MaxAttempts: 3,
		CreatedAt:   time.Now().UTC(),
	}

	if !s.dispatcher.Dispatch(job) {
		return domain.ErrQueueFull
	}

	return s.repo.UpdateStatus(ctx, asset.ID, domain.StatusPending, "")
}

func generateID(prefix string) (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(bytes)), nil
}
