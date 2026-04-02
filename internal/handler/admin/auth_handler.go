package admin

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"github.com/eminbekov/fiber-v3-template/internal/i18n"
	"github.com/eminbekov/fiber-v3-template/internal/middleware"
	"github.com/eminbekov/fiber-v3-template/internal/service"
	"github.com/gofiber/fiber/v3"
)

// AdminAuthHandler serves HTML login and logout for the admin panel (cookie session).
type AdminAuthHandler struct {
	authService         *service.AuthService
	translator          *i18n.Translator
	sessionDuration     time.Duration
	secureSessionCookie bool
}

// NewAdminAuthHandler constructs an [AdminAuthHandler].
func NewAdminAuthHandler(
	authService *service.AuthService,
	translator *i18n.Translator,
	sessionDuration time.Duration,
	secureSessionCookie bool,
) *AdminAuthHandler {
	return &AdminAuthHandler{
		authService:         authService,
		translator:          translator,
		sessionDuration:     sessionDuration,
		secureSessionCookie: secureSessionCookie,
	}
}

type loginPageViewData struct {
	Title        string
	ErrorMessage string
	T            func(string) string
}

// LoginPage renders the admin login form (unauthenticated).
func (handler *AdminAuthHandler) LoginPage(ctx fiber.Ctx) error {
	sessionToken := strings.TrimSpace(ctx.Cookies(middleware.SessionCookieName))
	if sessionToken != "" {
		if _, sessionError := handler.authService.Session(ctx.Context(), sessionToken); sessionError == nil {
			if redirectError := ctx.Redirect().Status(fiber.StatusFound).To("/admin/dashboard"); redirectError != nil {
				return fmt.Errorf("adminAuthHandler.LoginPage redirect: %w", redirectError)
			}
			return nil
		}
	}

	language, _ := ctx.Locals("language").(string)
	translate := func(key string) string {
		return handler.translator.Translate(language, key)
	}

	errorCode := strings.TrimSpace(ctx.Query("error"))
	errorMessage := ""
	switch errorCode {
	case "invalid_credentials":
		errorMessage = translate("login.error_invalid")
	case "validation":
		errorMessage = translate("login.error_validation")
	}

	if renderError := ctx.Render("admin/login", loginPageViewData{
		Title:        translate("login.title"),
		ErrorMessage: errorMessage,
		T:            translate,
	}, "layouts/auth"); renderError != nil {
		return fmt.Errorf("adminAuthHandler.LoginPage: %w", renderError)
	}

	return nil
}

// Login accepts credentials, creates a session, and sets the session cookie.
func (handler *AdminAuthHandler) Login(ctx fiber.Ctx) error {
	username := strings.TrimSpace(ctx.FormValue("username"))
	password := ctx.FormValue("password")
	if username == "" || strings.TrimSpace(password) == "" {
		if redirectError := ctx.Redirect().Status(fiber.StatusFound).To("/admin/login?error=validation"); redirectError != nil {
			return fmt.Errorf("adminAuthHandler.Login redirect: %w", redirectError)
		}
		return nil
	}

	sessionToken, loginError := handler.authService.Login(
		ctx.Context(),
		username,
		password,
		ctx.IP(),
		ctx.Get(fiber.HeaderUserAgent),
	)
	if loginError != nil {
		if errors.Is(loginError, domain.ErrUnauthorized) {
			if redirectError := ctx.Redirect().Status(fiber.StatusFound).To("/admin/login?error=invalid_credentials"); redirectError != nil {
				return fmt.Errorf("adminAuthHandler.Login redirect: %w", redirectError)
			}
			return nil
		}
		return fmt.Errorf("adminAuthHandler.Login: %w", loginError)
	}

	maxAgeSeconds := int(handler.sessionDuration.Seconds())
	ctx.Cookie(&fiber.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		HTTPOnly: true,
		Secure:   handler.secureSessionCookie,
		SameSite: fiber.CookieSameSiteLaxMode,
		MaxAge:   maxAgeSeconds,
	})

	if redirectError := ctx.Redirect().Status(fiber.StatusFound).To("/admin/dashboard"); redirectError != nil {
		return fmt.Errorf("adminAuthHandler.Login redirect: %w", redirectError)
	}
	return nil
}

// Logout ends the session and clears the session cookie.
func (handler *AdminAuthHandler) Logout(ctx fiber.Ctx) error {
	sessionToken, sessionTokenIsString := ctx.Locals("session_token").(string)
	sessionToken = strings.TrimSpace(sessionToken)
	if !sessionTokenIsString || sessionToken == "" {
		ctx.ClearCookie(middleware.SessionCookieName)
		if redirectError := ctx.Redirect().Status(fiber.StatusFound).To("/admin/login"); redirectError != nil {
			return fmt.Errorf("adminAuthHandler.Logout redirect: %w", redirectError)
		}
		return nil
	}

	if logoutError := handler.authService.Logout(ctx.Context(), sessionToken); logoutError != nil {
		return fmt.Errorf("adminAuthHandler.Logout: %w", logoutError)
	}

	ctx.ClearCookie(middleware.SessionCookieName)

	if redirectError := ctx.Redirect().Status(fiber.StatusFound).To("/admin/login"); redirectError != nil {
		return fmt.Errorf("adminAuthHandler.Logout redirect: %w", redirectError)
	}
	return nil
}
