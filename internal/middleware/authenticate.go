package middleware

import (
	"strings"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"github.com/eminbekov/fiber-v3-template/internal/service"
	"github.com/gofiber/fiber/v3"
)

func NewAuthenticate(authService *service.AuthService) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		authorizationHeader := strings.TrimSpace(ctx.Get(fiber.HeaderAuthorization))
		sessionToken, parseError := parseBearerToken(authorizationHeader)
		if parseError != nil {
			return parseError
		}

		metadata, sessionError := authService.Session(ctx.Context(), sessionToken)
		if sessionError != nil {
			return sessionError
		}

		ctx.Locals("user_id", metadata.UserID)
		ctx.Locals("session_token", sessionToken)

		return ctx.Next()
	}
}

func parseBearerToken(authorizationHeader string) (string, error) {
	authorizationParts := strings.SplitN(authorizationHeader, " ", 2)
	if len(authorizationParts) != 2 {
		return "", domain.ErrUnauthorized
	}
	if !strings.EqualFold(authorizationParts[0], "Bearer") {
		return "", domain.ErrUnauthorized
	}

	sessionToken := strings.TrimSpace(authorizationParts[1])
	if sessionToken == "" {
		return "", domain.ErrUnauthorized
	}

	return sessionToken, nil
}
