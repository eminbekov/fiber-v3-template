package middleware

import "github.com/gofiber/fiber/v3"

func NewBodyLimit(maximumBodySizeBytes int) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		contentLength := len(ctx.Body())
		if contentLength > maximumBodySizeBytes {
			return fiber.NewError(fiber.StatusRequestEntityTooLarge, "payload too large")
		}

		return ctx.Next()
	}
}
