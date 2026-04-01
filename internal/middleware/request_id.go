package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func NewRequestID() fiber.Handler {
	return requestid.New(requestid.Config{
		Header: "X-Request-ID",
	})
}
