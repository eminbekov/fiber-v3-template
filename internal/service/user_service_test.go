package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/eminbekov/fiber-v3-template/internal/cache"
	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"github.com/gofrs/uuid/v5"
)

func TestUserService_FindByID(testingContext *testing.T) {
	testingContext.Parallel()

	userID := uuid.Must(uuid.NewV7())
	expectedUser := &domain.User{ID: userID, Username: "cached-user"}

	testingContext.Run("returns cached user on cache hit", func(testingContext *testing.T) {
		testingContext.Parallel()

		service := NewUserService(
			&mockUserRepository{
				findByIDFunction: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
					testingContext.Fatalf("repository should not be called on cache hit")
					return nil, nil
				},
			},
			emptyRoleRepository(),
			&mockCache{
				getFunction: func(ctx context.Context, key string, destination any) error {
					cachedUser, ok := destination.(*domain.User)
					if !ok {
						testingContext.Fatalf("destination must be *domain.User")
					}
					*cachedUser = *expectedUser
					return nil
				},
				setFunction:            func(ctx context.Context, key string, value any, ttlDuration time.Duration) error { return nil },
				deleteFunction:         func(ctx context.Context, keys ...string) error { return nil },
				deleteByPrefixFunction: func(ctx context.Context, prefix string) error { return nil },
			},
			emptyPasswordHasher(),
		)

		user, findError := service.FindByID(context.Background(), userID)
		if findError != nil {
			testingContext.Fatalf("expected no error, got %v", findError)
		}
		if user.Username != expectedUser.Username {
			testingContext.Fatalf("expected %q, got %q", expectedUser.Username, user.Username)
		}
	})

	testingContext.Run("returns user from repository on cache miss", func(testingContext *testing.T) {
		testingContext.Parallel()

		repositoryCalled := false
		service := NewUserService(
			&mockUserRepository{
				findByIDFunction: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
					repositoryCalled = true
					return expectedUser, nil
				},
			},
			emptyRoleRepository(),
			&mockCache{
				getFunction:            func(ctx context.Context, key string, destination any) error { return cache.ErrCacheMiss },
				setFunction:            func(ctx context.Context, key string, value any, ttlDuration time.Duration) error { return nil },
				deleteFunction:         func(ctx context.Context, keys ...string) error { return nil },
				deleteByPrefixFunction: func(ctx context.Context, prefix string) error { return nil },
			},
			emptyPasswordHasher(),
		)

		_, findError := service.FindByID(context.Background(), userID)
		if findError != nil {
			testingContext.Fatalf("expected no error, got %v", findError)
		}
		if !repositoryCalled {
			testingContext.Fatalf("expected repository call on cache miss")
		}
	})

	testingContext.Run("returns validation error for nil uuid", func(testingContext *testing.T) {
		testingContext.Parallel()

		service := NewUserService(emptyUserRepository(), emptyRoleRepository(), nil, emptyPasswordHasher())
		_, findError := service.FindByID(context.Background(), uuid.Nil)
		if !errors.Is(findError, domain.ErrValidation) {
			testingContext.Fatalf("expected ErrValidation, got %v", findError)
		}
	})
}

func TestUserService_FindByUsername(testingContext *testing.T) {
	testingContext.Parallel()

	testingContext.Run("validates empty username", func(testingContext *testing.T) {
		testingContext.Parallel()

		service := NewUserService(emptyUserRepository(), emptyRoleRepository(), nil, emptyPasswordHasher())
		_, findError := service.FindByUsername(context.Background(), " ")
		if !errors.Is(findError, domain.ErrValidation) {
			testingContext.Fatalf("expected ErrValidation, got %v", findError)
		}
	})

	testingContext.Run("returns repository value after cache miss", func(testingContext *testing.T) {
		testingContext.Parallel()

		expectedUser := &domain.User{Username: "db-user"}
		service := NewUserService(
			&mockUserRepository{
				findByUsernameFunction: func(ctx context.Context, username string) (*domain.User, error) {
					return expectedUser, nil
				},
			},
			emptyRoleRepository(),
			&mockCache{
				getFunction:            func(ctx context.Context, key string, destination any) error { return cache.ErrCacheMiss },
				setFunction:            func(ctx context.Context, key string, value any, ttlDuration time.Duration) error { return nil },
				deleteFunction:         func(ctx context.Context, keys ...string) error { return nil },
				deleteByPrefixFunction: func(ctx context.Context, prefix string) error { return nil },
			},
			emptyPasswordHasher(),
		)

		user, findError := service.FindByUsername(context.Background(), " db-user ")
		if findError != nil {
			testingContext.Fatalf("expected no error, got %v", findError)
		}
		if user.Username != expectedUser.Username {
			testingContext.Fatalf("expected %q, got %q", expectedUser.Username, user.Username)
		}
	})
}

