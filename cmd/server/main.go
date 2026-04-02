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

	// [module:grpc:start]
	userv1 "github.com/eminbekov/fiber-v3-template/gen/proto/user/v1"
	// [module:grpc:end]
	"github.com/eminbekov/fiber-v3-template/internal/cache"
	"github.com/eminbekov/fiber-v3-template/internal/config"
	// [module:cron:start]
	"github.com/eminbekov/fiber-v3-template/internal/cron"
	// [module:cron:end]
	"github.com/eminbekov/fiber-v3-template/internal/database"
	// [module:grpc:start]
	internalgrpc "github.com/eminbekov/fiber-v3-template/internal/grpc"
	// [module:grpc:end]
	// [module:admin:start]
	"github.com/eminbekov/fiber-v3-template/internal/handler/admin"
	// [module:admin:end]
	// [module:web:start]
	"github.com/eminbekov/fiber-v3-template/internal/handler/web"
	// [module:web:end]
	// [module:i18n:start]
	"github.com/eminbekov/fiber-v3-template/internal/i18n"
	// [module:i18n:end]
	// [module:nats:start]
	appnats "github.com/eminbekov/fiber-v3-template/internal/nats"
	"github.com/eminbekov/fiber-v3-template/internal/nats/consumers"
	// [module:nats:end]
	"github.com/eminbekov/fiber-v3-template/internal/repository/postgres"
	"github.com/eminbekov/fiber-v3-template/internal/router"
	"github.com/eminbekov/fiber-v3-template/internal/service"
	"github.com/eminbekov/fiber-v3-template/internal/session"
	// [module:storage:start]
	"github.com/eminbekov/fiber-v3-template/internal/storage"
	// [module:storage:end]
	// [module:websocket:start]
	appwebsocket "github.com/eminbekov/fiber-v3-template/internal/websocket"
	// [module:websocket:end]
	"github.com/eminbekov/fiber-v3-template/package/hasher"
	"github.com/eminbekov/fiber-v3-template/package/health"
	"github.com/eminbekov/fiber-v3-template/package/logger"
	"github.com/eminbekov/fiber-v3-template/package/telemetry"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	// [module:nats:start]
	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	// [module:nats:end]
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	// [module:grpc:start]
	"google.golang.org/grpc"
	// [module:grpc:end]
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

//nolint:funlen // marker comments for removable modules increase function length.
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

	// [module:storage:start]
	fileStorage, fileStorageError := storage.NewFromApplicationConfig(parentContext, applicationConfiguration)
	if fileStorageError != nil {
		return fmt.Errorf("file storage: %w", fileStorageError)
	}
	// [module:storage:end]

	redisClient, redisClientError := cache.NewRedisClient(applicationConfiguration.RedisURL)
	if redisClientError != nil {
		return fmt.Errorf("redis: %w", redisClientError)
	}
	defer func() {
		if closeError := redisClient.Close(); closeError != nil {
			slog.Error("redis close", "error", closeError)
		}
	}()

	// [module:nats:start]
	natsConnection, jetStream, natsConnectError := appnats.Connect(parentContext, applicationConfiguration.NATSURL)
	if natsConnectError != nil {
		return fmt.Errorf("nats: %w", natsConnectError)
	}
	defer natsConnection.Close()
	// [module:nats:end]

	wiring, wiringError := wireApplication(
		parentContext,
		applicationConfiguration,
		databasePool,
		redisClient,
		// [module:nats:start]
		natsConnection,
		jetStream,
		// [module:nats:end]
		// [module:storage:start]
		fileStorage,
		// [module:storage:end]
	)
	if wiringError != nil {
		return wiringError
	}
	// [module:grpc:start]
	defer func() {
		if closeError := wiring.grpcListener.Close(); closeError != nil {
			slog.Error("grpc listener close", "error", closeError)
		}
	}()
	// [module:grpc:end]

	group, groupContext := errgroup.WithContext(parentContext)
	registerWorkers(
		group,
		groupContext,
		applicationConfiguration,
		wiring.application,
		// [module:grpc:start]
		wiring.grpcServer,
		wiring.grpcListener,
		// [module:grpc:end]
		// [module:cron:start]
		wiring.scheduler,
		// [module:cron:end]
		// [module:nats:start]
		wiring.notificationConsumer,
		wiring.auditLogConsumer,
		// [module:nats:end]
		// [module:websocket:start]
		wiring.webSocketHub,
		// [module:websocket:end]
	)

	if waitError := group.Wait(); waitError != nil && !errors.Is(waitError, context.Canceled) {
		return fmt.Errorf("run: %w", waitError)
	}

	return nil
}

