package request

import (
	"strings"

	"github.com/eminbekov/fiber-v3-template/internal/dto/response"
	"github.com/go-playground/validator/v10"
)

var dtoValidator = validator.New()

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required,max=100"`
	Phone    string `json:"phone" validate:"required,e164"`
}

type UpdateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	FullName string `json:"full_name" validate:"required,max=100"`
	Phone    string `json:"phone" validate:"required,e164"`
	Status   string `json:"status" validate:"required,oneof=active disabled"`
}

type ListUsersRequest struct {
	Page     int `query:"page" validate:"omitempty,min=1"`
	PageSize int `query:"page_size" validate:"omitempty,min=1,max=100"`
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

	validationErrors, ok := validationError.(validator.ValidationErrors)
	if !ok {
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
