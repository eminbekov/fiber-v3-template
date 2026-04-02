package domain

import "testing"

func TestPermission_String(testingContext *testing.T) {
	testingContext.Parallel()

	tests := []struct {
		name       string
		permission Permission
		expected   string
	}{
		{
			name: "resource and action joined with dot",
			permission: Permission{
				Resource: "user",
				Action:   "read",
			},
			expected: "user.read",
		},
		{
			name: "wildcard action keeps wildcard",
			permission: Permission{
				Resource: "admin",
				Action:   "*",
			},
			expected: "admin.*",
		},
	}

	for _, testCase := range tests {
		testCase := testCase
		testingContext.Run(testCase.name, func(testingContext *testing.T) {
			testingContext.Parallel()

			actual := testCase.permission.String()
			if actual != testCase.expected {
				testingContext.Fatalf("expected %q, got %q", testCase.expected, actual)
			}
		})
	}
}
