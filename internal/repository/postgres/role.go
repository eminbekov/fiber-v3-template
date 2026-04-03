package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"github.com/eminbekov/fiber-v3-template/internal/repository"
)

type roleRepository struct {
	pool *pgxpool.Pool
}

func NewRoleRepository(pool *pgxpool.Pool) repository.RoleRepository {
	return &roleRepository{
		pool: pool,
	}
}

func (repository *roleRepository) FindByID(ctx context.Context, id int64) (*domain.Role, error) {
	const query = `SELECT id, name, description FROM roles WHERE id = $1`
	rows, queryError := repository.pool.Query(ctx, query, id)
	if queryError != nil {
		return nil, fmt.Errorf("roleRepository.FindByID query: %w", queryError)
	}
	defer rows.Close()

	role, collectError := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Role])
	if collectError != nil {
		if errors.Is(collectError, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("roleRepository.FindByID: %w", collectError)
	}

	return &role, nil
}

func (repository *roleRepository) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	const query = `SELECT id, name, description FROM roles WHERE name = $1`
	rows, queryError := repository.pool.Query(ctx, query, name)
	if queryError != nil {
		return nil, fmt.Errorf("roleRepository.FindByName query: %w", queryError)
	}
	defer rows.Close()

	role, collectError := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Role])
	if collectError != nil {
		if errors.Is(collectError, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("roleRepository.FindByName: %w", collectError)
	}

	return &role, nil
}

func (repository *roleRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Role, error) {
	const query = `
SELECT roles.id, roles.name, roles.description
FROM roles
JOIN user_roles ON user_roles.role_id = roles.id
WHERE user_roles.user_id = $1
ORDER BY roles.name`

	rows, queryError := repository.pool.Query(ctx, query, userID)
	if queryError != nil {
		return nil, fmt.Errorf("roleRepository.FindByUserID query: %w", queryError)
	}
	defer rows.Close()

	roles, collectError := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Role])
	if collectError != nil {
		return nil, fmt.Errorf("roleRepository.FindByUserID: %w", collectError)
	}

	return roles, nil
}

func (repository *roleRepository) List(ctx context.Context) ([]domain.Role, error) {
	const query = `SELECT id, name, description FROM roles ORDER BY name`
	rows, queryError := repository.pool.Query(ctx, query)
	if queryError != nil {
		return nil, fmt.Errorf("roleRepository.List query: %w", queryError)
	}
	defer rows.Close()

	roles, collectError := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Role])
	if collectError != nil {
		return nil, fmt.Errorf("roleRepository.List: %w", collectError)
	}

	return roles, nil
}

func (repository *roleRepository) Create(ctx context.Context, role *domain.Role) error {
	const query = `INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id`
	if scanError := repository.pool.QueryRow(ctx, query, role.Name, role.Description).Scan(&role.ID); scanError != nil {
		return fmt.Errorf("roleRepository.Create: %w", scanError)
	}

	return nil
}

func (repository *roleRepository) AssignToUser(ctx context.Context, userID uuid.UUID, roleID int64) error {
	const query = `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	if _, execError := repository.pool.Exec(ctx, query, userID, roleID); execError != nil {
		return fmt.Errorf("roleRepository.AssignToUser: %w", execError)
	}

	return nil
}

func (repository *roleRepository) RemoveFromUser(ctx context.Context, userID uuid.UUID, roleID int64) error {
	const query = `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	if _, execError := repository.pool.Exec(ctx, query, userID, roleID); execError != nil {
		return fmt.Errorf("roleRepository.RemoveFromUser: %w", execError)
	}

	return nil
}
