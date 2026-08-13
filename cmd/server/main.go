package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/silasms/media-pulse-ai/internal/adapter/ai"
	adapterHTTP "github.com/silasms/media-pulse-ai/internal/adapter/http"
	"github.com/silasms/media-pulse-ai/internal/adapter/http/middleware"
	"github.com/silasms/media-pulse-ai/internal/adapter/repository"
	"github.com/silasms/media-pulse-ai/internal/config"
	"github.com/silasms/media-pulse-ai/internal/domain"
	"github.com/silasms/media-pulse-ai/internal/pkg/workerpool"
	"github.com/silasms/media-pulse-ai/internal/usecase"
)

type poolDispatcher struct {
	pool *workerpool.Pool[*domain.EnrichmentJob]
}

func (d *poolDispatcher) Dispatch(job *domain.EnrichmentJob) bool {
	return d.pool.Submit(job)
}

func main() {
	cfg := config.Load()

	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	logger.Info("Starting MediaPulse AI Service",
		slog.String("port", cfg.Port),
		slog.Int("workers", cfg.WorkerPoolSize),
		slog.String("ai_provider", cfg.AIProvider),
		slog.Bool("redis_enabled", cfg.RedisEnabled),
	)

	memRepo := repository.NewMemoryAssetRepository()
	var cacheClient repository.CacheClient

	if cfg.RedisEnabled {
		logger.Info("Connecting to Redis cache", slog.String("addr", cfg.RedisAddr))
		rdb := repository.NewRedisCacheClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
		ctxPing, cancelPing := context.WithTimeout(context.Background(), 2*time.Second)
		if err := rdb.Ping(ctxPing); err != nil {
			logger.Warn("Redis ping failed, falling back to in-memory cache", slog.String("error", err.Error()))
			cacheClient = repository.NewMemoryCache()
		} else {
			cacheClient = rdb
		}
		cancelPing()
	} else {
		cacheClient = repository.NewMemoryCache()
	}

	cachedRepo := repository.NewCachedAssetRepository(memRepo, cacheClient, cfg.CacheTTL)

	aiProvider := ai.NewMockProvider()
	cb := ai.NewCircuitBreaker(ai.CircuitBreakerConfig{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		Timeout:          10 * time.Second,
	})

	workerHandler := usecase.NewEnrichmentWorker(cachedRepo, aiProvider, cb, cachedRepo, logger)
	pool := workerpool.New[*domain.EnrichmentJob](
		cfg.WorkerPoolSize,
		cfg.QueueCapacity,
		func(ctx context.Context, job *domain.EnrichmentJob) {
			workerHandler.Process(ctx, job)
		},
	)
	pool.Start()
	defer pool.Stop()

	dispatcher := &poolDispatcher{pool: pool}
	assetService := usecase.NewAssetService(cachedRepo, dispatcher)

	assetHandler := adapterHTTP.NewAssetHandler(assetService)
	healthHandler := adapterHTTP.NewHealthHandler(func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		return cacheClient.Ping(ctx) == nil
	})
	metrics := middleware.NewMetricsCollector()
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPS)

	router := adapterHTTP.NewRouter(adapterHTTP.RouterConfig{
		AssetHandler:  assetHandler,
		HealthHandler: healthHandler,
		Metrics:       metrics,
		RateLimiter:   rateLimiter,
		Logger:        logger,
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("HTTP server listening", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failure", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	<-shutdownChan
	logger.Info("Graceful shutdown signal received")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Graceful shutdown error", slog.String("error", err.Error()))
	}

	logger.Info("MediaPulse AI terminated gracefully")
}
