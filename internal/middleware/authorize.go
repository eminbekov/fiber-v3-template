package middleware

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofrs/uuid/v5"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"github.com/eminbekov/fiber-v3-template/internal/service"
)

func RequirePermission(
	authorizationService *service.AuthorizationService,
	resource string,
	action string,
) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		userIDValue := ctx.Locals("user_id")
		userID, isUserID := userIDValue.(uuid.UUID)
		if !isUserID {
			return domain.ErrUnauthorized
		}

		hasPermission, permissionError := authorizationService.HasPermission(ctx.Context(), userID, resource, action)
		if permissionError != nil {
			slog.ErrorContext(ctx.Context(), "permission check failed", "error", permissionError)
			return domain.ErrInternal
		}

		if !hasPermission {
			return domain.ErrForbidden
		}

		return ctx.Next()
	}
}
