package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

const (
	UserStatusActive   = "active"
	UserStatusDisabled = "disabled"
)

// User is a pure domain model without framework dependencies.
type User struct {
	ID           uuid.UUID  `db:"id"`
	Username     string     `db:"username"`
	PasswordHash string     `db:"password_hash"`
	FullName     string     `db:"full_name"`
	Phone        string     `db:"phone"`
	Status       string     `db:"status"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
}

// IsActive reports whether user can be treated as active.
func (user *User) IsActive() bool {
	return user != nil && user.Status == UserStatusActive && user.DeletedAt == nil
}

// IsDeleted reports whether user is soft-deleted.
func (user *User) IsDeleted() bool {
	return user != nil && user.DeletedAt != nil
}
