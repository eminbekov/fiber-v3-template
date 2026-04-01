package postgres

import (
	"context"
	"fmt"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"github.com/eminbekov/fiber-v3-template/internal/repository"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type permissionRepository struct {
	pool *pgxpool.Pool
}

func NewPermissionRepository(pool *pgxpool.Pool) repository.PermissionRepository {
	return &permissionRepository{
		pool: pool,
	}
}

func (repository *permissionRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Permission, error) {
	const query = `
SELECT DISTINCT permissions.id, permissions.resource, permissions.action
FROM permissions
JOIN role_permissions ON role_permissions.permission_id = permissions.id
JOIN user_roles ON user_roles.role_id = role_permissions.role_id
WHERE user_roles.user_id = $1
ORDER BY permissions.resource, permissions.action`

	rows, queryError := repository.pool.Query(ctx, query, userID)
	if queryError != nil {
		return nil, fmt.Errorf("permissionRepository.FindByUserID query: %w", queryError)
	}
	defer rows.Close()

	permissions, collectError := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Permission])
	if collectError != nil {
		return nil, fmt.Errorf("permissionRepository.FindByUserID: %w", collectError)
	}

	return permissions, nil
}

func (repository *permissionRepository) FindByRoleID(ctx context.Context, roleID int64) ([]domain.Permission, error) {
	const query = `
SELECT permissions.id, permissions.resource, permissions.action
FROM permissions
JOIN role_permissions ON role_permissions.permission_id = permissions.id
WHERE role_permissions.role_id = $1
ORDER BY permissions.resource, permissions.action`

	rows, queryError := repository.pool.Query(ctx, query, roleID)
	if queryError != nil {
		return nil, fmt.Errorf("permissionRepository.FindByRoleID query: %w", queryError)
	}
	defer rows.Close()

	permissions, collectError := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Permission])
	if collectError != nil {
		return nil, fmt.Errorf("permissionRepository.FindByRoleID: %w", collectError)
	}

	return permissions, nil
}

func (repository *permissionRepository) List(ctx context.Context) ([]domain.Permission, error) {
	const query = `SELECT id, resource, action FROM permissions ORDER BY resource, action`
	rows, queryError := repository.pool.Query(ctx, query)
	if queryError != nil {
		return nil, fmt.Errorf("permissionRepository.List query: %w", queryError)
	}
	defer rows.Close()

	permissions, collectError := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Permission])
	if collectError != nil {
		return nil, fmt.Errorf("permissionRepository.List: %w", collectError)
	}

	return permissions, nil
}
