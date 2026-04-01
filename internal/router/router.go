package router

import (
	appHandler "github.com/eminbekov/fiber-v3-template/internal/handler"
	"github.com/gofiber/fiber/v3"
)

// New builds the Fiber application with routes and middleware (expand per GO_FIBER_PROJECT_GUIDE.md).
func New() *fiber.App {
	application := fiber.New(fiber.Config{
		AppName:      "fiber-v3-template",
		ErrorHandler: appHandler.ErrorHandler,
	})

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
