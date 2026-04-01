package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestUserRepository_CreateAndFindByID(testingContext *testing.T) {
	testingContext.Parallel()

	userRepository := newIntegrationUserRepository(testingContext)
	requestContext := context.Background()
	user := &domain.User{
		Username: "create-find-id",
		FullName: "Create Find User",
		Phone:    "+998901111111",
		Status:   domain.UserStatusActive,
	}

	createError := userRepository.Create(requestContext, user)
	if createError != nil {
		testingContext.Fatalf("create user: %v", createError)
	}

	foundUser, findByIDError := userRepository.FindByID(requestContext, user.ID)
	if findByIDError != nil {
		testingContext.Fatalf("find user by id: %v", findByIDError)
	}
	if foundUser.Username != user.Username || foundUser.FullName != user.FullName || foundUser.Phone != user.Phone {
		testingContext.Fatalf("unexpected user values: %+v", foundUser)
	}
}

func TestUserRepository_FindByUsername(testingContext *testing.T) {
	testingContext.Parallel()

	userRepository := newIntegrationUserRepository(testingContext)
	requestContext := context.Background()
	user := &domain.User{
		Username: "find-by-username",
		FullName: "Find By Username",
		Phone:    "+998902222222",
		Status:   domain.UserStatusActive,
	}

	createError := userRepository.Create(requestContext, user)
	if createError != nil {
		testingContext.Fatalf("create user: %v", createError)
	}

	foundUser, findByUsernameError := userRepository.FindByUsername(requestContext, user.Username)
	if findByUsernameError != nil {
		testingContext.Fatalf("find user by username: %v", findByUsernameError)
	}
	if foundUser.ID != user.ID {
		testingContext.Fatalf("expected id %s, got %s", user.ID.String(), foundUser.ID.String())
	}
}

func TestUserRepository_List(testingContext *testing.T) {
	testingContext.Parallel()

	userRepository := newIntegrationUserRepository(testingContext)
	requestContext := context.Background()

	usersToCreate := []*domain.User{
		{Username: "list-user-1", FullName: "List User One", Phone: "+998903333331", Status: domain.UserStatusActive},
		{Username: "list-user-2", FullName: "List User Two", Phone: "+998903333332", Status: domain.UserStatusActive},
		{Username: "list-user-3", FullName: "List User Three", Phone: "+998903333333", Status: domain.UserStatusActive},
	}
	for _, user := range usersToCreate {
		if createError := userRepository.Create(requestContext, user); createError != nil {
			testingContext.Fatalf("create user %s: %v", user.Username, createError)
		}
	}

	users, totalCount, listError := userRepository.List(requestContext, 1, 2)
	if listError != nil {
		testingContext.Fatalf("list users: %v", listError)
	}
	if len(users) != 2 {
		testingContext.Fatalf("expected 2 users in page, got %d", len(users))
	}
	if totalCount != 3 {
		testingContext.Fatalf("expected total count 3, got %d", totalCount)
	}
}

func TestUserRepository_Update(testingContext *testing.T) {
	testingContext.Parallel()

	userRepository := newIntegrationUserRepository(testingContext)
	requestContext := context.Background()
	user := &domain.User{
		Username: "before-update",
		FullName: "Before Update",
		Phone:    "+998904444444",
		Status:   domain.UserStatusActive,
	}
	if createError := userRepository.Create(requestContext, user); createError != nil {
		testingContext.Fatalf("create user: %v", createError)
	}

	user.FullName = "After Update"
	user.Phone = "+998905555555"
	updateError := userRepository.Update(requestContext, user)
	if updateError != nil {
		testingContext.Fatalf("update user: %v", updateError)
	}

	updatedUser, findByIDError := userRepository.FindByID(requestContext, user.ID)
	if findByIDError != nil {
		testingContext.Fatalf("find user by id: %v", findByIDError)
	}
	if updatedUser.FullName != "After Update" || updatedUser.Phone != "+998905555555" {
		testingContext.Fatalf("unexpected updated user: %+v", updatedUser)
	}
}

func TestUserRepository_SoftDelete(testingContext *testing.T) {
	testingContext.Parallel()

	userRepository := newIntegrationUserRepository(testingContext)
	requestContext := context.Background()
	user := &domain.User{
		Username: "soft-delete-user",
		FullName: "Soft Delete User",
		Phone:    "+998906666666",
		Status:   domain.UserStatusActive,
	}
	if createError := userRepository.Create(requestContext, user); createError != nil {
		testingContext.Fatalf("create user: %v", createError)
	}

	softDeleteError := userRepository.SoftDelete(requestContext, user.ID)
	if softDeleteError != nil {
		testingContext.Fatalf("soft delete user: %v", softDeleteError)
	}

	_, findByIDError := userRepository.FindByID(requestContext, user.ID)
	if !errors.Is(findByIDError, domain.ErrNotFound) {
		testingContext.Fatalf("expected domain.ErrNotFound, got %v", findByIDError)
	}
}

func newIntegrationUserRepository(testingContext *testing.T) *userRepository {
	testingContext.Helper()

	if testing.Short() {
		testingContext.Skip("skipping integration test in short mode")
	}

	requestContext, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	testingContext.Cleanup(cancel)

	postgresContainer, runError := runPostgresContainer(requestContext)
	if runError != nil {
		testingContext.Skipf("skipping integration test because postgres container is unavailable: %v", runError)
	}
	testingContext.Cleanup(func() {
		_ = postgresContainer.Terminate(context.Background())
	})

	connectionString, connectionError := postgresContainer.ConnectionString(requestContext, "sslmode=disable")
	if connectionError != nil {
		testingContext.Fatalf("container connection string: %v", connectionError)
	}

	pool, newPoolError := pgxpool.New(requestContext, connectionString)
	if newPoolError != nil {
		testingContext.Fatalf("pgxpool new: %v", newPoolError)
	}
	testingContext.Cleanup(pool.Close)

	migrationFilePath := filepath.Clean(filepath.Join("..", "..", "..", "migrations", "000001_create_users.up.sql"))
	migrationSQL, readFileError := os.ReadFile(migrationFilePath)
	if readFileError != nil {
		testingContext.Fatalf("read migration file: %v", readFileError)
	}
	if _, executeMigrationError := pool.Exec(requestContext, string(migrationSQL)); executeMigrationError != nil {
		testingContext.Fatalf("execute migration: %v", executeMigrationError)
	}

	return &userRepository{pool: pool}
}

func runPostgresContainer(requestContext context.Context) (postgresContainer *tcpostgres.PostgresContainer, runError error) {
	defer func() {
		if recoveredPanic := recover(); recoveredPanic != nil {
			runError = fmt.Errorf("panic while starting postgres container: %v", recoveredPanic)
		}
	}()

	return tcpostgres.Run(
		requestContext,
		"postgres:18.3-alpine",
		tcpostgres.WithDatabase("fiber_template"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
	)
}
