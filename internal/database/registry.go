package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Registry stores named PostgreSQL pools.
type Registry struct {
	pools map[string]*pgxpool.Pool
}

// NewRegistry creates pools for each named database URL.
func NewRegistry(ctx context.Context, databaseURLs map[string]string) (*Registry, error) {
	registry := &Registry{
		pools: make(map[string]*pgxpool.Pool, len(databaseURLs)),
	}

	for poolName, databaseURL := range databaseURLs {
		pool, poolError := NewPool(ctx, databaseURL)
		if poolError != nil {
			registry.Close()
			return nil, fmt.Errorf("database.NewRegistry: pool %s: %w", poolName, poolError)
		}

		registry.pools[poolName] = pool
	}

	return registry, nil
}

// Pool returns a named pool or nil if it is missing.
func (registry *Registry) Pool(name string) *pgxpool.Pool {
	return registry.pools[name]
}

// Close closes all pools.
func (registry *Registry) Close() {
	for _, pool := range registry.pools {
		pool.Close()
	}
}

// HealthCheck returns ping errors for each pool.
func (registry *Registry) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error, len(registry.pools))

	for name, pool := range registry.pools {
		results[name] = pool.Ping(ctx)
	}

	return results
}
