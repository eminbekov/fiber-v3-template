package router

import (
	"github.com/eminbekov/fiber-v3-template/internal/config"
	appHandler "github.com/eminbekov/fiber-v3-template/internal/handler"
	"github.com/eminbekov/fiber-v3-template/internal/middleware"
	"github.com/gofiber/fiber/v3"
)

// New builds the Fiber application with routes and middleware (expand per GO_FIBER_PROJECT_GUIDE.md).
func New(applicationConfiguration *config.Config) *fiber.App {
	application := fiber.New(fiber.Config{
		AppName:      "fiber-v3-template",
		BodyLimit:    applicationConfiguration.BodyLimit,
		ErrorHandler: appHandler.ErrorHandler,
	})
	application.Use(middleware.NewRecover())
	application.Use(middleware.NewRequestID())
	application.Use(middleware.NewRequestLogger())
	application.Use(middleware.NewHelmet())
	application.Use(middleware.NewCORS(applicationConfiguration.CORSAllowOrigins))
	application.Use(middleware.NewBodyLimit(applicationConfiguration.BodyLimit))

	application.Get("/health/live", func(context fiber.Ctx) error {
		return context.SendStatus(fiber.StatusOK)
	})

	application.Get("/health/ready", func(context fiber.Ctx) error {
		return context.SendStatus(fiber.StatusOK)
	})

	application.Get("/", func(context fiber.Ctx) error {
		return context.JSON(fiber.Map{
			"name": "fiber-v3-template",
		})
	})

	return application
}
