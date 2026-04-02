package router

import (
	"time"

	_ "github.com/eminbekov/fiber-v3-template/docs"
	"github.com/eminbekov/fiber-v3-template/internal/cache"
	"github.com/eminbekov/fiber-v3-template/internal/config"
	"github.com/eminbekov/fiber-v3-template/internal/dto/response"
	appHandler "github.com/eminbekov/fiber-v3-template/internal/handler"
	"github.com/eminbekov/fiber-v3-template/internal/handler/admin"
	v1 "github.com/eminbekov/fiber-v3-template/internal/handler/api/v1"
	"github.com/eminbekov/fiber-v3-template/internal/i18n"
	"github.com/eminbekov/fiber-v3-template/internal/middleware"
	"github.com/eminbekov/fiber-v3-template/internal/repository"
	"github.com/eminbekov/fiber-v3-template/internal/service"
	appwebsocket "github.com/eminbekov/fiber-v3-template/internal/websocket"
	"github.com/eminbekov/fiber-v3-template/package/health"
	fiberwebsocket "github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/template/html/v2"
)

type Dependencies struct {
	UserRepository       repository.UserRepository
	RoleRepository       repository.RoleRepository
	PermissionRepository repository.PermissionRepository
	UserService          *service.UserService
	AuthService          *service.AuthService
	AuthorizationService *service.AuthorizationService
	AdminAuthHandler     *admin.AdminAuthHandler
	DashboardHandler     *admin.DashboardHandler
	Translator           *i18n.Translator
	Cache                cache.Cache
	FileService          *service.FileService
	WebSocketHub         *appwebsocket.Hub
	HealthCheckers       []health.Checker
}

// New builds the Fiber application with routes and middleware (expand per GO_FIBER_PROJECT_GUIDE.md).
func New(applicationConfiguration *config.Config, dependencies Dependencies) *fiber.App {
	type RootResponse struct {
		Name string `json:"name"`
	}

	templateEngine := html.New(applicationConfiguration.ViewsPath, ".html")
	templateEngine.AddFunc("formatDate", func(value time.Time) string {
		return value.Format("2006-01-02 15:04")
	})

	application := fiber.New(fiber.Config{
		AppName:      "fiber-v3-template",
		BodyLimit:    applicationConfiguration.BodyLimit,
		ErrorHandler: appHandler.ErrorHandler,
		Views:        templateEngine,
	})
	registerMiddleware(application, applicationConfiguration)
	apiV1Handler := v1.NewHandler()
	authHandler := v1.NewAuthHandler(dependencies.AuthService)
	userHandler := v1.NewUserHandler(dependencies.UserService, dependencies.Translator)
	fileHandler := v1.NewFileHandler(dependencies.FileService, dependencies.Translator)
	webSocketHandler := appwebsocket.NewHandler(dependencies.WebSocketHub)
	dashboardHandler := dependencies.DashboardHandler
	healthHandler := health.NewHandler(dependencies.HealthCheckers...)
	adminAuthHandler := dependencies.AdminAuthHandler
	application.Get("/admin/login", adminAuthHandler.LoginPage)
	application.Post("/admin/login", adminAuthHandler.Login)

	apiV1Group := application.Group("/api/v1")
	protectedAPIGroup := apiV1Group.Group("", middleware.NewAuthenticate(dependencies.AuthService))
	adminGroup := application.Group("/admin", middleware.NewAdminAuthenticate(dependencies.AuthService))

	application.Get("/health/live", healthHandler.Liveness)
	application.Get("/health/ready", healthHandler.Readiness)
	application.Get("/metrics", middleware.MetricsHandler())
	if applicationConfiguration.Environment != "production" {
		application.Get("/swagger/*", swaggo.HandlerDefault)
	}

	application.Get("/", func(context fiber.Ctx) error {
		return context.JSON(response.Response{
			Data: RootResponse{
				Name: "fiber-v3-template",
			},
		})
	})
	adminGroup.Get("/dashboard", dashboardHandler.Index)
	adminGroup.Post("/logout", adminAuthHandler.Logout)
	apiV1Group.Get("/ping", apiV1Handler.Ping)
	apiV1Group.Post("/auth/login", authHandler.Login)
	protectedAPIGroup.Post("/auth/logout", authHandler.Logout)
	application.Get(
		"/api/files/:filename",
		middleware.NewSignedURLValidator([]byte(applicationConfiguration.FileSigningKey)),
		fileHandler.Download,
	)
	protectedAPIGroup.Post("/files", middleware.RequirePermission(dependencies.AuthorizationService, "files", "create"), fileHandler.Upload)
	protectedAPIGroup.Post("/users", middleware.RequirePermission(dependencies.AuthorizationService, "users", "create"), userHandler.Create)
	protectedAPIGroup.Get("/users", middleware.RequirePermission(dependencies.AuthorizationService, "users", "read"), userHandler.List)
	protectedAPIGroup.Get("/users/:id", middleware.RequirePermission(dependencies.AuthorizationService, "users", "read"), userHandler.FindByID)
	protectedAPIGroup.Put("/users/:id", middleware.RequirePermission(dependencies.AuthorizationService, "users", "update"), userHandler.Update)
	protectedAPIGroup.Delete("/users/:id", middleware.RequirePermission(dependencies.AuthorizationService, "users", "delete"), userHandler.Delete)
	application.Get("/ws", appwebsocket.RequireUpgrade, fiberwebsocket.New(webSocketHandler.HandleConnection))

	return application
}

func registerMiddleware(application *fiber.App, applicationConfiguration *config.Config) {
	application.Use(middleware.NewRecover())
	application.Use(middleware.NewMetrics())
	application.Use(middleware.NewRequestID())
	application.Use(middleware.NewRequestLogger())
	application.Use(middleware.NewHelmet())
	application.Use(middleware.NewCORS(applicationConfiguration.CORSAllowOrigins))
	application.Use(middleware.NewBodyLimit(applicationConfiguration.BodyLimit))
	application.Use(middleware.LanguageDetector([]string{"en", "uz", "ru"}, "en"))
}
