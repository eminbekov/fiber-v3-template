package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eminbekov/fiber-v3-template/internal/config"
	"github.com/eminbekov/fiber-v3-template/internal/database"
	"github.com/eminbekov/fiber-v3-template/internal/repository/postgres"
	"github.com/eminbekov/fiber-v3-template/internal/router"
	"github.com/eminbekov/fiber-v3-template/package/health"
	"github.com/eminbekov/fiber-v3-template/package/logger"
	"github.com/eminbekov/fiber-v3-template/package/telemetry"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		slog.Error("application exited with error", "error", err)
		os.Exit(1)
	}
}

func run(parentContext context.Context) error {
	applicationConfiguration, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	logger.Setup(applicationConfiguration.LogLevel, applicationConfiguration.Environment)

	shutdownTelemetry, telemetrySetupError := telemetry.Setup(parentContext, applicationConfiguration.OTELExporterEndpoint)
	if telemetrySetupError != nil {
		return fmt.Errorf("telemetry: %w", telemetrySetupError)
	}
	defer func() {
		if telemetryShutdownError := shutdownTelemetry(parentContext); telemetryShutdownError != nil {
			slog.Error("telemetry shutdown", "error", telemetryShutdownError)
		}
	}()

	databasePool, databasePoolError := database.NewPool(parentContext, applicationConfiguration.DatabaseURL)
	if databasePoolError != nil {
		return fmt.Errorf("database pool: %w", databasePoolError)
	}
	defer databasePool.Close()

	userRepository := postgres.NewUserRepository(databasePool)

	application := router.New(applicationConfiguration, router.Dependencies{
		UserRepository: userRepository,
		HealthCheckers: []health.Checker{
			health.NewDatabaseChecker("postgres", databasePool.Ping),
		},
	})

	go func() {
		<-parentContext.Done()
		slog.Info("shutting down HTTP server")
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if shutdownErr := application.ShutdownWithContext(shutdownContext); shutdownErr != nil {
			slog.Error("HTTP server shutdown", "error", shutdownErr)
		}
	}()

	slog.Info("HTTP server starting", "address", applicationConfiguration.HTTPListenAddress)
	if listenErr := application.Listen(applicationConfiguration.HTTPListenAddress); listenErr != nil {
		return fmt.Errorf("listen: %w", listenErr)
	}

	return nil
}
