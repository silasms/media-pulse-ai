package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/silasms/media-pulse-ai/internal/domain"
	"github.com/silasms/media-pulse-ai/internal/usecase"
)

type AssetHandler struct {
	service *usecase.AssetService
}

func NewAssetHandler(service *usecase.AssetService) *AssetHandler {
	return &AssetHandler{service: service}
}

func (h *AssetHandler) Ingest(w http.ResponseWriter, r *http.Request) {
	var input usecase.IngestAssetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	asset, err := h.service.IngestAsset(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTitle) ||
			errors.Is(err, domain.ErrInvalidMediaURL) ||
			errors.Is(err, domain.ErrInvalidDuration) {
			writeJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(asset)
}

func (h *AssetHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, "missing asset id", http.StatusBadRequest)
		return
	}

	asset, err := h.service.GetAsset(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrAssetNotFound) {
			writeJSONError(w, "asset not found", http.StatusNotFound)
			return
		}
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(asset)
}

func (h *AssetHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	status := domain.AssetStatus(q.Get("status"))

	if limit <= 0 {
		limit = 20
	}

	assets, total, err := h.service.ListAssets(r.Context(), domain.ListFilter{
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data":   assets,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *AssetHandler) TriggerEnrichment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, "missing asset id", http.StatusBadRequest)
		return
	}

	if err := h.service.TriggerEnrichment(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrAssetNotFound) {
			writeJSONError(w, "asset not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrQueueFull) {
			writeJSONError(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "enrichment job dispatched",
		"asset_id": id,
	})
}

func writeJSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
