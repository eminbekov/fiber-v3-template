package postgres

import (
	"context"
	"testing"

	"github.com/gofrs/uuid/v5"
)

func TestPermissionRepository_ListAndFind(testingContext *testing.T) {
	roleRepository, pool, cleanup := newIntegrationRoleRepository(testingContext)
	defer cleanup()

	permissionRepository := &permissionRepository{pool: pool}
	requestContext := context.Background()

	permissions, listError := permissionRepository.List(requestContext)
	if listError != nil {
		testingContext.Fatalf("list permissions: %v", listError)
	}
	if len(permissions) == 0 {
		testingContext.Fatalf("expected seeded permissions")
	}

	managerRole, findRoleError := roleRepository.FindByName(requestContext, "manager")
	if findRoleError != nil {
		testingContext.Fatalf("find manager role: %v", findRoleError)
	}

	rolePermissions, findByRoleIDError := permissionRepository.FindByRoleID(requestContext, managerRole.ID)
	if findByRoleIDError != nil {
		testingContext.Fatalf("find permissions by role id: %v", findByRoleIDError)
	}
	if len(rolePermissions) == 0 {
		testingContext.Fatalf("expected permissions for manager role")
	}
}

func TestPermissionRepository_FindByUserID(testingContext *testing.T) {
	roleRepository, pool, cleanup := newIntegrationRoleRepository(testingContext)
	defer cleanup()

	permissionRepository := &permissionRepository{pool: pool}
	requestContext := context.Background()
	userID := uuid.Must(uuid.NewV7())
	if insertUserError := insertTestUser(requestContext, pool, userID, "permission-user"); insertUserError != nil {
		testingContext.Fatalf("insert test user: %v", insertUserError)
	}

	viewerRole, findRoleError := roleRepository.FindByName(requestContext, "viewer")
	if findRoleError != nil {
		testingContext.Fatalf("find viewer role: %v", findRoleError)
	}
	if assignError := roleRepository.AssignToUser(requestContext, userID, viewerRole.ID); assignError != nil {
		testingContext.Fatalf("assign role to user: %v", assignError)
	}

	userPermissions, findByUserIDError := permissionRepository.FindByUserID(requestContext, userID)
	if findByUserIDError != nil {
		testingContext.Fatalf("find permissions by user id: %v", findByUserIDError)
	}
	if len(userPermissions) == 0 {
		testingContext.Fatalf("expected viewer permissions")
	}
	for _, permission := range userPermissions {
		if permission.Action != "read" {
			testingContext.Fatalf("viewer should only have read actions, got %q", permission.Action)
		}
	}
}
