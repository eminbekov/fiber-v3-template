package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"github.com/eminbekov/fiber-v3-template/internal/repository"
	"github.com/eminbekov/fiber-v3-template/internal/session"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password string, encodedHash string) (bool, error)
}

type AuthService struct {
	userRepository  repository.UserRepository
	sessionStore    session.Store
	passwordHasher  PasswordHasher
	sessionDuration time.Duration
}

func NewAuthService(
	userRepository repository.UserRepository,
	sessionStore session.Store,
	passwordHasher PasswordHasher,
	sessionDuration time.Duration,
) *AuthService {
	return &AuthService{
		userRepository:  userRepository,
		sessionStore:    sessionStore,
		passwordHasher:  passwordHasher,
		sessionDuration: sessionDuration,
	}
}

func (service *AuthService) Login(
	ctx context.Context,
	username string,
	password string,
	ipAddress string,
	userAgent string,
) (string, error) {
	normalizedUsername := strings.TrimSpace(username)
	if normalizedUsername == "" || strings.TrimSpace(password) == "" {
		return "", domain.ErrValidation
	}

	user, findError := service.userRepository.FindByUsername(ctx, normalizedUsername)
	if findError != nil {
		return "", fmt.Errorf("authService.Login find by username: %w", findError)
	}

	isValidPassword, verifyError := service.passwordHasher.Verify(password, user.PasswordHash)
	if verifyError != nil {
		return "", fmt.Errorf("authService.Login verify: %w", verifyError)
	}
	if !isValidPassword {
		return "", domain.ErrUnauthorized
	}

	sessionToken, createError := service.sessionStore.Create(ctx, session.Metadata{
		UserID:    user.ID,
		Role:      "",
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: time.Now().UTC(),
	})
	if createError != nil {
		return "", fmt.Errorf("authService.Login create session: %w", createError)
	}

	return sessionToken, nil
}

func (service *AuthService) Logout(ctx context.Context, sessionToken string) error {
	if strings.TrimSpace(sessionToken) == "" {
		return domain.ErrValidation
	}

	if deleteError := service.sessionStore.Delete(ctx, sessionToken); deleteError != nil {
		return fmt.Errorf("authService.Logout: %w", deleteError)
	}

	return nil
}

func (service *AuthService) Session(ctx context.Context, sessionToken string) (*session.Metadata, error) {
	if strings.TrimSpace(sessionToken) == "" {
		return nil, domain.ErrValidation
	}

	metadata, getError := service.sessionStore.Get(ctx, sessionToken)
	if getError != nil {
		return nil, fmt.Errorf("authService.Session: %w", getError)
	}

	if service.sessionDuration > 0 {
		_ = service.sessionStore.Extend(ctx, sessionToken)
	}

	return metadata, nil
}

func (service *AuthService) SessionDuration() time.Duration {
	return service.sessionDuration
}
