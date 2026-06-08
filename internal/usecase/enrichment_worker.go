package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/silasms/media-pulse-ai/internal/adapter/ai"
	"github.com/silasms/media-pulse-ai/internal/domain"
)

type CacheInvalidator interface {
	Invalidate(ctx context.Context, assetID string) error
}

type EnrichmentWorker struct {
	repo           domain.AssetRepository
	provider       ai.Provider
	circuitBreaker *ai.CircuitBreaker
	cache          CacheInvalidator
	logger         *slog.Logger
}

func NewEnrichmentWorker(
	repo domain.AssetRepository,
	provider ai.Provider,
	circuitBreaker *ai.CircuitBreaker,
	cache CacheInvalidator,
	logger *slog.Logger,
) *EnrichmentWorker {
	if logger == nil {
		logger = slog.Default()
	}
	return &EnrichmentWorker{
		repo:           repo,
		provider:       provider,
		circuitBreaker: circuitBreaker,
		cache:          cache,
		logger:         logger,
	}
}

func (w *EnrichmentWorker) Process(ctx context.Context, job *domain.EnrichmentJob) {
	start := time.Now()
	w.logger.Info("Starting AI enrichment job",
		slog.String("job_id", job.ID),
		slog.String("asset_id", job.AssetID),
	)

	asset, err := w.repo.FindByID(ctx, job.AssetID)
	if err != nil {
		w.logger.Error("Failed to find asset for enrichment",
			slog.String("asset_id", job.AssetID),
			slog.String("error", err.Error()),
		)
		return
	}

	if err := w.repo.UpdateStatus(ctx, asset.ID, domain.StatusProcessing, ""); err != nil {
		w.logger.Error("Failed to update asset status to processing", slog.String("error", err.Error()))
	}

	var metadata *domain.EnrichedMetadata
	cbErr := w.circuitBreaker.Execute(ctx, func(c context.Context) error {
		input := ai.EnrichmentInput{
			AssetID:         asset.ID,
			Title:           asset.Title,
			Description:     asset.Description,
			DurationSeconds: asset.DurationSeconds,
			Transcript:      asset.Transcript,
			Language:        asset.Language,
		}

		res, err := w.provider.Enrich(c, input)
		if err != nil {
			return err
		}

		vec, err := w.provider.GenerateEmbeddings(c, asset.Title+" "+asset.Description)
		if err == nil {
			res.Embeddings = vec
		}

		metadata = res
		return nil
	})

	if cbErr != nil {
		w.logger.Warn("Enrichment failed",
			slog.String("asset_id", asset.ID),
			slog.String("error", cbErr.Error()),
		)
		_ = w.repo.UpdateStatus(ctx, asset.ID, domain.StatusFailed, cbErr.Error())
		return
	}

	if err := w.repo.SetMetadata(ctx, asset.ID, metadata); err != nil {
		w.logger.Error("Failed to persist enriched metadata", slog.String("error", err.Error()))
		return
	}

	if w.cache != nil {
		_ = w.cache.Invalidate(ctx, asset.ID)
	}

	w.logger.Info("Completed AI enrichment job successfully",
		slog.String("asset_id", asset.ID),
		slog.Duration("latency", time.Since(start)),
		slog.String("rating", string(metadata.ParentalRating)),
		slog.Int("chapters_count", len(metadata.Chapters)),
	)
}
