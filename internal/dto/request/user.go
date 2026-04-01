package request

import (
	"errors"
	"strings"

	"github.com/eminbekov/fiber-v3-template/internal/dto/response"
	"github.com/go-playground/validator/v10"
)

var dtoValidator = validator.New()

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum" example:"john_doe"`
	Password string `json:"password" validate:"required,min=8" example:"StrongPass123!"`
	FullName string `json:"full_name" validate:"required,max=100" example:"John Doe"`
	Phone    string `json:"phone" validate:"required,e164" example:"+998901234567"`
}

type UpdateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum" example:"john_doe"`
	FullName string `json:"full_name" validate:"required,max=100" example:"John Doe"`
	Phone    string `json:"phone" validate:"required,e164" example:"+998901234567"`
	Status   string `json:"status" validate:"required,oneof=active disabled" example:"active"`
}

type ListUsersRequest struct {
	Page     int `query:"page" validate:"omitempty,min=1" example:"1"`
	PageSize int `query:"page_size" validate:"omitempty,min=1,max=100" example:"20"`
}

func (request *CreateUserRequest) Normalize() {
	request.Username = strings.TrimSpace(request.Username)
	request.Password = strings.TrimSpace(request.Password)
	request.FullName = strings.TrimSpace(request.FullName)
	request.Phone = strings.TrimSpace(request.Phone)
}

func (request *UpdateUserRequest) Normalize() {
	request.Username = strings.TrimSpace(request.Username)
	request.FullName = strings.TrimSpace(request.FullName)
	request.Phone = strings.TrimSpace(request.Phone)
	request.Status = strings.TrimSpace(request.Status)
}

func (request *ListUsersRequest) ApplyDefaults() {
	if request.Page <= 0 {
		request.Page = 1
	}
	if request.PageSize <= 0 {
		request.PageSize = 20
	}
}

func ValidateDTO(target any) []response.FieldError {
	validationError := dtoValidator.Struct(target)
	if validationError == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(validationError, &validationErrors) {
		return []response.FieldError{
			{
				Field:   "request",
				Message: "invalid request payload",
			},
		}
	}

	fieldErrors := make([]response.FieldError, 0, len(validationErrors))
	for _, item := range validationErrors {
		fieldErrors = append(fieldErrors, response.FieldError{
			Field:   strings.ToLower(item.Field()),
			Message: item.Error(),
		})
	}

	return fieldErrors
}
