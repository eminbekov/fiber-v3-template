package v1

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	requestDTO "github.com/eminbekov/fiber-v3-template/internal/dto/request"
	"github.com/eminbekov/fiber-v3-template/internal/dto/response"
	responseV1 "github.com/eminbekov/fiber-v3-template/internal/dto/response/v1"
	"github.com/eminbekov/fiber-v3-template/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login authenticates user credentials and returns a session token.
//
// @Summary      Login
// @Description  Authenticates with username and password and returns a session token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      requestDTO.LoginRequest  true  "Login credentials"
// @Success      200   {object}  response.Response[responseV1.LoginResponse]
// @Failure      400   {object}  response.ErrorResponse  "Validation error"
// @Failure      401   {object}  response.ErrorResponse  "Unauthorized"
// @Failure      500   {object}  response.ErrorResponse  "Internal server error"
// @Router       /v1/auth/login [post]
func (handler *AuthHandler) Login(ctx fiber.Ctx) error {
	var request requestDTO.LoginRequest
	if bindError := ctx.Bind().Body(&request); bindError != nil {
		return domain.ErrValidation
	}
	request.Normalize()

	fieldErrors := requestDTO.ValidateDTO(request)
	if len(fieldErrors) > 0 {
		if jsonError := ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Error: response.ErrorBody{
				Message: "validation failed",
				Details: fieldErrors,
			},
		}); jsonError != nil {
			return fmt.Errorf("authHandler.Login: %w", jsonError)
		}
		return nil
	}

	sessionToken, loginError := handler.authService.Login(
		ctx.Context(),
		request.Username,
		request.Password,
		ctx.IP(),
		ctx.Get(fiber.HeaderUserAgent),
	)
	if loginError != nil {
		return fmt.Errorf("authHandler.Login: %w", loginError)
	}

	sessionExpiresAt := time.Now().UTC().Add(handler.authService.SessionDuration())
	if jsonError := ctx.JSON(response.Response[responseV1.LoginResponse]{
		Data: responseV1.LoginResponse{
			Token:     sessionToken,
			ExpiresAt: sessionExpiresAt,
		},
	}); jsonError != nil {
		return fmt.Errorf("authHandler.Login: %w", jsonError)
	}
	return nil
}

// Logout invalidates the active session token.
//
// @Summary      Logout
// @Description  Invalidates the current authenticated session.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Success      204  {string}  string  "No Content"
// @Failure      401  {object}  response.ErrorResponse  "Unauthorized"
// @Failure      500  {object}  response.ErrorResponse  "Internal server error"
// @Security     BearerAuth
// @Router       /v1/auth/logout [post]
func (handler *AuthHandler) Logout(ctx fiber.Ctx) error {
	sessionToken, isSessionToken := ctx.Locals("session_token").(string)
	if !isSessionToken || sessionToken == "" {
		return domain.ErrUnauthorized
	}

	if logoutError := handler.authService.Logout(ctx.Context(), sessionToken); logoutError != nil {
		return fmt.Errorf("authHandler.Logout: %w", logoutError)
	}

	if sendError := ctx.SendStatus(fiber.StatusNoContent); sendError != nil {
		return fmt.Errorf("authHandler.Logout: %w", sendError)
	}
	return nil
}
