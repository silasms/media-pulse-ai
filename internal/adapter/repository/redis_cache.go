package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/silasms/media-pulse-ai/internal/domain"
)

type CacheClient interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Ping(ctx context.Context) error
}

type memItem struct {
	data      []byte
	expiresAt time.Time
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]memItem
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]memItem),
	}
}

func (m *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, found := m.items[key]
	if !found {
		return nil, domain.ErrAssetNotFound
	}
	if time.Now().After(item.expiresAt) {
		return nil, domain.ErrAssetNotFound
	}

	return item.data, nil
}

func (m *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items[key] = memItem{
		data:      value,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.items, key)
	return nil
}

func (m *MemoryCache) Ping(ctx context.Context) error {
	return nil
}

type RedisCacheClient struct {
	rdb *redis.Client
}

func NewRedisCacheClient(addr, password string, db int) *RedisCacheClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisCacheClient{rdb: rdb}
}

func (r *RedisCacheClient) Get(ctx context.Context, key string) ([]byte, error) {
	return r.rdb.Get(ctx, key).Bytes()
}

func (r *RedisCacheClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.rdb.Set(ctx, key, value, ttl).Err()
}

func (r *RedisCacheClient) Delete(ctx context.Context, key string) error {
	return r.rdb.Del(ctx, key).Err()
}

func (r *RedisCacheClient) Ping(ctx context.Context) error {
	return r.rdb.Ping(ctx).Err()
}

type CachedAssetRepository struct {
	inner domain.AssetRepository
	cache CacheClient
	ttl   time.Duration
}

func NewCachedAssetRepository(inner domain.AssetRepository, cache CacheClient, ttl time.Duration) *CachedAssetRepository {
	return &CachedAssetRepository{
		inner: inner,
		cache: cache,
		ttl:   ttl,
	}
}

func (c *CachedAssetRepository) cacheKey(id string) string {
	return fmt.Sprintf("asset:%s", id)
}

func (c *CachedAssetRepository) FindByID(ctx context.Context, id string) (*domain.MediaAsset, error) {
	if c.cache != nil {
		if data, err := c.cache.Get(ctx, c.cacheKey(id)); err == nil && len(data) > 0 {
			var asset domain.MediaAsset
			if err := json.Unmarshal(data, &asset); err == nil {
				return &asset, nil
			}
		}
	}

	asset, err := c.inner.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if c.cache != nil {
		if data, err := json.Marshal(asset); err == nil {
			_ = c.cache.Set(ctx, c.cacheKey(id), data, c.ttl)
		}
	}

	return asset, nil
}

func (c *CachedAssetRepository) Save(ctx context.Context, asset *domain.MediaAsset) error {
	return c.inner.Save(ctx, asset)
}

func (c *CachedAssetRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.MediaAsset, int, error) {
	return c.inner.List(ctx, filter)
}

func (c *CachedAssetRepository) Update(ctx context.Context, asset *domain.MediaAsset) error {
	if err := c.inner.Update(ctx, asset); err != nil {
		return err
	}
	if c.cache != nil {
		_ = c.cache.Delete(ctx, c.cacheKey(asset.ID))
	}
	return nil
}

func (c *CachedAssetRepository) UpdateStatus(ctx context.Context, id string, status domain.AssetStatus, errMsg string) error {
	if err := c.inner.UpdateStatus(ctx, id, status, errMsg); err != nil {
		return err
	}
	if c.cache != nil {
		_ = c.cache.Delete(ctx, c.cacheKey(id))
	}
	return nil
}

func (c *CachedAssetRepository) SetMetadata(ctx context.Context, id string, metadata *domain.EnrichedMetadata) error {
	if err := c.inner.SetMetadata(ctx, id, metadata); err != nil {
		return err
	}
	if c.cache != nil {
		_ = c.cache.Delete(ctx, c.cacheKey(id))
	}
	return nil
}

func (c *CachedAssetRepository) Invalidate(ctx context.Context, assetID string) error {
	if c.cache != nil {
		return c.cache.Delete(ctx, c.cacheKey(assetID))
	}
	return nil
}
