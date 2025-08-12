package cache

import (
	"fmt"
	"ads-txt-service/internal/config"
)

func InitCache(cfg *config.Config) (Cache, error) {
	switch cfg.CacheBackend {
	case "redis":
		return NewRedisCache(cfg.RedisAddr, cfg.RedisPassword)
	default:
		return nil, fmt.Errorf("unsupported cache backend: %s", cfg.CacheBackend)
	}
}