func TestUserService_List(testingContext *testing.T) {
	testingContext.Parallel()

	testingContext.Run("applies defaults and page size cap", func(testingContext *testing.T) {
		testingContext.Parallel()

		capturedPage := 0
		capturedPageSize := 0
		service := NewUserService(
			&mockUserRepository{
				listFunction: func(ctx context.Context, page int, pageSize int) ([]domain.User, int64, error) {
					capturedPage = page
					capturedPageSize = pageSize
					return []domain.User{}, 0, nil
				},
			},
			emptyRoleRepository(),
			nil,
			emptyPasswordHasher(),
		)

		_, _, listError := service.List(context.Background(), 0, 200)
		if listError != nil {
			testingContext.Fatalf("expected no error, got %v", listError)
		}
		if capturedPage != 1 || capturedPageSize != 100 {
			testingContext.Fatalf("expected page=1 page_size=100, got page=%d page_size=%d", capturedPage, capturedPageSize)
		}
	})
}

func TestUserService_Create(testingContext *testing.T) {
	testingContext.Parallel()

	testingContext.Run("creates user and hashes password", func(testingContext *testing.T) {
		testingContext.Parallel()

		hashedPassword := ""
		service := NewUserService(
			&mockUserRepository{
				findByUsernameFunction: func(ctx context.Context, username string) (*domain.User, error) {
					return nil, domain.ErrNotFound
				},
				createFunction: func(ctx context.Context, user *domain.User) error {
					hashedPassword = user.PasswordHash
					return nil
				},
			},
			emptyRoleRepository(),
			&mockCache{
				getFunction:            func(ctx context.Context, key string, destination any) error { return cache.ErrCacheMiss },
				setFunction:            func(ctx context.Context, key string, value any, ttlDuration time.Duration) error { return nil },
				deleteFunction:         func(ctx context.Context, keys ...string) error { return nil },
				deleteByPrefixFunction: func(ctx context.Context, prefix string) error { return nil },
			},
			&mockPasswordHasher{
				hashFunction:   func(password string) (string, error) { return "hashed-value", nil },
				verifyFunction: func(password string, encodedHash string) (bool, error) { return true, nil },
			},
		)

		createError := service.Create(context.Background(), &domain.User{
			ID:           uuid.Must(uuid.NewV7()),
			Username:     "username",
			PasswordHash: "plain-text",
			FullName:     "User Name",
			Phone:        "+998901234567",
		})
		if createError != nil {
			testingContext.Fatalf("expected no error, got %v", createError)
		}
		if hashedPassword != "hashed-value" {
			testingContext.Fatalf("expected hashed password to be stored, got %q", hashedPassword)
		}
	})

	testingContext.Run("returns conflict for duplicate username", func(testingContext *testing.T) {
		testingContext.Parallel()

		service := NewUserService(
			&mockUserRepository{
				findByUsernameFunction: func(ctx context.Context, username string) (*domain.User, error) {
					return &domain.User{Username: username}, nil
				},
			},
			emptyRoleRepository(),
			nil,
			emptyPasswordHasher(),
		)

		createError := service.Create(context.Background(), &domain.User{
			Username:     "existing",
			PasswordHash: "password",
			FullName:     "Existing User",
			Phone:        "+998901111111",
		})
		if !errors.Is(createError, domain.ErrConflict) {
			testingContext.Fatalf("expected ErrConflict, got %v", createError)
		}
	})
}

func TestUserService_Update_ForbidsNonAdmin(testingContext *testing.T) {
	testingContext.Parallel()

	requesterID := uuid.Must(uuid.NewV7())
	targetID := uuid.Must(uuid.NewV7())

	service := NewUserService(
		&mockUserRepository{
			findByIDFunction: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
				return &domain.User{ID: targetID, Username: "target", PasswordHash: "hash"}, nil
			},
		},
		mockRoleRepositoryWithRoles([]domain.Role{{Name: "user"}}),
		nil,
		emptyPasswordHasher(),
	)

	updateError := service.Update(context.Background(), requesterID, &domain.User{
		ID:       targetID,
		Username: "target",
		FullName: "Target User",
		Phone:    "+998901234000",
	})
	if !errors.Is(updateError, domain.ErrForbidden) {
		testingContext.Fatalf("expected ErrForbidden, got %v", updateError)
	}
}

