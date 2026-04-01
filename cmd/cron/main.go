package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eminbekov/fiber-v3-template/internal/cache"
	"github.com/eminbekov/fiber-v3-template/internal/config"
	"github.com/eminbekov/fiber-v3-template/internal/cron"
	"github.com/eminbekov/fiber-v3-template/internal/database"
	"github.com/eminbekov/fiber-v3-template/package/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if runError := run(ctx); runError != nil {
		slog.Error("cron exited with error", "error", runError)
		os.Exit(1)
	}
}

func run(parentContext context.Context) error {
	applicationConfiguration, configError := config.Load()
	if configError != nil {
		return fmt.Errorf("config: %w", configError)
	}

	logger.Setup(applicationConfiguration.LogLevel, applicationConfiguration.Environment)

	databasePool, databasePoolError := database.NewPool(parentContext, applicationConfiguration.DatabaseURL)
	if databasePoolError != nil {
		return fmt.Errorf("database pool: %w", databasePoolError)
	}
	defer databasePool.Close()

	redisClient, redisClientError := cache.NewRedisClient(applicationConfiguration.RedisURL)
	if redisClientError != nil {
		return fmt.Errorf("redis: %w", redisClientError)
	}
	defer func() {
		if closeError := redisClient.Close(); closeError != nil {
			slog.Error("redis close", "error", closeError)
		}
	}()

	scheduler := cron.NewScheduler()
	scheduler.Register(cron.Job{
		Name:     "database-connectivity-check",
		Schedule: 10 * time.Minute,
		Run: func(ctx context.Context) error {
			jobContext, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return databasePool.Ping(jobContext)
		},
	})
	scheduler.Register(cron.Job{
		Name:     "redis-connectivity-check",
		Schedule: 5 * time.Minute,
		Run: func(ctx context.Context) error {
			jobContext, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return redisClient.Ping(jobContext).Err()
		},
	})

	scheduler.Start(parentContext)
	<-parentContext.Done()
	return nil
}
