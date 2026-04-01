package repository

import (
	"context"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"github.com/gofrs/uuid/v5"
)

type RoleRepository interface {
	FindByID(ctx context.Context, id int64) (*domain.Role, error)
	FindByName(ctx context.Context, name string) (*domain.Role, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Role, error)
	List(ctx context.Context) ([]domain.Role, error)
	Create(ctx context.Context, role *domain.Role) error
	AssignToUser(ctx context.Context, userID uuid.UUID, roleID int64) error
	RemoveFromUser(ctx context.Context, userID uuid.UUID, roleID int64) error
}
