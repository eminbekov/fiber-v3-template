package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/eminbekov/fiber-v3-template/internal/cache"
	"github.com/eminbekov/fiber-v3-template/internal/repository"
	"github.com/gofrs/uuid/v5"
)

const permissionsCacheTTL = 10 * time.Minute

type AuthorizationService struct {
	permissionRepository repository.PermissionRepository
	cache                cache.Cache
}

func NewAuthorizationService(
	permissionRepository repository.PermissionRepository,
	applicationCache cache.Cache,
) *AuthorizationService {
	return &AuthorizationService{
		permissionRepository: permissionRepository,
		cache:                applicationCache,
	}
}

func (service *AuthorizationService) HasPermission(
	ctx context.Context,
	userID uuid.UUID,
	resource string,
	action string,
) (bool, error) {
	cacheKey := cache.PermissionsByUserIDKey(userID)
	targetPermission := resource + "." + action

	permissionKeys := make([]string, 0, 16)
	if service.cache != nil {
		cacheError := service.cache.Get(ctx, cacheKey, &permissionKeys)
		if cacheError == nil {
			return permissionListContains(permissionKeys, targetPermission, resource), nil
		}
		if !errors.Is(cacheError, cache.ErrCacheMiss) {
			return false, fmt.Errorf("authorizationService.HasPermission cache get: %w", cacheError)
		}
	}

	permissions, findError := service.permissionRepository.FindByUserID(ctx, userID)
	if findError != nil {
		return false, fmt.Errorf("authorizationService.HasPermission find by user id: %w", findError)
	}

	for _, permission := range permissions {
		permissionKeys = append(permissionKeys, permission.String())
	}

	if service.cache != nil {
		_ = service.cache.Set(ctx, cacheKey, permissionKeys, permissionsCacheTTL)
	}

	return permissionListContains(permissionKeys, targetPermission, resource), nil
}

func permissionListContains(permissionKeys []string, targetPermission string, resource string) bool {
	for _, permissionKey := range permissionKeys {
		if permissionKey == targetPermission || permissionKey == resource+".*" {
			return true
		}
	}

	return false
}
