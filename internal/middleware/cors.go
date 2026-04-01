package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	fiberCORS "github.com/gofiber/fiber/v3/middleware/cors"
)

func NewCORS(allowOriginsValue string) fiber.Handler {
	allowOriginsList := make([]string, 0)
	for _, allowOrigin := range strings.Split(allowOriginsValue, ",") {
		trimmedAllowOrigin := strings.TrimSpace(allowOrigin)
		if trimmedAllowOrigin == "" {
			continue
		}
		allowOriginsList = append(allowOriginsList, trimmedAllowOrigin)
	}

	return fiberCORS.New(fiberCORS.Config{
		AllowOrigins: allowOriginsList,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposeHeaders: []string{
			"X-Request-ID",
			"X-Total-Count",
			"X-Page",
			"X-Page-Size",
			"X-Total-Pages",
			"RateLimit-Limit",
			"RateLimit-Remaining",
			"RateLimit-Reset",
			"Retry-After",
		},
	})
}
