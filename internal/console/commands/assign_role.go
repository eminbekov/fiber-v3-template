package commands

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/gofrs/uuid/v5"
)

// AssignRole assigns a role to an existing user.
func AssignRole(ctx context.Context, dependencies *Dependencies, arguments []string) error {
	flagSet := flag.NewFlagSet("assign-role", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	userIDValue := flagSet.String("user-id", "", "user UUID (required)")
	roleNameValue := flagSet.String("role", "", "role name (required)")
	if parseError := flagSet.Parse(arguments); parseError != nil {
		return fmt.Errorf("assign-role parse flags: %w", parseError)
	}

	normalizedUserID := strings.TrimSpace(*userIDValue)
	normalizedRoleName := strings.TrimSpace(*roleNameValue)
	if normalizedUserID == "" || normalizedRoleName == "" {
		return fmt.Errorf("assign-role: --user-id and --role are required")
	}

	userID, parseUserIDError := uuid.FromString(normalizedUserID)
	if parseUserIDError != nil {
		return fmt.Errorf("assign-role parse user id: %w", parseUserIDError)
	}

	if _, findUserError := dependencies.UserService.FindByID(ctx, userID); findUserError != nil {
		return fmt.Errorf("assign-role find user: %w", findUserError)
	}

	role, findRoleError := dependencies.RoleRepository.FindByName(ctx, normalizedRoleName)
	if findRoleError != nil {
		return fmt.Errorf("assign-role find role: %w", findRoleError)
	}

	if assignRoleError := dependencies.RoleRepository.AssignToUser(ctx, userID, role.ID); assignRoleError != nil {
		return fmt.Errorf("assign-role persist assignment: %w", assignRoleError)
	}

	fmt.Printf("assigned role %q to user %s\n", normalizedRoleName, userID.String())
	return nil
}
