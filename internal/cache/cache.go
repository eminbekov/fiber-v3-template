package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string, destination any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeleteByPrefix(ctx context.Context, prefix string) error
}
