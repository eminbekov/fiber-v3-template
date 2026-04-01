package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisPingTimeout = 5 * time.Second

func NewRedisClient(redisURL string) (*redis.Client, error) {
	redisOptions, parseError := redis.ParseURL(redisURL)
	if parseError != nil {
		return nil, fmt.Errorf("cache.NewRedisClient parse url: %w", parseError)
	}

	redisClient := redis.NewClient(redisOptions)

	pingContext, cancelPing := context.WithTimeout(context.Background(), redisPingTimeout)
	defer cancelPing()
	if pingError := redisClient.Ping(pingContext).Err(); pingError != nil {
		_ = redisClient.Close()
		return nil, fmt.Errorf("cache.NewRedisClient ping: %w", pingError)
	}

	return redisClient, nil
}
