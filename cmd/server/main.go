package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eminbekov/fiber-v3-template/internal/router"
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
	application := router.New()

	go func() {
		<-parentContext.Done()
		slog.Info("shutting down HTTP server")
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if shutdownErr := application.ShutdownWithContext(shutdownContext); shutdownErr != nil {
			slog.Error("HTTP server shutdown", "error", shutdownErr)
		}
	}()

	listenAddress := os.Getenv("HTTP_LISTEN_ADDRESS")
	if listenAddress == "" {
		listenAddress = ":8080"
	}

	slog.Info("HTTP server starting", "address", listenAddress)
	if listenErr := application.Listen(listenAddress); listenErr != nil {
		return fmt.Errorf("listen: %w", listenErr)
	}

	return nil
}
