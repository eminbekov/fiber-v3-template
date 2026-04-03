package middleware

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/eminbekov/fiber-v3-template/internal/dto/response"
	"github.com/eminbekov/fiber-v3-template/internal/storage"
)

// NewSignedURLValidator ensures ?token and ?expires match the HMAC for :filename.
func NewSignedURLValidator(signingKey []byte) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		filename := ctx.Params("filename")
		token := ctx.Query("token")
		expiresQuery := ctx.Query("expires")

		expiresUnix, parseError := strconv.ParseInt(expiresQuery, 10, 64)
		if parseError != nil || !storage.ValidateSignedURL(filename, token, expiresUnix, signingKey) {
			return ctx.Status(fiber.StatusForbidden).JSON(response.ErrorResponse{
				Error: response.ErrorBody{
					Message: "forbidden",
				},
			})
		}

		return ctx.Next()
	}
}
