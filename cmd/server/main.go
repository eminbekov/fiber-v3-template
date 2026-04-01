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
	"github.com/eminbekov/fiber-v3-template/internal/router"
	"github.com/eminbekov/fiber-v3-template/package/logger"
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

	application := router.New(applicationConfiguration)

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
