package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
)

func TestRoleRepository_CreateAndFind(testingContext *testing.T) {
	roleRepository, _, cleanup := newIntegrationRoleRepository(testingContext)
	defer cleanup()

	requestContext := context.Background()
	role := &domain.Role{Name: "auditor", Description: "Audit records"}
	if createError := roleRepository.Create(requestContext, role); createError != nil {
		testingContext.Fatalf("create role: %v", createError)
	}

	foundByID, findByIDError := roleRepository.FindByID(requestContext, role.ID)
	if findByIDError != nil {
		testingContext.Fatalf("find role by id: %v", findByIDError)
	}
	if foundByID.Name != role.Name {
		testingContext.Fatalf("expected role name %q, got %q", role.Name, foundByID.Name)
	}

	foundByName, findByNameError := roleRepository.FindByName(requestContext, role.Name)
	if findByNameError != nil {
		testingContext.Fatalf("find role by name: %v", findByNameError)
	}
	if foundByName.ID != role.ID {
		testingContext.Fatalf("expected role id %d, got %d", role.ID, foundByName.ID)
	}
}

func TestRoleRepository_AssignAndRemoveUserRole(testingContext *testing.T) {
	roleRepository, pool, cleanup := newIntegrationRoleRepository(testingContext)
	defer cleanup()

	requestContext := context.Background()
	userID := uuid.Must(uuid.NewV7())
	insertUserError := insertTestUser(requestContext, pool, userID, "role-user")
	if insertUserError != nil {
		testingContext.Fatalf("insert test user: %v", insertUserError)
	}

	adminRole, findRoleError := roleRepository.FindByName(requestContext, "admin")
	if findRoleError != nil {
		testingContext.Fatalf("find admin role: %v", findRoleError)
	}

	if assignError := roleRepository.AssignToUser(requestContext, userID, adminRole.ID); assignError != nil {
		testingContext.Fatalf("assign role to user: %v", assignError)
	}

	roles, findByUserIDError := roleRepository.FindByUserID(requestContext, userID)
	if findByUserIDError != nil {
		testingContext.Fatalf("find roles by user id: %v", findByUserIDError)
	}
	if len(roles) != 1 || roles[0].Name != "admin" {
		testingContext.Fatalf("expected one admin role assignment, got %+v", roles)
	}

	if removeError := roleRepository.RemoveFromUser(requestContext, userID, adminRole.ID); removeError != nil {
		testingContext.Fatalf("remove role from user: %v", removeError)
	}

	roles, findByUserIDError = roleRepository.FindByUserID(requestContext, userID)
	if findByUserIDError != nil {
		testingContext.Fatalf("find roles by user id after remove: %v", findByUserIDError)
	}
	if len(roles) != 0 {
		testingContext.Fatalf("expected no role assignments after remove, got %d", len(roles))
	}
}

func newIntegrationRoleRepository(testingContext *testing.T) (*roleRepository, *pgxpool.Pool, func()) {
	testingContext.Helper()

	requestContext, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	postgresContainer, runError := tcpostgres.Run(
		requestContext,
		"postgres:18.3-alpine",
		tcpostgres.WithDatabase("fiber_template"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
	)
	if runError != nil {
		cancel()
		testingContext.Skipf("skip integration test because postgres container is unavailable: %v", runError)
	}

	connectionString, connectionError := postgresContainer.ConnectionString(requestContext, "sslmode=disable")
	if connectionError != nil {
		cancel()
		_ = postgresContainer.Terminate(context.Background())
		testingContext.Fatalf("container connection string: %v", connectionError)
	}

	pool, newPoolError := pgxpool.New(requestContext, connectionString)
	if newPoolError != nil {
		cancel()
		_ = postgresContainer.Terminate(context.Background())
		testingContext.Fatalf("new pool: %v", newPoolError)
	}
	if waitError := waitForPoolReady(requestContext, pool); waitError != nil {
		pool.Close()
		cancel()
		_ = postgresContainer.Terminate(context.Background())
		testingContext.Fatalf("wait for postgres readiness: %v", waitError)
	}

	migrationFileNames := []string{
		"000001_create_users.up.sql",
		"000002_add_password_hash_to_users.up.sql",
		"000003_create_rbac_tables.up.sql",
		"000004_seed_rbac_data.up.sql",
	}
	if migrationError := executeMigrationFiles(requestContext, pool, migrationFileNames); migrationError != nil {
		pool.Close()
		cancel()
		_ = postgresContainer.Terminate(context.Background())
		testingContext.Fatalf("apply migrations: %v", migrationError)
	}

	cleanup := func() {
		pool.Close()
		cancel()
		_ = postgresContainer.Terminate(context.Background())
	}

	return &roleRepository{pool: pool}, pool, cleanup
}

func executeMigrationFiles(requestContext context.Context, pool *pgxpool.Pool, migrationFileNames []string) error {
	for _, migrationFileName := range migrationFileNames {
		migrationFilePath := filepath.Clean(filepath.Join("..", "..", "..", "migrations", migrationFileName))
		migrationSQL, readFileError := os.ReadFile(migrationFilePath)
		if readFileError != nil {
			return fmt.Errorf("read migration %s: %w", migrationFileName, readFileError)
		}
		if _, executeError := pool.Exec(requestContext, string(migrationSQL)); executeError != nil {
			return fmt.Errorf("execute migration %s: %w", migrationFileName, executeError)
		}
	}

	return nil
}

func insertTestUser(requestContext context.Context, pool *pgxpool.Pool, userID uuid.UUID, username string) error {
	const query = `
INSERT INTO users (id, username, password_hash, full_name, phone, status)
VALUES ($1, $2, $3, $4, $5, $6)`
	_, executeError := pool.Exec(
		requestContext,
		query,
		userID,
		username,
		"hashed-password",
		"Integration User",
		"+998900000001",
		domain.UserStatusActive,
	)
	if executeError != nil {
		return fmt.Errorf("insert test user: %w", executeError)
	}

	return nil
}

func waitForPoolReady(requestContext context.Context, pool *pgxpool.Pool) error {
	for retryIndex := 0; retryIndex < 20; retryIndex++ {
		pingError := pool.Ping(requestContext)
		if pingError == nil {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("postgres pool is not ready after retries")
}
