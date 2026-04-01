package v1

import (
	"time"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	requestDTO "github.com/eminbekov/fiber-v3-template/internal/dto/request"
	"github.com/eminbekov/fiber-v3-template/internal/dto/response"
	responseV1 "github.com/eminbekov/fiber-v3-template/internal/dto/response/v1"
	"github.com/eminbekov/fiber-v3-template/internal/service"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (handler *AuthHandler) Login(ctx fiber.Ctx) error {
	var request requestDTO.LoginRequest
	if bindError := ctx.Bind().Body(&request); bindError != nil {
		return domain.ErrValidation
	}
	request.Normalize()

	fieldErrors := requestDTO.ValidateDTO(request)
	if len(fieldErrors) > 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Error: response.ErrorBody{
				Message: "validation failed",
				Details: fieldErrors,
			},
		})
	}

	sessionToken, loginError := handler.authService.Login(
		ctx.Context(),
		request.Username,
		request.Password,
		ctx.IP(),
		ctx.Get(fiber.HeaderUserAgent),
	)
	if loginError != nil {
		return loginError
	}

	sessionExpiresAt := time.Now().UTC().Add(handler.authService.SessionDuration())
	return ctx.JSON(response.Response{
		Data: responseV1.LoginResponse{
			Token:     sessionToken,
			ExpiresAt: sessionExpiresAt,
		},
	})
}

func (handler *AuthHandler) Logout(ctx fiber.Ctx) error {
	sessionToken, isSessionToken := ctx.Locals("session_token").(string)
	if !isSessionToken || sessionToken == "" {
		return domain.ErrUnauthorized
	}

	if logoutError := handler.authService.Logout(ctx.Context(), sessionToken); logoutError != nil {
		return logoutError
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
