package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"github.com/eminbekov/fiber-v3-template/internal/session"
)

func TestAuthService_Login(testingContext *testing.T) {
	testingContext.Parallel()

	userID := uuid.Must(uuid.NewV7())

	testingContext.Run("returns token on successful login", func(testingContext *testing.T) {
		testingContext.Parallel()

		service := NewAuthService(
			&mockUserRepository{
				findByUsernameFunction: func(ctx context.Context, username string) (*domain.User, error) {
					return &domain.User{ID: userID, Username: username, PasswordHash: "hash-value"}, nil
				},
			},
			&mockSessionStore{
				createFunction: func(ctx context.Context, metadata session.Metadata) (string, error) { return "session-token", nil },
				getFunction:    func(ctx context.Context, token string) (*session.Metadata, error) { return nil, nil },
				deleteFunction: func(ctx context.Context, token string) error { return nil },
				extendFunction: func(ctx context.Context, token string) error { return nil },
			},
			&mockPasswordHasher{
				hashFunction:   func(password string) (string, error) { return password, nil },
				verifyFunction: func(password string, encodedHash string) (bool, error) { return true, nil },
			},
			time.Hour,
		)

		token, loginError := service.Login(context.Background(), "john", "strong-password", "127.0.0.1", "test-agent")
		if loginError != nil {
			testingContext.Fatalf("expected no error, got %v", loginError)
		}
		if token == "" {
			testingContext.Fatalf("expected non-empty token")
		}
	})

	testingContext.Run("returns validation error on empty credentials", func(testingContext *testing.T) {
		testingContext.Parallel()

		service := NewAuthService(emptyUserRepository(), emptySessionStore(), emptyPasswordHasher(), time.Hour)
		_, loginError := service.Login(context.Background(), " ", " ", "", "")
		if !errors.Is(loginError, domain.ErrValidation) {
			testingContext.Fatalf("expected ErrValidation, got %v", loginError)
		}
	})

	testingContext.Run("returns unauthorized when password does not match", func(testingContext *testing.T) {
		testingContext.Parallel()

		service := NewAuthService(
			&mockUserRepository{
				findByUsernameFunction: func(ctx context.Context, username string) (*domain.User, error) {
					return &domain.User{ID: userID, Username: username, PasswordHash: "stored-hash"}, nil
				},
			},
			emptySessionStore(),
			&mockPasswordHasher{
				hashFunction:   func(password string) (string, error) { return password, nil },
				verifyFunction: func(password string, encodedHash string) (bool, error) { return false, nil },
			},
			time.Hour,
		)

		_, loginError := service.Login(context.Background(), "john", "invalid", "", "")
		if !errors.Is(loginError, domain.ErrUnauthorized) {
			testingContext.Fatalf("expected ErrUnauthorized, got %v", loginError)
		}
	})
}

func TestAuthService_Logout(testingContext *testing.T) {
	testingContext.Parallel()

	service := NewAuthService(emptyUserRepository(), emptySessionStore(), emptyPasswordHasher(), time.Hour)
	logoutError := service.Logout(context.Background(), "")
	if !errors.Is(logoutError, domain.ErrValidation) {
		testingContext.Fatalf("expected ErrValidation, got %v", logoutError)
	}
}

func TestAuthService_Session(testingContext *testing.T) {
	testingContext.Parallel()

	testingContext.Run("returns session and extends when duration is positive", func(testingContext *testing.T) {
		testingContext.Parallel()

		extended := false
		service := NewAuthService(
			emptyUserRepository(),
			&mockSessionStore{
				createFunction: func(ctx context.Context, metadata session.Metadata) (string, error) { return "", nil },
				getFunction: func(ctx context.Context, token string) (*session.Metadata, error) {
					return &session.Metadata{UserID: uuid.Must(uuid.NewV7())}, nil
				},
				deleteFunction: func(ctx context.Context, token string) error { return nil },
				extendFunction: func(ctx context.Context, token string) error {
					extended = true
					return nil
				},
			},
			emptyPasswordHasher(),
			time.Minute,
		)

		metadata, sessionError := service.Session(context.Background(), "token")
		if sessionError != nil {
			testingContext.Fatalf("expected no error, got %v", sessionError)
		}
		if metadata == nil || metadata.UserID == uuid.Nil {
			testingContext.Fatalf("expected valid metadata")
		}
		if !extended {
			testingContext.Fatalf("expected session extension call")
		}
	})
}

func emptySessionStore() *mockSessionStore {
	return &mockSessionStore{
		createFunction: func(ctx context.Context, metadata session.Metadata) (string, error) { return "", nil },
		getFunction:    func(ctx context.Context, token string) (*session.Metadata, error) { return nil, domain.ErrNotFound },
		deleteFunction: func(ctx context.Context, token string) error { return nil },
		extendFunction: func(ctx context.Context, token string) error { return nil },
	}
}
