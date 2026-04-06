package domain

import "testing"

func TestRole_FieldAssignment(testingContext *testing.T) {
	testingContext.Parallel()

	tests := []struct {
		name              string
		role              Role
		expectedID        int64
		expectedName      string
		permissionCount   int
	}{
		{
			name:            "role with permissions",
			role:            Role{ID: 1, Name: "admin", Description: "Administrator", Permissions: []Permission{{ID: 1}, {ID: 2}}},
			expectedID:      1,
			expectedName:    "admin",
			permissionCount: 2,
		},
		{
			name:            "role without permissions",
			role:            Role{ID: 2, Name: "viewer", Description: "Read-only user"},
			expectedID:      2,
			expectedName:    "viewer",
			permissionCount: 0,
		},
		{
			name:            "zero value role",
			role:            Role{},
			expectedID:      0,
			expectedName:    "",
			permissionCount: 0,
		},
	}

	for _, testCase := range tests {
		testCase := testCase
		testingContext.Run(testCase.name, func(testingContext *testing.T) {
			testingContext.Parallel()

			if testCase.role.ID != testCase.expectedID {
				testingContext.Fatalf("expected ID=%d, got %d", testCase.expectedID, testCase.role.ID)
			}
			if testCase.role.Name != testCase.expectedName {
				testingContext.Fatalf("expected Name=%q, got %q", testCase.expectedName, testCase.role.Name)
			}
			if len(testCase.role.Permissions) != testCase.permissionCount {
				testingContext.Fatalf("expected %d permissions, got %d", testCase.permissionCount, len(testCase.role.Permissions))
			}
		})
	}
}
