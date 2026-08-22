package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	adapterHTTP "github.com/silasms/media-pulse-ai/internal/adapter/http"
	"github.com/silasms/media-pulse-ai/internal/adapter/http/middleware"
	"github.com/silasms/media-pulse-ai/internal/adapter/repository"
	"github.com/silasms/media-pulse-ai/internal/domain"
	"github.com/silasms/media-pulse-ai/internal/usecase"
)

func setupTestServer() http.Handler {
	repo := repository.NewMemoryAssetRepository()
	svc := usecase.NewAssetService(repo, nil)
	assetHandler := adapterHTTP.NewAssetHandler(svc)
	healthHandler := adapterHTTP.NewHealthHandler(func() bool { return true })
	metrics := middleware.NewMetricsCollector()

	return adapterHTTP.NewRouter(adapterHTTP.RouterConfig{
		AssetHandler:  assetHandler,
		HealthHandler: healthHandler,
		Metrics:       metrics,
		RateLimiter:   nil,
		Logger:        nil,
	})
}

func TestAssetHandler_IngestAndGet(t *testing.T) {
	router := setupTestServer()

	payload := `{"title":"Tech Weekly Review","description":"Análise política e econômica","duration_seconds":2400,"media_url":"https://cdn.example.com/vod/tech-weekly.mp4","language":"pt-BR"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/assets", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	var created domain.MediaAsset
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if created.ID == "" || created.Title != "Tech Weekly Review" {
		t.Fatalf("unexpected asset payload: %+v", created)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/assets/"+created.ID, nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", getW.Code)
	}

	var fetched domain.MediaAsset
	_ = json.Unmarshal(getW.Body.Bytes(), &fetched)
	if fetched.ID != created.ID {
		t.Fatalf("expected ID %s, got %s", created.ID, fetched.ID)
	}
}

func TestHealthEndpoints(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for /health/live, got %d", w.Code)
	}
}
