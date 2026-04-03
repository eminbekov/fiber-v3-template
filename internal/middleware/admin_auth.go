package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/eminbekov/fiber-v3-template/internal/service"
)

// SessionCookieName is the HTTP cookie name that stores the session token for browser admin flows.
const SessionCookieName = "session_token"

const adminLoginPath = "/admin/login"

// NewAdminAuthenticate validates the session cookie and attaches user identity to [fiber.Ctx.Locals].
// Unauthenticated requests are redirected to the admin login page (HTML/browser flow).
func NewAdminAuthenticate(authService *service.AuthService) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		sessionToken := strings.TrimSpace(ctx.Cookies(SessionCookieName))
		if sessionToken == "" {
			return ctx.Redirect().Status(fiber.StatusFound).To(adminLoginPath)
		}

		metadata, sessionError := authService.Session(ctx.Context(), sessionToken)
		if sessionError != nil {
			return ctx.Redirect().Status(fiber.StatusFound).To(adminLoginPath)
		}

		ctx.Locals("user_id", metadata.UserID)
		ctx.Locals("session_token", sessionToken)

		if nextError := ctx.Next(); nextError != nil {
			return fmt.Errorf("adminAuthenticate: %w", nextError)
		}

		return nil
	}
}
