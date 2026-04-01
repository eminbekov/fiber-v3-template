package repository

import (
	"context"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"github.com/gofrs/uuid/v5"
)

type PermissionRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Permission, error)
	FindByRoleID(ctx context.Context, roleID int64) ([]domain.Permission, error)
	List(ctx context.Context) ([]domain.Permission, error)
}
