package middleware

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
)

func NewRequestLogger() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		startTime := time.Now()
		nextError := ctx.Next()
		elapsedDuration := time.Since(startTime)

		slog.InfoContext(
			ctx.Context(),
			"http request",
			"method", ctx.Method(),
			"path", ctx.Path(),
			"status_code", ctx.Response().StatusCode(),
			"duration_ms", elapsedDuration.Milliseconds(),
			"request_id", ctx.Get("X-Request-ID"),
		)

		if nextError != nil {
			return fmt.Errorf("requestLogger: %w", nextError)
		}
		return nil
	}
}
