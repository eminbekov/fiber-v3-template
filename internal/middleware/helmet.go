package middleware

import (
	"github.com/gofiber/fiber/v3"
	fiberHelmet "github.com/gofiber/fiber/v3/middleware/helmet"
)

func NewHelmet() fiber.Handler {
	return fiberHelmet.New()
}
