package router

import (
	"github.com/eminbekov/fiber-v3-template/internal/config"
	"github.com/eminbekov/fiber-v3-template/internal/dto/response"
	appHandler "github.com/eminbekov/fiber-v3-template/internal/handler"
	v1 "github.com/eminbekov/fiber-v3-template/internal/handler/api/v1"
	"github.com/eminbekov/fiber-v3-template/internal/middleware"
	"github.com/eminbekov/fiber-v3-template/package/health"
	"github.com/gofiber/fiber/v3"
)

// New builds the Fiber application with routes and middleware (expand per GO_FIBER_PROJECT_GUIDE.md).
func New(applicationConfiguration *config.Config) *fiber.App {
	type RootResponse struct {
		Name string `json:"name"`
	}

	application := fiber.New(fiber.Config{
		AppName:      "fiber-v3-template",
		BodyLimit:    applicationConfiguration.BodyLimit,
		ErrorHandler: appHandler.ErrorHandler,
	})
	application.Use(middleware.NewRecover())
	application.Use(middleware.NewMetrics())
	application.Use(middleware.NewRequestID())
	application.Use(middleware.NewRequestLogger())
	application.Use(middleware.NewHelmet())
	application.Use(middleware.NewCORS(applicationConfiguration.CORSAllowOrigins))
	application.Use(middleware.NewBodyLimit(applicationConfiguration.BodyLimit))
	apiV1Handler := v1.NewHandler()
	healthHandler := health.NewHandler()
	apiV1Group := application.Group("/api/v1")

	application.Get("/health/live", healthHandler.Liveness)
	application.Get("/health/ready", healthHandler.Readiness)
	application.Get("/metrics", middleware.MetricsHandler())

	application.Get("/", func(context fiber.Ctx) error {
		return context.JSON(response.Response{
			Data: RootResponse{
				Name: "fiber-v3-template",
			},
		})
	})
	apiV1Group.Get("/ping", apiV1Handler.Ping)

	return application
}
