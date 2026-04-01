package v1

import (
	"time"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
)

type UserResponse struct {
	ID        string    `json:"id" example:"01912e5a-4b7c-7f3a-9d1e-1a2b3c4d5e6f"`
	Username  string    `json:"username" example:"john_doe"`
	FullName  string    `json:"full_name" example:"John Doe"`
	Phone     string    `json:"phone" example:"+998901234567"`
	Status    string    `json:"status" example:"active"`
	CreatedAt time.Time `json:"created_at" example:"2026-03-30T12:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2026-03-31T12:00:00Z"`
}

type UserListResponse struct {
	Users    []UserResponse `json:"users"`
	Total    int64          `json:"total" example:"125"`
	Page     int            `json:"page" example:"1"`
	PageSize int            `json:"page_size" example:"20"`
}

func NewUserResponse(user domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		FullName:  user.FullName,
		Phone:     user.Phone,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func NewUserListResponse(users []domain.User, total int64, page int, pageSize int) UserListResponse {
	userResponses := make([]UserResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, NewUserResponse(user))
	}

	return UserListResponse{
		Users:    userResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}
