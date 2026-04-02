package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/eminbekov/fiber-v3-template/internal/database"
	"github.com/eminbekov/fiber-v3-template/package/hasher"
)

const databaseURLVariableName = "DATABASE_URL"

func main() {
	if runError := run(); runError != nil {
		log.Fatalf("seed: %v", runError)
	}
}

func run() error {
	ctx := context.Background()

	databaseURL := os.Getenv(databaseURLVariableName)
	if databaseURL == "" {
		return fmt.Errorf("%s cannot be empty", databaseURLVariableName)
	}

	pool, poolError := database.NewPool(ctx, databaseURL)
	if poolError != nil {
		return fmt.Errorf("database pool: %w", poolError)
	}
	defer pool.Close()

	if seedError := seedDevelopmentData(ctx, pool); seedError != nil {
		return fmt.Errorf("seed development data: %w", seedError)
	}

	fmt.Println("seed: development data seeded successfully")
	return nil
}

type seedablePool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func seedDevelopmentData(ctx context.Context, pool seedablePool) error {
	passwordHash, hashError := hasher.Hash("password123")
	if hashError != nil {
		return fmt.Errorf("hash password: %w", hashError)
	}

	_, insertError := pool.Exec(ctx, `
		INSERT INTO users (id, username, password_hash, full_name, phone, status)
		VALUES (gen_random_uuid(), 'testuser', $1, 'Test User', '+998901234567', 'active')
		ON CONFLICT (username) DO NOTHING
	`, passwordHash)
	if insertError != nil {
		return fmt.Errorf("insert test user: %w", insertError)
	}

	adminPasswordHash, adminHashError := hasher.Hash("admin123")
	if adminHashError != nil {
		return fmt.Errorf("hash admin password: %w", adminHashError)
	}

	_, adminInsertError := pool.Exec(ctx, `
		INSERT INTO users (id, username, password_hash, full_name, phone, status)
		VALUES (gen_random_uuid(), 'admin', $1, 'Administrator', '+998900000000', 'active')
		ON CONFLICT (username) DO NOTHING
	`, adminPasswordHash)
	if adminInsertError != nil {
		return fmt.Errorf("insert admin user: %w", adminInsertError)
	}

	_, roleAssignError := pool.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		SELECT users.id, roles.id
		FROM users, roles
		WHERE users.username = 'admin' AND roles.name = 'admin'
		ON CONFLICT DO NOTHING
	`)
	if roleAssignError != nil {
		return fmt.Errorf("assign admin role: %w", roleAssignError)
	}

	return nil
}
