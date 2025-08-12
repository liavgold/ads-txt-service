package cache

import (
	"context"
	"time"
    "fmt" 

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	cli *redis.Client
}

func NewRedisCache(addr, password string) (*RedisCache, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cli := redis.NewClient(&redis.Options{Addr: addr, Password: password})
	if err := cli.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{cli: cli}, nil
}


func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	s, err := r.cli.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil 
	}
	if err != nil {
		return nil, fmt.Errorf("redis GET failed: %w", err)
	}
	return []byte(s), nil
}


func (r *RedisCache) Set(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	if err := r.cli.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis SET failed: %w", err)
	}
	return nil
}

func (r *RedisCache) Del(ctx context.Context, key string) error {
	if err := r.cli.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis DEL failed: %w", err)
	}
	return nil
}