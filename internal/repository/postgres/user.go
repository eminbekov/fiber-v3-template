package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"github.com/eminbekov/fiber-v3-template/internal/repository"
)

type userRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a PostgreSQL-backed user repository.
func NewUserRepository(pool *pgxpool.Pool) repository.UserRepository {
	return &userRepository{
		pool: pool,
	}
}

func (repository *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const query = `SELECT id, username, password_hash, full_name, phone, status, created_at, updated_at, deleted_at
FROM users
WHERE id = $1 AND deleted_at IS NULL`

	rows, queryError := repository.pool.Query(ctx, query, id)
	if queryError != nil {
		return nil, fmt.Errorf("userRepository.FindByID query: %w", queryError)
	}
	defer rows.Close()

	user, collectError := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if collectError != nil {
		if errors.Is(collectError, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("userRepository.FindByID: %w", collectError)
	}

	return &user, nil
}

func (repository *userRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	const query = `SELECT id, username, password_hash, full_name, phone, status, created_at, updated_at, deleted_at
FROM users
WHERE username = $1 AND deleted_at IS NULL`

	rows, queryError := repository.pool.Query(ctx, query, username)
	if queryError != nil {
		return nil, fmt.Errorf("userRepository.FindByUsername query: %w", queryError)
	}
	defer rows.Close()

	user, collectError := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if collectError != nil {
		if errors.Is(collectError, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("userRepository.FindByUsername: %w", collectError)
	}

	return &user, nil
}

func (repository *userRepository) List(ctx context.Context, page int, pageSize int) ([]domain.User, int64, error) {
	const listQuery = `SELECT id, username, password_hash, full_name, phone, status, created_at, updated_at, deleted_at
FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2`
	const countQuery = `SELECT COUNT(1) FROM users WHERE deleted_at IS NULL`

	offset := (page - 1) * pageSize
	rows, queryError := repository.pool.Query(ctx, listQuery, pageSize, offset)
	if queryError != nil {
		return nil, 0, fmt.Errorf("userRepository.List query: %w", queryError)
	}
	defer rows.Close()

	users, collectRowsError := pgx.CollectRows(rows, pgx.RowToStructByName[domain.User])
	if collectRowsError != nil {
		return nil, 0, fmt.Errorf("userRepository.List collect rows: %w", collectRowsError)
	}

	var totalCount int64
	if scanError := repository.pool.QueryRow(ctx, countQuery).Scan(&totalCount); scanError != nil {
		return nil, 0, fmt.Errorf("userRepository.List count: %w", scanError)
	}

	return users, totalCount, nil
}

func (repository *userRepository) Create(ctx context.Context, user *domain.User) error {
	const query = `INSERT INTO users (id, username, password_hash, full_name, phone, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	now := time.Now().UTC()
	if user.ID == uuid.Nil {
		createdID, newIDError := uuid.NewV7()
		if newIDError != nil {
			return fmt.Errorf("userRepository.Create new v7 uuid: %w", newIDError)
		}
		user.ID = createdID
	}
	if user.Status == "" {
		user.Status = domain.UserStatusActive
	}
	user.CreatedAt = now
	user.UpdatedAt = now

	_, execError := repository.pool.Exec(
		ctx,
		query,
		user.ID,
		user.Username,
		user.PasswordHash,
		user.FullName,
		user.Phone,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if execError != nil {
		return fmt.Errorf("userRepository.Create: %w", execError)
	}

	return nil
}

func (repository *userRepository) Update(ctx context.Context, user *domain.User) error {
	const query = `UPDATE users
SET username = $2, password_hash = $3, full_name = $4, phone = $5, status = $6, updated_at = $7
WHERE id = $1 AND deleted_at IS NULL`

	user.UpdatedAt = time.Now().UTC()
	commandTag, execError := repository.pool.Exec(
		ctx,
		query,
		user.ID,
		user.Username,
		user.PasswordHash,
		user.FullName,
		user.Phone,
		user.Status,
		user.UpdatedAt,
	)
	if execError != nil {
		return fmt.Errorf("userRepository.Update: %w", execError)
	}
	if commandTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (repository *userRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	const query = `UPDATE users
SET deleted_at = now(), updated_at = now()
WHERE id = $1 AND deleted_at IS NULL`

	commandTag, execError := repository.pool.Exec(ctx, query, id)
	if execError != nil {
		return fmt.Errorf("userRepository.SoftDelete: %w", execError)
	}
	if commandTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
