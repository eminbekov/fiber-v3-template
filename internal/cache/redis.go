package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

const scanBatchSize = int64(100)

type RedisCache struct {
	redisClient *redis.Client
}

func NewRedisCache(redisClient *redis.Client) *RedisCache {
	return &RedisCache{
		redisClient: redisClient,
	}
}

func (cacheImplementation *RedisCache) Get(ctx context.Context, key string, destination any) error {
	value, getError := cacheImplementation.redisClient.Get(ctx, key).Bytes()
	if getError != nil {
		if errors.Is(getError, redis.Nil) {
			return ErrCacheMiss
		}
		return fmt.Errorf("redisCache.Get: %w", getError)
	}

	if unmarshalError := sonic.Unmarshal(value, destination); unmarshalError != nil {
		return fmt.Errorf("redisCache.Get unmarshal: %w", unmarshalError)
	}

	return nil
}

func (cacheImplementation *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	serializedValue, marshalError := sonic.Marshal(value)
	if marshalError != nil {
		return fmt.Errorf("redisCache.Set marshal: %w", marshalError)
	}

	if setError := cacheImplementation.redisClient.Set(ctx, key, serializedValue, ttl).Err(); setError != nil {
		return fmt.Errorf("redisCache.Set: %w", setError)
	}

	return nil
}

func (cacheImplementation *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	if deleteError := cacheImplementation.redisClient.Del(ctx, keys...).Err(); deleteError != nil {
		return fmt.Errorf("redisCache.Delete: %w", deleteError)
	}

	return nil
}

func (cacheImplementation *RedisCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	cursor := uint64(0)
	for {
		keys, nextCursor, scanError := cacheImplementation.redisClient.Scan(ctx, cursor, prefix+"*", scanBatchSize).Result()
		if scanError != nil {
			return fmt.Errorf("redisCache.DeleteByPrefix scan: %w", scanError)
		}

		if len(keys) > 0 {
			pipeline := cacheImplementation.redisClient.Pipeline()
			pipeline.Del(ctx, keys...)
			if _, executeError := pipeline.Exec(ctx); executeError != nil {
				return fmt.Errorf("redisCache.DeleteByPrefix delete: %w", executeError)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}
