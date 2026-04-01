package middleware

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	fiberRecover "github.com/gofiber/fiber/v3/middleware/recover"
)

func NewRecover() fiber.Handler {
	return fiberRecover.New(fiberRecover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(ctx fiber.Ctx, recovered any) {
			slog.ErrorContext(
				ctx.Context(),
				"panic recovered",
				"panic", recovered,
				"path", ctx.Path(),
				"method", ctx.Method(),
			)
		},
	})
}
