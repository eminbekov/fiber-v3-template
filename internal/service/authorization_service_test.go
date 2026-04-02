package service

import (
	"context"
	"testing"
	"time"

	"github.com/eminbekov/fiber-v3-template/internal/cache"
	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"github.com/gofrs/uuid/v5"
)

func TestAuthorizationService_HasPermission(testingContext *testing.T) {
	testingContext.Parallel()

	userID := uuid.Must(uuid.NewV7())

	testingContext.Run("returns true from cache hit", func(testingContext *testing.T) {
		testingContext.Parallel()

		service := NewAuthorizationService(
			&mockPermissionRepository{
				findByUserIDFunction: func(ctx context.Context, userID uuid.UUID) ([]domain.Permission, error) {
					testingContext.Fatalf("repository should not be called on cache hit")
					return nil, nil
				},
				findByRoleIDFunction: func(ctx context.Context, roleID int64) ([]domain.Permission, error) { return nil, nil },
				listFunction:         func(ctx context.Context) ([]domain.Permission, error) { return nil, nil },
			},
			&mockCache{
				getFunction: func(ctx context.Context, key string, destination any) error {
					keys, ok := destination.(*[]string)
					if !ok {
						testingContext.Fatalf("destination must be *[]string")
					}
					*keys = []string{"user.read"}
					return nil
				},
				setFunction:            func(ctx context.Context, key string, value any, ttlDuration time.Duration) error { return nil },
				deleteFunction:         func(ctx context.Context, keys ...string) error { return nil },
				deleteByPrefixFunction: func(ctx context.Context, prefix string) error { return nil },
			},
		)

		hasPermission, permissionError := service.HasPermission(context.Background(), userID, "user", "read")
		if permissionError != nil {
			testingContext.Fatalf("expected no error, got %v", permissionError)
		}
		if !hasPermission {
			testingContext.Fatalf("expected permission grant")
		}
	})

	testingContext.Run("returns wildcard permission from repository", func(testingContext *testing.T) {
		testingContext.Parallel()

		service := NewAuthorizationService(
			&mockPermissionRepository{
				findByUserIDFunction: func(ctx context.Context, userID uuid.UUID) ([]domain.Permission, error) {
					return []domain.Permission{{Resource: "admin", Action: "*"}}, nil
				},
				findByRoleIDFunction: func(ctx context.Context, roleID int64) ([]domain.Permission, error) { return nil, nil },
				listFunction:         func(ctx context.Context) ([]domain.Permission, error) { return nil, nil },
			},
			&mockCache{
				getFunction:            func(ctx context.Context, key string, destination any) error { return cache.ErrCacheMiss },
				setFunction:            func(ctx context.Context, key string, value any, ttlDuration time.Duration) error { return nil },
				deleteFunction:         func(ctx context.Context, keys ...string) error { return nil },
				deleteByPrefixFunction: func(ctx context.Context, prefix string) error { return nil },
			},
		)

		hasPermission, permissionError := service.HasPermission(context.Background(), userID, "admin", "delete")
		if permissionError != nil {
			testingContext.Fatalf("expected no error, got %v", permissionError)
		}
		if !hasPermission {
			testingContext.Fatalf("expected wildcard permission grant")
		}
	})
}
