package handler

import (
	"errors"
	"log/slog"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"github.com/eminbekov/fiber-v3-template/internal/dto/response"
	"github.com/gofiber/fiber/v3"
)

func ErrorHandler(ctx fiber.Ctx, err error) error {
	statusCode := fiber.StatusInternalServerError
	message := "internal server error"

	switch {
	case errors.Is(err, domain.ErrNotFound):
		statusCode = fiber.StatusNotFound
		message = "resource not found"
	case errors.Is(err, domain.ErrUnauthorized):
		statusCode = fiber.StatusUnauthorized
		message = "unauthorized"
	case errors.Is(err, domain.ErrForbidden):
		statusCode = fiber.StatusForbidden
		message = "forbidden"
	case errors.Is(err, domain.ErrConflict):
		statusCode = fiber.StatusConflict
		message = "resource already exists"
	case errors.Is(err, domain.ErrValidation):
		statusCode = fiber.StatusBadRequest
		message = "validation failed"
	}

	slog.ErrorContext(
		ctx.Context(),
		"request failed",
		"error", err,
		"path", ctx.Path(),
		"method", ctx.Method(),
		"status_code", statusCode,
		"request_id", ctx.Get("X-Request-ID"),
	)

	return ctx.Status(statusCode).JSON(response.ErrorResponse{
		Error: response.ErrorBody{
			Message: message,
		},
	})
}
