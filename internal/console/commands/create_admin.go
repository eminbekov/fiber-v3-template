package commands

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
)

// CreateAdmin creates an admin user and assigns the admin role.
func CreateAdmin(ctx context.Context, dependencies *Dependencies, arguments []string) error {
	flagSet := flag.NewFlagSet("create-admin", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	usernameValue := flagSet.String("username", "", "admin username (required)")
	passwordValue := flagSet.String("password", "", "admin password (required)")
	fullNameValue := flagSet.String("full-name", "Administrator", "admin full name")
	phoneValue := flagSet.String("phone", "", "admin phone in E.164 format (required)")

	if parseError := flagSet.Parse(arguments); parseError != nil {
		return fmt.Errorf("create-admin parse flags: %w", parseError)
	}

	normalizedUsername := strings.TrimSpace(*usernameValue)
	normalizedPassword := strings.TrimSpace(*passwordValue)
	normalizedFullName := strings.TrimSpace(*fullNameValue)
	normalizedPhone := strings.TrimSpace(*phoneValue)
	if normalizedUsername == "" || normalizedPassword == "" || normalizedPhone == "" {
		return fmt.Errorf("create-admin: --username, --password, and --phone are required")
	}

	user := &domain.User{
		Username:     normalizedUsername,
		PasswordHash: normalizedPassword,
		FullName:     normalizedFullName,
		Phone:        normalizedPhone,
		Status:       domain.UserStatusActive,
	}
	if createUserError := dependencies.UserService.Create(ctx, user); createUserError != nil {
		return fmt.Errorf("create-admin create user: %w", createUserError)
	}

	adminRole, findRoleError := dependencies.RoleRepository.FindByName(ctx, "admin")
	if findRoleError != nil {
		return fmt.Errorf("create-admin find admin role: %w", findRoleError)
	}

	if assignRoleError := dependencies.RoleRepository.AssignToUser(ctx, user.ID, adminRole.ID); assignRoleError != nil {
		return fmt.Errorf("create-admin assign role: %w", assignRoleError)
	}

	fmt.Printf("created admin user %q with id %s\n", user.Username, user.ID.String())
	return nil
}
