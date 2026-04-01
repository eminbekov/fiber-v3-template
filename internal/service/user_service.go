package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eminbekov/fiber-v3-template/internal/cache"
	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"github.com/eminbekov/fiber-v3-template/internal/repository"
	"github.com/gofrs/uuid/v5"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
	userCacheTTL    = 5 * time.Minute
)

type UserService struct {
	userRepository repository.UserRepository
	cache          cache.Cache
	passwordHasher PasswordHasher
}

func NewUserService(userRepository repository.UserRepository, cache cache.Cache, passwordHasher PasswordHasher) *UserService {
	return &UserService{
		userRepository: userRepository,
		cache:          cache,
		passwordHasher: passwordHasher,
	}
}

func (service *UserService) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if id == uuid.Nil {
		return nil, domain.ErrValidation
	}

	cacheKey := cache.UserByIDKey(id)
	var cachedUser domain.User
	if service.cache != nil {
		if cacheError := service.cache.Get(ctx, cacheKey, &cachedUser); cacheError == nil {
			return &cachedUser, nil
		}
	}

	user, findByIDError := service.userRepository.FindByID(ctx, id)
	if findByIDError != nil {
		return nil, fmt.Errorf("UserService.FindByID: %w", findByIDError)
	}

	if service.cache != nil {
		_ = service.cache.Set(ctx, cacheKey, user, userCacheTTL)
	}

	return user, nil
}

func (service *UserService) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	normalizedUsername := strings.TrimSpace(username)
	if normalizedUsername == "" {
		return nil, domain.ErrValidation
	}

	cacheKey := cache.UserByUsernameKey(normalizedUsername)
	var cachedUser domain.User
	if service.cache != nil {
		if cacheError := service.cache.Get(ctx, cacheKey, &cachedUser); cacheError == nil {
			return &cachedUser, nil
		}
	}

	user, findByUsernameError := service.userRepository.FindByUsername(ctx, normalizedUsername)
	if findByUsernameError != nil {
		return nil, fmt.Errorf("UserService.FindByUsername: %w", findByUsernameError)
	}

	if service.cache != nil {
		_ = service.cache.Set(ctx, cacheKey, user, userCacheTTL)
	}

	return user, nil
}

func (service *UserService) List(ctx context.Context, page int, pageSize int) ([]domain.User, int64, error) {
	if page < 1 {
		page = defaultPage
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	users, totalCount, listError := service.userRepository.List(ctx, page, pageSize)
	if listError != nil {
		return nil, 0, fmt.Errorf("UserService.List: %w", listError)
	}

	return users, totalCount, nil
}

func (service *UserService) Create(ctx context.Context, user *domain.User) error {
	if user == nil {
		return domain.ErrValidation
	}

	user.Username = strings.TrimSpace(user.Username)
	user.PasswordHash = strings.TrimSpace(user.PasswordHash)
	user.FullName = strings.TrimSpace(user.FullName)
	user.Phone = strings.TrimSpace(user.Phone)
	if user.Username == "" || user.FullName == "" || user.Phone == "" || user.PasswordHash == "" {
		return domain.ErrValidation
	}

	hashedPassword, hashError := service.passwordHasher.Hash(user.PasswordHash)
	if hashError != nil {
		return fmt.Errorf("UserService.Create hash password: %w", hashError)
	}
	user.PasswordHash = hashedPassword

	existingUser, findByUsernameError := service.userRepository.FindByUsername(ctx, user.Username)
	if findByUsernameError == nil && existingUser != nil {
		return domain.ErrConflict
	}
	if findByUsernameError != nil && !errors.Is(findByUsernameError, domain.ErrNotFound) {
		return fmt.Errorf("UserService.Create find by username: %w", findByUsernameError)
	}

	if createError := service.userRepository.Create(ctx, user); createError != nil {
		return fmt.Errorf("UserService.Create: %w", createError)
	}
	if service.cache != nil {
		_ = service.cache.DeleteByPrefix(ctx, "user:list:")
		_ = service.cache.Delete(ctx, cache.UserByIDKey(user.ID), cache.UserByUsernameKey(user.Username))
	}

	return nil
}

func (service *UserService) Update(ctx context.Context, user *domain.User) error {
	if user == nil || user.ID == uuid.Nil {
		return domain.ErrValidation
	}

	user.Username = strings.TrimSpace(user.Username)
	user.FullName = strings.TrimSpace(user.FullName)
	user.Phone = strings.TrimSpace(user.Phone)
	if user.Username == "" || user.FullName == "" || user.Phone == "" {
		return domain.ErrValidation
	}

	currentUser, findByIDError := service.userRepository.FindByID(ctx, user.ID)
	if user.PasswordHash == "" {
		user.PasswordHash = currentUser.PasswordHash
	}

	if findByIDError != nil {
		return fmt.Errorf("UserService.Update find by id: %w", findByIDError)
	}

	if currentUser.Username != user.Username {
		existingUser, findByUsernameError := service.userRepository.FindByUsername(ctx, user.Username)
		if findByUsernameError == nil && existingUser != nil {
			return domain.ErrConflict
		}
		if findByUsernameError != nil && !errors.Is(findByUsernameError, domain.ErrNotFound) {
			return fmt.Errorf("UserService.Update find by username: %w", findByUsernameError)
		}
	}

	if updateError := service.userRepository.Update(ctx, user); updateError != nil {
		return fmt.Errorf("UserService.Update: %w", updateError)
	}
	if service.cache != nil {
		_ = service.cache.DeleteByPrefix(ctx, "user:list:")
		_ = service.cache.Delete(ctx, cache.UserByIDKey(user.ID), cache.UserByUsernameKey(user.Username))
		if currentUser.Username != user.Username {
			_ = service.cache.Delete(ctx, cache.UserByUsernameKey(currentUser.Username))
		}
	}

	return nil
}

func (service *UserService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return domain.ErrValidation
	}

	if softDeleteError := service.userRepository.SoftDelete(ctx, id); softDeleteError != nil {
		return fmt.Errorf("UserService.SoftDelete: %w", softDeleteError)
	}
	if service.cache != nil {
		_ = service.cache.DeleteByPrefix(ctx, "user:list:")
		_ = service.cache.Delete(ctx, cache.UserByIDKey(id))
	}

	return nil
}
