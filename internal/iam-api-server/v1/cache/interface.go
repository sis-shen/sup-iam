package cache

import (
	"context"
	redis "github.com/redis/go-redis/v9"
	"time"
)

type CacheInterface interface {
	Add(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
}

type TokenBlackListInterface interface {
	Add(ctx context.Context, token string) error
	IsBlacklisted(ctx context.Context, token string) (bool, error)
}

type RedisTokenBlackList struct {
	client redis.UniversalClient
	ttl    time.Duration
}

func NewRedisTokenBlackList(client redis.UniversalClient, ttl time.Duration) *RedisTokenBlackList {
	return &RedisTokenBlackList{
		client: client,
		ttl:    ttl,
	}
}

var _ TokenBlackListInterface = (*RedisTokenBlackList)(nil)

func (rt *RedisTokenBlackList) Add(ctx context.Context, token string) error {
	return rt.client.Set(ctx, "blacklist:"+token, "1", rt.ttl).Err()

}

func (rt *RedisTokenBlackList) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	count, err := rt.client.Exists(ctx, "blacklist:"+token).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
