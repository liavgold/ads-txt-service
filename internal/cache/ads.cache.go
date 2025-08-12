 package cache

import (
    "context"
    "encoding/json"
    "time"
    "ads-txt-service/internal/models"
)

type AdsCache struct {
	cache Cache
}

func NewAdsCache(c Cache) *AdsCache {
	return &AdsCache{cache: c}
}

func (a *AdsCache) GetAds(ctx context.Context, key string) (*models.AdsResponse, bool) {
	b, err := a.cache.Get(ctx, key)
	if err != nil || b == nil {
		return nil, false
	}

	var resp models.AdsResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, false
	}
	return &resp, true
}

func (a *AdsCache) SetAds(ctx context.Context, key string, resp *models.AdsResponse, ttl time.Duration) error {
	b, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return a.cache.Set(ctx, key, b, ttl)
}