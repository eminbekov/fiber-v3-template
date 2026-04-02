package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	userv1 "github.com/eminbekov/fiber-v3-template/gen/proto/user/v1"
	"github.com/eminbekov/fiber-v3-template/internal/cache"
	"github.com/eminbekov/fiber-v3-template/internal/config"
	"github.com/eminbekov/fiber-v3-template/internal/cron"
	"github.com/eminbekov/fiber-v3-template/internal/database"
	internalgrpc "github.com/eminbekov/fiber-v3-template/internal/grpc"
	"github.com/eminbekov/fiber-v3-template/internal/handler/admin"
	"github.com/eminbekov/fiber-v3-template/internal/handler/web"
	"github.com/eminbekov/fiber-v3-template/internal/i18n"
	appnats "github.com/eminbekov/fiber-v3-template/internal/nats"
	"github.com/eminbekov/fiber-v3-template/internal/nats/consumers"
	"github.com/eminbekov/fiber-v3-template/internal/repository/postgres"
	"github.com/eminbekov/fiber-v3-template/internal/router"
	"github.com/eminbekov/fiber-v3-template/internal/service"
	"github.com/eminbekov/fiber-v3-template/internal/session"
	"github.com/eminbekov/fiber-v3-template/internal/storage"
	appwebsocket "github.com/eminbekov/fiber-v3-template/internal/websocket"
	"github.com/eminbekov/fiber-v3-template/package/hasher"
	"github.com/eminbekov/fiber-v3-template/package/health"
	"github.com/eminbekov/fiber-v3-template/package/logger"
	"github.com/eminbekov/fiber-v3-template/package/telemetry"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
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

	if err := run(ctx); err != nil {
		slog.Error("application exited with error", "error", err)
		stop()
		os.Exit(1)
	}
	stop()
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

	fileStorage, fileStorageError := storage.NewFromApplicationConfig(parentContext, applicationConfiguration)
	if fileStorageError != nil {
		return fmt.Errorf("file storage: %w", fileStorageError)
	}

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

	wiring, wiringError := wireApplication(parentContext, applicationConfiguration, databasePool, redisClient, natsConnection, jetStream, fileStorage)
	if wiringError != nil {
		return wiringError
	}
	defer func() {
		if closeError := wiring.grpcListener.Close(); closeError != nil {
			slog.Error("grpc listener close", "error", closeError)
		}
	}()

	group, groupContext := errgroup.WithContext(parentContext)
	registerWorkers(group, groupContext, applicationConfiguration, wiring.application, wiring.grpcServer, wiring.grpcListener, wiring.scheduler, wiring.notificationConsumer, wiring.auditLogConsumer, wiring.webSocketHub)

	if waitError := group.Wait(); waitError != nil && !errors.Is(waitError, context.Canceled) {
		return fmt.Errorf("run: %w", waitError)
	}

	return nil
}

type applicationWiring struct {
	application          *fiber.App
	grpcServer           *grpc.Server
	grpcListener         net.Listener
	scheduler            *cron.Scheduler
	notificationConsumer *consumers.NotificationConsumer
	auditLogConsumer     *consumers.AuditLogConsumer
	webSocketHub         *appwebsocket.Hub
}

