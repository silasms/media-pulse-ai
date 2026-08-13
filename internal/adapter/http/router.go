package http

import (
	"log/slog"
	"net/http"

	"github.com/silasms/media-pulse-ai/internal/adapter/http/middleware"
)

type RouterConfig struct {
	AssetHandler *AssetHandler
	HealthHandler *HealthHandler
	Metrics       *middleware.MetricsCollector
	RateLimiter   *middleware.RateLimiter
	Logger        *slog.Logger
}

func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/live", cfg.HealthHandler.Live)
	mux.HandleFunc("GET /health/ready", cfg.HealthHandler.Ready)
	mux.HandleFunc("GET /metrics", cfg.Metrics.Handler())

	mux.HandleFunc("POST /api/v1/assets", cfg.AssetHandler.Ingest)
	mux.HandleFunc("GET /api/v1/assets/{id}", cfg.AssetHandler.Get)
	mux.HandleFunc("GET /api/v1/assets", cfg.AssetHandler.List)
	mux.HandleFunc("POST /api/v1/assets/{id}/enrich", cfg.AssetHandler.TriggerEnrichment)

	var handler http.Handler = mux

	if cfg.RateLimiter != nil {
		handler = cfg.RateLimiter.Middleware()(handler)
	}
	if cfg.Metrics != nil {
		handler = cfg.Metrics.Middleware()(handler)
	}
	if cfg.Logger != nil {
		handler = middleware.RequestLogger(cfg.Logger)(handler)
	}

	return handler
}
