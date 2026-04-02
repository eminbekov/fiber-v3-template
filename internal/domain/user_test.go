package domain

import (
	"testing"
	"time"
)

func TestUser_IsActive(testingContext *testing.T) {
	testingContext.Parallel()

	now := time.Now()
	tests := []struct {
		name     string
		user     *User
		expected bool
	}{
		{
			name: "active user returns true",
			user: &User{
				Status: UserStatusActive,
			},
			expected: true,
		},
		{
			name: "disabled user returns false",
			user: &User{
				Status: UserStatusDisabled,
			},
			expected: false,
		},
		{
			name: "soft deleted user returns false",
			user: &User{
				Status:    UserStatusActive,
				DeletedAt: &now,
			},
			expected: false,
		},
		{
			name:     "nil user returns false",
			user:     nil,
			expected: false,
		},
	}

	for _, testCase := range tests {
		testCase := testCase
		testingContext.Run(testCase.name, func(testingContext *testing.T) {
			testingContext.Parallel()

			actual := testCase.user.IsActive()
			if actual != testCase.expected {
				testingContext.Fatalf("expected IsActive=%v, got %v", testCase.expected, actual)
			}
		})
	}
}

func TestUser_IsDeleted(testingContext *testing.T) {
	testingContext.Parallel()

	now := time.Now()
	tests := []struct {
		name     string
		user     *User
		expected bool
	}{
		{
			name: "soft deleted user returns true",
			user: &User{
				DeletedAt: &now,
			},
			expected: true,
		},
		{
			name: "non deleted user returns false",
			user: &User{
				DeletedAt: nil,
			},
			expected: false,
		},
		{
			name:     "nil user returns false",
			user:     nil,
			expected: false,
		},
	}

	for _, testCase := range tests {
		testCase := testCase
		testingContext.Run(testCase.name, func(testingContext *testing.T) {
			testingContext.Parallel()

			actual := testCase.user.IsDeleted()
			if actual != testCase.expected {
				testingContext.Fatalf("expected IsDeleted=%v, got %v", testCase.expected, actual)
			}
		})
	}
}
