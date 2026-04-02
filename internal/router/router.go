package router

import (
	"time"

	// [module:swagger:start]
	_ "github.com/eminbekov/fiber-v3-template/docs"
	// [module:swagger:end]
	"github.com/eminbekov/fiber-v3-template/internal/cache"
	"github.com/eminbekov/fiber-v3-template/internal/config"
	appHandler "github.com/eminbekov/fiber-v3-template/internal/handler"
	// [module:admin:start]
	"github.com/eminbekov/fiber-v3-template/internal/handler/admin"
	// [module:admin:end]
	v1 "github.com/eminbekov/fiber-v3-template/internal/handler/api/v1"
	// [module:web:start]
	"github.com/eminbekov/fiber-v3-template/internal/handler/web"
	// [module:web:end]
	// [module:i18n:start]
	"github.com/eminbekov/fiber-v3-template/internal/i18n"
	// [module:i18n:end]
	"github.com/eminbekov/fiber-v3-template/internal/middleware"
	"github.com/eminbekov/fiber-v3-template/internal/repository"
	"github.com/eminbekov/fiber-v3-template/internal/service"
	// [module:websocket:start]
	appwebsocket "github.com/eminbekov/fiber-v3-template/internal/websocket"
	// [module:websocket:end]
	"github.com/eminbekov/fiber-v3-template/package/health"
	// [module:swagger:start]
	"github.com/gofiber/contrib/v3/swaggo"
	// [module:swagger:end]
	// [module:websocket:start]
	fiberwebsocket "github.com/gofiber/contrib/v3/websocket"
	// [module:websocket:end]
	"github.com/gofiber/fiber/v3"
	// [module:views:start]
	"github.com/gofiber/template/html/v2"
	// [module:views:end]
)

type Dependencies struct {
	UserRepository       repository.UserRepository
	RoleRepository       repository.RoleRepository
	PermissionRepository repository.PermissionRepository
	UserService          *service.UserService
	AuthService          *service.AuthService
	AuthorizationService *service.AuthorizationService
	// [module:admin:start]
	AdminAuthHandler *admin.AdminAuthHandler
	DashboardHandler *admin.DashboardHandler
	// [module:admin:end]
	// [module:web:start]
	WelcomeHandler *web.WelcomeHandler
	// [module:web:end]
	// [module:i18n:start]
	Translator *i18n.Translator
	// [module:i18n:end]
	Cache cache.Cache
	// [module:storage:start]
	FileService *service.FileService
	// [module:storage:end]
	// [module:websocket:start]
	WebSocketHub *appwebsocket.Hub
	// [module:websocket:end]
	HealthCheckers []health.Checker
}

// New builds the Fiber application with routes and middleware (expand per GO_FIBER_PROJECT_GUIDE.md).
func New(applicationConfiguration *config.Config, dependencies Dependencies) *fiber.App {
	// [module:views:start]
	templateEngine := html.New(applicationConfiguration.ViewsPath, ".html")
	templateEngine.AddFunc("formatDate", func(value time.Time) string {
		return value.Format("2006-01-02 15:04")
	})
	// [module:views:end]

	application := fiber.New(fiber.Config{
		AppName:      "fiber-v3-template",
		BodyLimit:    applicationConfiguration.BodyLimit,
		ErrorHandler: appHandler.ErrorHandler,
		// [module:views:start]
		Views: templateEngine,
		// [module:views:end]
	})
	registerMiddleware(application, applicationConfiguration)
	apiV1Handler := v1.NewHandler()
	authHandler := v1.NewAuthHandler(dependencies.AuthService)
	// [module:i18n:start]
	userHandler := v1.NewUserHandler(dependencies.UserService, dependencies.Translator)
	// [module:i18n:end]
	// [module:storage:start]
	fileHandler := v1.NewFileHandler(dependencies.FileService, dependencies.Translator)
	// [module:storage:end]
	// [module:websocket:start]
	webSocketHandler := appwebsocket.NewHandler(dependencies.WebSocketHub)
	// [module:websocket:end]
	// [module:admin:start]
	dashboardHandler := dependencies.DashboardHandler
	// [module:admin:end]
	healthHandler := health.NewHandler(dependencies.HealthCheckers...)
	// [module:admin:start]
	adminAuthHandler := dependencies.AdminAuthHandler
	application.Get("/admin/login", adminAuthHandler.LoginPage)
	application.Post("/admin/login", adminAuthHandler.Login)
	// [module:admin:end]

	apiV1Group := application.Group("/api/v1")
	protectedAPIGroup := apiV1Group.Group("", middleware.NewAuthenticate(dependencies.AuthService))
	// [module:admin:start]
	adminGroup := application.Group("/admin", middleware.NewAdminAuthenticate(dependencies.AuthService))
	// [module:admin:end]

	application.Get("/health/live", healthHandler.Liveness)
	application.Get("/health/ready", healthHandler.Readiness)
	application.Get("/metrics", middleware.MetricsHandler())
	// [module:swagger:start]
	if applicationConfiguration.Environment != "production" {
		application.Get("/swagger/*", swaggo.HandlerDefault)
	}
	// [module:swagger:end]

	// [module:web:start]
	application.Get("/", dependencies.WelcomeHandler.Index)
	// [module:web:end]
	// [module:admin:start]
	adminGroup.Get("/dashboard", dashboardHandler.Index)
	adminGroup.Post("/logout", adminAuthHandler.Logout)
	// [module:admin:end]
	apiV1Group.Get("/ping", apiV1Handler.Ping)
	apiV1Group.Post("/auth/login", authHandler.Login)
	protectedAPIGroup.Post("/auth/logout", authHandler.Logout)
	// [module:storage:start]
	application.Get(
		"/api/files/:filename",
		middleware.NewSignedURLValidator([]byte(applicationConfiguration.FileSigningKey)),
		fileHandler.Download,
	)
	protectedAPIGroup.Post("/files", middleware.RequirePermission(dependencies.AuthorizationService, "files", "create"), fileHandler.Upload)
	// [module:storage:end]
	protectedAPIGroup.Post("/users", middleware.RequirePermission(dependencies.AuthorizationService, "users", "create"), userHandler.Create)
	protectedAPIGroup.Get("/users", middleware.RequirePermission(dependencies.AuthorizationService, "users", "read"), userHandler.List)
	protectedAPIGroup.Get("/users/:id", middleware.RequirePermission(dependencies.AuthorizationService, "users", "read"), userHandler.FindByID)
	protectedAPIGroup.Put("/users/:id", middleware.RequirePermission(dependencies.AuthorizationService, "users", "update"), userHandler.Update)
	protectedAPIGroup.Delete("/users/:id", middleware.RequirePermission(dependencies.AuthorizationService, "users", "delete"), userHandler.Delete)
	// [module:websocket:start]
	application.Get("/ws", appwebsocket.RequireUpgrade, fiberwebsocket.New(webSocketHandler.HandleConnection))
	// [module:websocket:end]

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
	// [module:i18n:start]
	application.Use(middleware.LanguageDetector([]string{"en", "uz", "ru"}, "en"))
	// [module:i18n:end]
}