func TestUserService_Update_AllowsAdmin(testingContext *testing.T) {
	testingContext.Parallel()

	requesterID := uuid.Must(uuid.NewV7())
	targetID := uuid.Must(uuid.NewV7())
	updateCalled := false
	service := NewUserService(
		&mockUserRepository{
			findByIDFunction: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
				return &domain.User{ID: targetID, Username: "target", PasswordHash: "hash"}, nil
			},
			findByUsernameFunction: func(ctx context.Context, username string) (*domain.User, error) {
				return nil, domain.ErrNotFound
			},
			updateFunction: func(ctx context.Context, user *domain.User) error {
				updateCalled = true
				return nil
			},
		},
		mockRoleRepositoryWithRoles([]domain.Role{{Name: "admin"}}),
		nil,
		emptyPasswordHasher(),
	)

	updateError := service.Update(context.Background(), requesterID, &domain.User{
		ID:       targetID,
		Username: "new-target",
		FullName: "Target User",
		Phone:    "+998901234000",
	})
	if updateError != nil {
		testingContext.Fatalf("expected no error, got %v", updateError)
	}
	if !updateCalled {
		testingContext.Fatalf("expected update to be called")
	}
}

func TestUserService_SoftDelete(testingContext *testing.T) {
	testingContext.Parallel()

	service := NewUserService(
		&mockUserRepository{
			softDeleteFunction: func(ctx context.Context, id uuid.UUID) error { return nil },
		},
		emptyRoleRepository(),
		nil,
		emptyPasswordHasher(),
	)

	softDeleteError := service.SoftDelete(context.Background(), uuid.Nil)
	if !errors.Is(softDeleteError, domain.ErrValidation) {
		testingContext.Fatalf("expected ErrValidation, got %v", softDeleteError)
	}
}

func emptyUserRepository() *mockUserRepository {
	return &mockUserRepository{
		findByIDFunction:       func(ctx context.Context, id uuid.UUID) (*domain.User, error) { return nil, domain.ErrNotFound },
		findByUsernameFunction: func(ctx context.Context, username string) (*domain.User, error) { return nil, domain.ErrNotFound },
		listFunction:           func(ctx context.Context, page int, pageSize int) ([]domain.User, int64, error) { return nil, 0, nil },
		createFunction:         func(ctx context.Context, user *domain.User) error { return nil },
		updateFunction:         func(ctx context.Context, user *domain.User) error { return nil },
		softDeleteFunction:     func(ctx context.Context, id uuid.UUID) error { return nil },
	}
}

func emptyRoleRepository() *mockRoleRepository {
	return &mockRoleRepository{
		findByIDFunction:       func(ctx context.Context, id int64) (*domain.Role, error) { return nil, domain.ErrNotFound },
		findByNameFunction:     func(ctx context.Context, name string) (*domain.Role, error) { return nil, domain.ErrNotFound },
		findByUserIDFunction:   func(ctx context.Context, userID uuid.UUID) ([]domain.Role, error) { return nil, nil },
		listFunction:           func(ctx context.Context) ([]domain.Role, error) { return nil, nil },
		createFunction:         func(ctx context.Context, role *domain.Role) error { return nil },
		assignToUserFunction:   func(ctx context.Context, userID uuid.UUID, roleID int64) error { return nil },
		removeFromUserFunction: func(ctx context.Context, userID uuid.UUID, roleID int64) error { return nil },
	}
}

func mockRoleRepositoryWithRoles(roles []domain.Role) *mockRoleRepository {
	return &mockRoleRepository{
		findByIDFunction:       func(ctx context.Context, id int64) (*domain.Role, error) { return nil, domain.ErrNotFound },
		findByNameFunction:     func(ctx context.Context, name string) (*domain.Role, error) { return nil, domain.ErrNotFound },
		findByUserIDFunction:   func(ctx context.Context, userID uuid.UUID) ([]domain.Role, error) { return roles, nil },
		listFunction:           func(ctx context.Context) ([]domain.Role, error) { return nil, nil },
		createFunction:         func(ctx context.Context, role *domain.Role) error { return nil },
		assignToUserFunction:   func(ctx context.Context, userID uuid.UUID, roleID int64) error { return nil },
		removeFromUserFunction: func(ctx context.Context, userID uuid.UUID, roleID int64) error { return nil },
	}
}

func emptyPasswordHasher() *mockPasswordHasher {
	return &mockPasswordHasher{
		hashFunction:   func(password string) (string, error) { return password, nil },
		verifyFunction: func(password string, encodedHash string) (bool, error) { return password == encodedHash, nil },
	}
}
