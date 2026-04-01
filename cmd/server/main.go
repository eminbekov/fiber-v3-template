package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eminbekov/fiber-v3-template/internal/cache"
	"github.com/eminbekov/fiber-v3-template/internal/config"
	"github.com/eminbekov/fiber-v3-template/internal/database"
	"github.com/eminbekov/fiber-v3-template/internal/handler/admin"
	"github.com/eminbekov/fiber-v3-template/internal/i18n"
	appnats "github.com/eminbekov/fiber-v3-template/internal/nats"
	"github.com/eminbekov/fiber-v3-template/internal/nats/consumers"
	"github.com/eminbekov/fiber-v3-template/internal/repository/postgres"
	"github.com/eminbekov/fiber-v3-template/internal/router"
	"github.com/eminbekov/fiber-v3-template/internal/service"
	"github.com/eminbekov/fiber-v3-template/internal/session"
	"github.com/eminbekov/fiber-v3-template/package/hasher"
	"github.com/eminbekov/fiber-v3-template/package/health"
	"github.com/eminbekov/fiber-v3-template/package/logger"
	"github.com/eminbekov/fiber-v3-template/package/telemetry"
	natsgo "github.com/nats-io/nats.go"
)

// @title           Fiber v3 Template API
// @version         1.0.0
// @description     REST API for authentication, authorization, and user management.
// @contact.name    API Support
// @contact.email   support@example.com
// @host            localhost:8080
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Enter Bearer token as: Bearer <token>
// @schemes        http https
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		slog.Error("application exited with error", "error", err)
		os.Exit(1)
	}
}

func run(parentContext context.Context) error {
	applicationConfiguration, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	logger.Setup(applicationConfiguration.LogLevel, applicationConfiguration.Environment)

	shutdownTelemetry, telemetrySetupError := telemetry.Setup(parentContext, applicationConfiguration.OTELExporterEndpoint)
	if telemetrySetupError != nil {
		return fmt.Errorf("telemetry: %w", telemetrySetupError)
	}
	defer func() {
		if telemetryShutdownError := shutdownTelemetry(parentContext); telemetryShutdownError != nil {
			slog.Error("telemetry shutdown", "error", telemetryShutdownError)
		}
	}()

	databasePool, databasePoolError := database.NewPool(parentContext, applicationConfiguration.DatabaseURL)
	if databasePoolError != nil {
		return fmt.Errorf("database pool: %w", databasePoolError)
	}
	defer databasePool.Close()

	redisClient, redisClientError := cache.NewRedisClient(applicationConfiguration.RedisURL)
	if redisClientError != nil {
		return fmt.Errorf("redis: %w", redisClientError)
	}
	defer func() {
		if closeError := redisClient.Close(); closeError != nil {
			slog.Error("redis close", "error", closeError)
		}
	}()

	natsConnection, jetStream, natsConnectError := appnats.Connect(parentContext, applicationConfiguration.NATSURL)
	if natsConnectError != nil {
		return fmt.Errorf("nats: %w", natsConnectError)
	}
	defer natsConnection.Close()

	notificationConsumer := consumers.NewNotificationConsumer(jetStream)
	auditLogConsumer := consumers.NewAuditLogConsumer(jetStream)
	go func() {
		if runError := notificationConsumer.Run(parentContext); runError != nil {
			slog.Error("notification consumer failed", "error", runError)
		}
	}()
	go func() {
		if runError := auditLogConsumer.Run(parentContext); runError != nil {
			slog.Error("audit consumer failed", "error", runError)
		}
	}()

	userRepository := postgres.NewUserRepository(databasePool)
	roleRepository := postgres.NewRoleRepository(databasePool)
	permissionRepository := postgres.NewPermissionRepository(databasePool)
	applicationCache := cache.NewRedisCache(redisClient)
	passwordHasher := hasher.NewArgon2ID()
	sessionStore := session.NewRedisStore(redisClient, applicationConfiguration.SessionDuration)
	userService := service.NewUserService(userRepository, roleRepository, applicationCache, passwordHasher)
	authService := service.NewAuthService(
		userRepository,
		sessionStore,
		passwordHasher,
		applicationConfiguration.SessionDuration,
	)
	authorizationService := service.NewAuthorizationService(permissionRepository, applicationCache)
	translator, translatorError := i18n.NewTranslator("en")
	if translatorError != nil {
		return fmt.Errorf("translator: %w", translatorError)
	}
	dashboardHandler := admin.NewDashboardHandler(translator)

	application := router.New(applicationConfiguration, router.Dependencies{
		UserRepository:       userRepository,
		RoleRepository:       roleRepository,
		PermissionRepository: permissionRepository,
		UserService:          userService,
		AuthService:          authService,
		AuthorizationService: authorizationService,
		DashboardHandler:     dashboardHandler,
		Translator:           translator,
		Cache:                applicationCache,
		HealthCheckers: []health.Checker{
			health.NewDatabaseChecker("postgres", databasePool.Ping),
			health.NewRedisChecker("redis", func(ctx context.Context) error {
				return redisClient.Ping(ctx).Err()
			}),
			health.NewNATSChecker("nats", func(context.Context) error {
				if natsConnection.Status() != natsgo.CONNECTED {
					return fmt.Errorf("nats not connected: %s", natsConnection.Status().String())
				}
				return natsConnection.FlushTimeout(2 * time.Second)
			}),
		},
	})

	go func() {
		<-parentContext.Done()
		slog.Info("shutting down HTTP server")
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if shutdownErr := application.ShutdownWithContext(shutdownContext); shutdownErr != nil {
			slog.Error("HTTP server shutdown", "error", shutdownErr)
		}
	}()

	slog.Info("HTTP server starting", "address", applicationConfiguration.HTTPListenAddress)
	if listenErr := application.Listen(applicationConfiguration.HTTPListenAddress); listenErr != nil {
		return fmt.Errorf("listen: %w", listenErr)
	}

	return nil
}
