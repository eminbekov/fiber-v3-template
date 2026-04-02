package main

import (
	"context"
	"fmt"
	"os"

	"github.com/eminbekov/fiber-v3-template/internal/cache"
	"github.com/eminbekov/fiber-v3-template/internal/config"
	"github.com/eminbekov/fiber-v3-template/internal/console/commands"
	"github.com/eminbekov/fiber-v3-template/internal/database"
	"github.com/eminbekov/fiber-v3-template/internal/repository/postgres"
	"github.com/eminbekov/fiber-v3-template/internal/service"
	"github.com/eminbekov/fiber-v3-template/package/hasher"
	"github.com/eminbekov/fiber-v3-template/package/logger"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	if runError := runConsole(context.Background(), os.Args[1], os.Args[2:]); runError != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", runError)
		os.Exit(1)
	}
}

func runConsole(ctx context.Context, commandName string, arguments []string) error {
	applicationConfiguration, configurationError := config.Load()
	if configurationError != nil {
		return fmt.Errorf("config: %w", configurationError)
	}

	logger.Setup(applicationConfiguration.LogLevel, applicationConfiguration.Environment)

	databasePool, databasePoolError := database.NewPool(ctx, applicationConfiguration.DatabaseURL)
	if databasePoolError != nil {
		return fmt.Errorf("database pool: %w", databasePoolError)
	}
	defer databasePool.Close()

	redisClient, redisClientError := cache.NewRedisClient(applicationConfiguration.RedisURL)
	if redisClientError != nil {
		return fmt.Errorf("redis: %w", redisClientError)
	}
	defer func() {
		_ = redisClient.Close()
	}()

	userRepository := postgres.NewUserRepository(databasePool)
	roleRepository := postgres.NewRoleRepository(databasePool)
	applicationCache := cache.NewRedisCache(redisClient)
	passwordHasher := hasher.NewArgon2ID()
	userService := service.NewUserService(userRepository, roleRepository, applicationCache, passwordHasher)
	dependencies := &commands.Dependencies{
		UserService:    userService,
		RoleRepository: roleRepository,
		Cache:          applicationCache,
		RedisClient:    redisClient,
	}

	switch commandName {
	case "create-admin":
		if createAdminError := commands.CreateAdmin(ctx, dependencies, arguments); createAdminError != nil {
			return fmt.Errorf("console create-admin: %w", createAdminError)
		}
		return nil
	case "assign-role":
		if assignRoleError := commands.AssignRole(ctx, dependencies, arguments); assignRoleError != nil {
			return fmt.Errorf("console assign-role: %w", assignRoleError)
		}
		return nil
	case "cache-clear":
		if cacheClearError := commands.CacheClear(ctx, dependencies, arguments); cacheClearError != nil {
			return fmt.Errorf("console cache-clear: %w", cacheClearError)
		}
		return nil
	case "export-users":
		if exportUsersError := commands.ExportUsers(ctx, dependencies, arguments); exportUsersError != nil {
			return fmt.Errorf("console export-users: %w", exportUsersError)
		}
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command: %s", commandName)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/console <command> [options]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  create-admin   Create an admin user")
	fmt.Fprintln(os.Stderr, "  assign-role    Assign role to user")
	fmt.Fprintln(os.Stderr, "  cache-clear    Clear redis cache (optionally by prefix)")
	fmt.Fprintln(os.Stderr, "  export-users   Export users to CSV")
}