type applicationWiring struct {
	application *fiber.App
	// [module:grpc:start]
	grpcServer   *grpc.Server
	grpcListener net.Listener
	// [module:grpc:end]
	// [module:cron:start]
	scheduler *cron.Scheduler
	// [module:cron:end]
	// [module:nats:start]
	notificationConsumer *consumers.NotificationConsumer
	auditLogConsumer     *consumers.AuditLogConsumer
	// [module:nats:end]
	// [module:websocket:start]
	webSocketHub *appwebsocket.Hub
	// [module:websocket:end]
}

//nolint:funlen // marker comments for removable modules increase function length.
func wireApplication(
	parentContext context.Context,
	applicationConfiguration *config.Config,
	databasePool *pgxpool.Pool,
	redisClient *redis.Client,
	// [module:nats:start]
	natsConnection *natsgo.Conn,
	jetStream jetstream.JetStream,
	// [module:nats:end]
	// [module:storage:start]
	fileStorage storage.FileStorage,
	// [module:storage:end]
) (*applicationWiring, error) {
	// [module:nats:start]
	notificationConsumer := consumers.NewNotificationConsumer(jetStream)
	auditLogConsumer := consumers.NewAuditLogConsumer(jetStream)
	// [module:nats:end]

	userRepository := postgres.NewUserRepository(databasePool)
	roleRepository := postgres.NewRoleRepository(databasePool)
	permissionRepository := postgres.NewPermissionRepository(databasePool)
	applicationCache := cache.NewRedisCache(redisClient)
	// [module:websocket:start]
	webSocketHub := appwebsocket.NewHub(redisClient)
	// [module:websocket:end]
	passwordHasher := hasher.NewArgon2ID()
	sessionStore := session.NewRedisStore(redisClient, applicationConfiguration.SessionDuration)
	// [module:storage:start]
	fileService := service.NewFileService(fileStorage, applicationConfiguration.SignedURLTTL)
	// [module:storage:end]
	userService := service.NewUserService(userRepository, roleRepository, applicationCache, passwordHasher)
	authService := service.NewAuthService(
		userRepository,
		sessionStore,
		passwordHasher,
		applicationConfiguration.SessionDuration,
	)
	authorizationService := service.NewAuthorizationService(permissionRepository, applicationCache)
	// [module:cron:start]
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
	// [module:cron:end]

	// [module:i18n:start]
	translator, translatorError := i18n.NewTranslator("en")
	if translatorError != nil {
		return nil, fmt.Errorf("translator: %w", translatorError)
	}
	// [module:i18n:end]
	// [module:admin:start]
	dashboardHandler := admin.NewDashboardHandler(translator)
	// [module:admin:end]
	// [module:web:start]
	welcomeHandler := web.NewWelcomeHandler(translator)
	// [module:web:end]
	// [module:admin:start]
	secureSessionCookie := applicationConfiguration.Environment == "production"
	adminAuthHandler := admin.NewAdminAuthHandler(
		authService,
		translator,
		applicationConfiguration.SessionDuration,
		secureSessionCookie,
	)
	// [module:admin:end]
	// [module:grpc:start]
	userGRPCServer := internalgrpc.NewUserServer(userService)
	grpcServer := internalgrpc.NewServer()
	userv1.RegisterUserServiceServer(grpcServer, userGRPCServer)

	grpcListener, grpcListenError := (&net.ListenConfig{}).Listen(parentContext, "tcp", applicationConfiguration.GRPCListenAddress)
	if grpcListenError != nil {
		return nil, fmt.Errorf("grpc listen: %w", grpcListenError)
	}
	// [module:grpc:end]

	routerDependencies := router.Dependencies{
		UserRepository: userRepository, RoleRepository: roleRepository, PermissionRepository: permissionRepository,
		UserService: userService, AuthService: authService, AuthorizationService: authorizationService,
		// [module:admin:start]
		AdminAuthHandler: adminAuthHandler, DashboardHandler: dashboardHandler,
		// [module:admin:end]
		// [module:web:start]
		WelcomeHandler: welcomeHandler,
		// [module:web:end]
		// [module:i18n:start]
		Translator: translator,
		// [module:i18n:end]
		Cache: applicationCache,
		// [module:storage:start]
		FileService: fileService,
		// [module:storage:end]
		// [module:websocket:start]
		WebSocketHub: webSocketHub,
		// [module:websocket:end]
		HealthCheckers: []health.Checker{
			health.NewDatabaseChecker("postgres", databasePool.Ping),
			health.NewRedisChecker("redis", func(ctx context.Context) error { return redisClient.Ping(ctx).Err() }),
			// [module:nats:start]
			health.NewNATSChecker("nats", natsHealthCheck(natsConnection)),
			// [module:nats:end]
		},
	}
	return &applicationWiring{
		application: router.New(applicationConfiguration, routerDependencies),
		// [module:grpc:start]
		grpcServer: grpcServer, grpcListener: grpcListener,
		// [module:grpc:end]
		// [module:cron:start]
		scheduler: scheduler,
		// [module:cron:end]
		// [module:nats:start]
		notificationConsumer: notificationConsumer, auditLogConsumer: auditLogConsumer,
		// [module:nats:end]
		// [module:websocket:start]
		webSocketHub: webSocketHub,
		// [module:websocket:end]
	}, nil
}

