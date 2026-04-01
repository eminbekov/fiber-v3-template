package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxConnections           = int32(50)
	defaultMinConnections           = int32(10)
	defaultMaxConnectionLifetime    = time.Hour
	defaultMaxConnectionIdleTime    = 30 * time.Minute
	defaultConnectionLifetimeJitter = 5 * time.Minute
	defaultHealthCheckPeriod        = 30 * time.Second
)

// NewPool creates and verifies a PostgreSQL connection pool.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	poolConfiguration, parseError := pgxpool.ParseConfig(databaseURL)
	if parseError != nil {
		return nil, fmt.Errorf("database.NewPool: parse config: %w", parseError)
	}

	poolConfiguration.MaxConns = defaultMaxConnections
	poolConfiguration.MinConns = defaultMinConnections
	poolConfiguration.MaxConnLifetime = defaultMaxConnectionLifetime
	poolConfiguration.MaxConnIdleTime = defaultMaxConnectionIdleTime
	poolConfiguration.MaxConnLifetimeJitter = defaultConnectionLifetimeJitter
	poolConfiguration.HealthCheckPeriod = defaultHealthCheckPeriod
	poolConfiguration.ConnConfig.Tracer = &queryTracer{slowThreshold: 100 * time.Millisecond}

	pool, newPoolError := pgxpool.NewWithConfig(ctx, poolConfiguration)
	if newPoolError != nil {
		return nil, fmt.Errorf("database.NewPool: new pool: %w", newPoolError)
	}

	if pingError := pool.Ping(ctx); pingError != nil {
		pool.Close()
		return nil, fmt.Errorf("database.NewPool: ping: %w", pingError)
	}

	return pool, nil
}