func wireApplication(
	parentContext context.Context,
	applicationConfiguration *config.Config,
	databasePool *pgxpool.Pool,
	redisClient *redis.Client,
	natsConnection *natsgo.Conn,
	jetStream jetstream.JetStream,
	fileStorage storage.FileStorage,
) (*applicationWiring, error) {
	notificationConsumer := consumers.NewNotificationConsumer(jetStream)
	auditLogConsumer := consumers.NewAuditLogConsumer(jetStream)

	userRepository := postgres.NewUserRepository(databasePool)
	roleRepository := postgres.NewRoleRepository(databasePool)
	permissionRepository := postgres.NewPermissionRepository(databasePool)
	applicationCache := cache.NewRedisCache(redisClient)
	webSocketHub := appwebsocket.NewHub(redisClient)
	passwordHasher := hasher.NewArgon2ID()
	sessionStore := session.NewRedisStore(redisClient, applicationConfiguration.SessionDuration)
	fileService := service.NewFileService(fileStorage, applicationConfiguration.SignedURLTTL)
	userService := service.NewUserService(userRepository, roleRepository, applicationCache, passwordHasher)
	authService := service.NewAuthService(
		userRepository,
		sessionStore,
		passwordHasher,
		applicationConfiguration.SessionDuration,
	)
	authorizationService := service.NewAuthorizationService(permissionRepository, applicationCache)
	scheduler := cron.NewScheduler()
	scheduler.Register(cron.Job{
		Name:     "redis-connectivity-check",
		Schedule: 5 * time.Minute,
		Run: func(ctx context.Context) error {
			jobContext, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return redisClient.Ping(jobContext).Err()
		},
	})

	translator, translatorError := i18n.NewTranslator("en")
	if translatorError != nil {
		return nil, fmt.Errorf("translator: %w", translatorError)
	}
	dashboardHandler := admin.NewDashboardHandler(translator)
	welcomeHandler := web.NewWelcomeHandler(translator)
	secureSessionCookie := applicationConfiguration.Environment == "production"
	adminAuthHandler := admin.NewAdminAuthHandler(
		authService,
		translator,
		applicationConfiguration.SessionDuration,
		secureSessionCookie,
	)
	userGRPCServer := internalgrpc.NewUserServer(userService)
	grpcServer := internalgrpc.NewServer()
	userv1.RegisterUserServiceServer(grpcServer, userGRPCServer)

	grpcListener, grpcListenError := (&net.ListenConfig{}).Listen(parentContext, "tcp", applicationConfiguration.GRPCListenAddress)
	if grpcListenError != nil {
		return nil, fmt.Errorf("grpc listen: %w", grpcListenError)
	}

	routerDependencies := router.Dependencies{
		UserRepository: userRepository, RoleRepository: roleRepository, PermissionRepository: permissionRepository,
		UserService: userService, AuthService: authService, AuthorizationService: authorizationService,
		AdminAuthHandler: adminAuthHandler, DashboardHandler: dashboardHandler, WelcomeHandler: welcomeHandler,
		Translator: translator, Cache: applicationCache,
		FileService: fileService, WebSocketHub: webSocketHub,
		HealthCheckers: []health.Checker{
			health.NewDatabaseChecker("postgres", databasePool.Ping),
			health.NewRedisChecker("redis", func(ctx context.Context) error { return redisClient.Ping(ctx).Err() }),
			health.NewNATSChecker("nats", natsHealthCheck(natsConnection)),
		},
	}
	return &applicationWiring{
		application: router.New(applicationConfiguration, routerDependencies), grpcServer: grpcServer,
		grpcListener: grpcListener, scheduler: scheduler, notificationConsumer: notificationConsumer,
		auditLogConsumer: auditLogConsumer, webSocketHub: webSocketHub,
	}, nil
}

func natsHealthCheck(natsConnection *natsgo.Conn) func(context.Context) error {
	return func(context.Context) error {
		if natsConnection.Status() != natsgo.CONNECTED {
			return fmt.Errorf("nats not connected: %s", natsConnection.Status().String())
		}
		return natsConnection.FlushTimeout(2 * time.Second)
	}
}

func registerWorkers(
	group *errgroup.Group,
	groupContext context.Context,
	applicationConfiguration *config.Config,
	application *fiber.App,
	grpcServer *grpc.Server,
	grpcListener net.Listener,
	scheduler *cron.Scheduler,
	notificationConsumer *consumers.NotificationConsumer,
	auditLogConsumer *consumers.AuditLogConsumer,
	webSocketHub *appwebsocket.Hub,
) {
	group.Go(func() error {
		scheduler.Start(groupContext)
		<-groupContext.Done()
		return nil
	})

	group.Go(func() error {
		slog.Info("http server starting", "address", applicationConfiguration.HTTPListenAddress)
		listenError := application.Listen(applicationConfiguration.HTTPListenAddress)
		if listenError != nil && !errors.Is(groupContext.Err(), context.Canceled) {
			return fmt.Errorf("http listen: %w", listenError)
		}
		return nil
	})

	group.Go(func() error {
		slog.Info("grpc server starting", "address", applicationConfiguration.GRPCListenAddress)
		serveError := grpcServer.Serve(grpcListener)
		if serveError != nil && !errors.Is(serveError, grpc.ErrServerStopped) && !errors.Is(groupContext.Err(), context.Canceled) {
			return fmt.Errorf("grpc serve: %w", serveError)
		}
		return nil
	})

	group.Go(func() error {
		if runError := notificationConsumer.Run(groupContext); runError != nil && !errors.Is(groupContext.Err(), context.Canceled) {
			return fmt.Errorf("notification consumer: %w", runError)
		}
		return nil
	})

	group.Go(func() error {
		if runError := auditLogConsumer.Run(groupContext); runError != nil && !errors.Is(groupContext.Err(), context.Canceled) {
			return fmt.Errorf("audit consumer: %w", runError)
		}
		return nil
	})

	group.Go(func() error {
		if subscribeError := webSocketHub.Subscribe(groupContext, appwebsocket.DefaultChannel); subscribeError != nil && !errors.Is(groupContext.Err(), context.Canceled) {
			return fmt.Errorf("websocket subscribe: %w", subscribeError)
		}
		return nil
	})

	group.Go(func() error {
		<-groupContext.Done()
		slog.Info("shutting down http and grpc servers")

		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if shutdownError := application.ShutdownWithContext(shutdownContext); shutdownError != nil {
			slog.Error("http server shutdown", "error", shutdownError)
		}
		grpcServer.GracefulStop()
		return nil
	})
}
