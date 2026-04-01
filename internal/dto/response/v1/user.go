package v1

import (
	"time"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
)

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	FullName  string    `json:"full_name"`
	Phone     string    `json:"phone"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserListResponse struct {
	Users    []UserResponse `json:"users"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
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