// [module:nats:start]
func natsHealthCheck(natsConnection *natsgo.Conn) func(context.Context) error {
	return func(context.Context) error {
		if natsConnection.Status() != natsgo.CONNECTED {
			return fmt.Errorf("nats not connected: %s", natsConnection.Status().String())
		}
		return natsConnection.FlushTimeout(2 * time.Second)
	}
}

// [module:nats:end]

func registerWorkers(
	group *errgroup.Group,
	groupContext context.Context,
	applicationConfiguration *config.Config,
	application *fiber.App,
	// [module:grpc:start]
	grpcServer *grpc.Server,
	grpcListener net.Listener,
	// [module:grpc:end]
	// [module:cron:start]
	scheduler *cron.Scheduler,
	// [module:cron:end]
	// [module:nats:start]
	notificationConsumer *consumers.NotificationConsumer,
	auditLogConsumer *consumers.AuditLogConsumer,
	// [module:nats:end]
	// [module:websocket:start]
	webSocketHub *appwebsocket.Hub,
	// [module:websocket:end]
) {
	// [module:cron:start]
	group.Go(func() error {
		scheduler.Start(groupContext)
		<-groupContext.Done()
		return nil
	})
	// [module:cron:end]

	group.Go(func() error {
		slog.Info("http server starting", "address", applicationConfiguration.HTTPListenAddress)
		listenError := application.Listen(applicationConfiguration.HTTPListenAddress)
		if listenError != nil && !errors.Is(groupContext.Err(), context.Canceled) {
			return fmt.Errorf("http listen: %w", listenError)
		}
		return nil
	})

	// [module:grpc:start]
	group.Go(func() error {
		slog.Info("grpc server starting", "address", applicationConfiguration.GRPCListenAddress)
		serveError := grpcServer.Serve(grpcListener)
		if serveError != nil && !errors.Is(serveError, grpc.ErrServerStopped) && !errors.Is(groupContext.Err(), context.Canceled) {
			return fmt.Errorf("grpc serve: %w", serveError)
		}
		return nil
	})
	// [module:grpc:end]

	// [module:nats:start]
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
	// [module:nats:end]

	// [module:websocket:start]
	group.Go(func() error {
		if subscribeError := webSocketHub.Subscribe(groupContext, appwebsocket.DefaultChannel); subscribeError != nil && !errors.Is(groupContext.Err(), context.Canceled) {
			return fmt.Errorf("websocket subscribe: %w", subscribeError)
		}
		return nil
	})
	// [module:websocket:end]

	group.Go(func() error {
		<-groupContext.Done()
		slog.Info("shutting down http and grpc servers")

		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if shutdownError := application.ShutdownWithContext(shutdownContext); shutdownError != nil {
			slog.Error("http server shutdown", "error", shutdownError)
		}
		// [module:grpc:start]
		grpcServer.GracefulStop()
		// [module:grpc:end]
		return nil
	})
}
