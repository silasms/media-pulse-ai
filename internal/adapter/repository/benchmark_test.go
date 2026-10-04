package repository_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/silasms/media-pulse-ai/internal/adapter/repository"
	"github.com/silasms/media-pulse-ai/internal/domain"
)

func BenchmarkCachedAssetRepository_CacheHit(b *testing.B) {
	ctx := context.Background()
	inner := repository.NewMemoryAssetRepository()
	cache := repository.NewMemoryCache()
	cachedRepo := repository.NewCachedAssetRepository(inner, cache, 10*time.Minute)

	asset := &domain.MediaAsset{
		ID:              "ast_bench",
		Title:           "Tech Conference - Benchmark Test",
		Description:     "Aferição de latência de leitura com cache aside",
		DurationSeconds: 2700,
		MediaURL:        "https://cdn.example.com/vod/tech-conf.mp4",
		Status:          domain.StatusCompleted,
	}
	_ = cachedRepo.Save(ctx, asset)

	_, _ = cachedRepo.FindByID(ctx, "ast_bench")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cachedRepo.FindByID(ctx, "ast_bench")
	}
}

func BenchmarkJSONSerialization(b *testing.B) {
	asset := &domain.MediaAsset{
		ID:              "ast_bench_serial",
		Title:           "Série Original - Episódio 1",
		Description:     "Episódio de estreia com metadados enriquecidos",
		DurationSeconds: 3600,
		MediaURL:        "https://cdn.example.com/vod/serie-ep1.mp4",
		Status:          domain.StatusCompleted,
		Metadata: &domain.EnrichedMetadata{
			Summary:        "Abertura empolgante com narrativa dinâmica.",
			ParentalRating: domain.Rating14,
			KeyThemes:      []string{"Drama", "Suspense", "Brasil"},
			ModelUsed:      "mock-llm-v1",
			ProcessedAt:    time.Now(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := json.Marshal(asset)
		var out domain.MediaAsset
		_ = json.Unmarshal(data, &out)
	}
}
