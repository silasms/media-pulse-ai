package repository

import (
	"context"
	"sync"
	"time"

	"github.com/silasms/media-pulse-ai/internal/domain"
)

type MemoryAssetRepository struct {
	mu     sync.RWMutex
	assets map[string]*domain.MediaAsset
	order  []string
}

func NewMemoryAssetRepository() *MemoryAssetRepository {
	return &MemoryAssetRepository{
		assets: make(map[string]*domain.MediaAsset),
		order:  make([]string, 0),
	}
}

func (r *MemoryAssetRepository) Save(ctx context.Context, asset *domain.MediaAsset) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *asset
	r.assets[asset.ID] = &cp
	r.order = append(r.order, asset.ID)
	return nil
}

func (r *MemoryAssetRepository) FindByID(ctx context.Context, id string) (*domain.MediaAsset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	asset, found := r.assets[id]
	if !found {
		return nil, domain.ErrAssetNotFound
	}

	cp := *asset
	return &cp, nil
}

func (r *MemoryAssetRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.MediaAsset, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	filtered := make([]*domain.MediaAsset, 0)
	for i := len(r.order) - 1; i >= 0; i-- {
		id := r.order[i]
		asset := r.assets[id]
		if filter.Status != "" && asset.Status != filter.Status {
			continue
		}
		cp := *asset
		filtered = append(filtered, &cp)
	}

	total := len(filtered)
	start := filter.Offset
	if start > total {
		return []*domain.MediaAsset{}, total, nil
	}

	end := start + filter.Limit
	if filter.Limit <= 0 || end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

func (r *MemoryAssetRepository) Update(ctx context.Context, asset *domain.MediaAsset) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, found := r.assets[asset.ID]; !found {
		return domain.ErrAssetNotFound
	}

	asset.UpdatedAt = time.Now().UTC()
	cp := *asset
	r.assets[asset.ID] = &cp
	return nil
}

func (r *MemoryAssetRepository) UpdateStatus(ctx context.Context, id string, status domain.AssetStatus, errMsg string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	asset, found := r.assets[id]
	if !found {
		return domain.ErrAssetNotFound
	}

	asset.Status = status
	asset.ErrorMessage = errMsg
	asset.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *MemoryAssetRepository) SetMetadata(ctx context.Context, id string, metadata *domain.EnrichedMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	asset, found := r.assets[id]
	if !found {
		return domain.ErrAssetNotFound
	}

	asset.Metadata = metadata
	asset.Status = domain.StatusCompleted
	asset.ErrorMessage = ""
	asset.UpdatedAt = time.Now().UTC()
	return nil
}
