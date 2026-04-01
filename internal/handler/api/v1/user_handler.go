package v1

import (
	"errors"
	"strconv"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	requestDTO "github.com/eminbekov/fiber-v3-template/internal/dto/request"
	"github.com/eminbekov/fiber-v3-template/internal/dto/response"
	responseV1 "github.com/eminbekov/fiber-v3-template/internal/dto/response/v1"
	"github.com/eminbekov/fiber-v3-template/internal/service"
	"github.com/gofiber/fiber/v3"
	"github.com/gofrs/uuid/v5"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (handler *UserHandler) Create(ctx fiber.Ctx) error {
	var request requestDTO.CreateUserRequest
	if bindError := ctx.Bind().Body(&request); bindError != nil {
		return domain.ErrValidation
	}
	request.Normalize()

	fieldErrors := requestDTO.ValidateDTO(request)
	if len(fieldErrors) > 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Error: response.ErrorBody{
				Message: "validation failed",
				Details: fieldErrors,
			},
		})
	}

	user := &domain.User{
		Username: request.Username,
		FullName: request.FullName,
		Phone:    request.Phone,
		Status:   domain.UserStatusActive,
	}
	if createError := handler.userService.Create(ctx.Context(), user); createError != nil {
		return createError
	}

	return ctx.Status(fiber.StatusCreated).JSON(response.Response{
		Data: responseV1.NewUserResponse(*user),
	})
}

func (handler *UserHandler) FindByID(ctx fiber.Ctx) error {
	id, idError := uuid.FromString(ctx.Params("id"))
	if idError != nil {
		return domain.ErrValidation
	}

	user, findByIDError := handler.userService.FindByID(ctx.Context(), id)
	if findByIDError != nil {
		return findByIDError
	}

	return ctx.JSON(response.Response{
		Data: responseV1.NewUserResponse(*user),
	})
}

func (handler *UserHandler) List(ctx fiber.Ctx) error {
	request := requestDTO.ListUsersRequest{}
	if rawPage := ctx.Query("page"); rawPage != "" {
		page, parseError := strconv.Atoi(rawPage)
		if parseError != nil {
			return domain.ErrValidation
		}
		request.Page = page
	}
	if rawPageSize := ctx.Query("page_size"); rawPageSize != "" {
		pageSize, parseError := strconv.Atoi(rawPageSize)
		if parseError != nil {
			return domain.ErrValidation
		}
		request.PageSize = pageSize
	}
	request.ApplyDefaults()

	fieldErrors := requestDTO.ValidateDTO(request)
	if len(fieldErrors) > 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Error: response.ErrorBody{
				Message: "validation failed",
				Details: fieldErrors,
			},
		})
	}

	users, totalCount, listError := handler.userService.List(ctx.Context(), request.Page, request.PageSize)
	if listError != nil {
		return listError
	}

	return ctx.JSON(response.Response{
		Data: responseV1.NewUserListResponse(users, totalCount, request.Page, request.PageSize),
	})
}

func (handler *UserHandler) Update(ctx fiber.Ctx) error {
	id, idError := uuid.FromString(ctx.Params("id"))
	if idError != nil {
		return domain.ErrValidation
	}

	var request requestDTO.UpdateUserRequest
	if bindError := ctx.Bind().Body(&request); bindError != nil {
		return domain.ErrValidation
	}
	request.Normalize()

	fieldErrors := requestDTO.ValidateDTO(request)
	if len(fieldErrors) > 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Error: response.ErrorBody{
				Message: "validation failed",
				Details: fieldErrors,
			},
		})
	}

	user := &domain.User{
		ID:       id,
		Username: request.Username,
		FullName: request.FullName,
		Phone:    request.Phone,
		Status:   request.Status,
	}
	if updateError := handler.userService.Update(ctx.Context(), user); updateError != nil {
		return updateError
	}

	updatedUser, findByIDError := handler.userService.FindByID(ctx.Context(), id)
	if findByIDError != nil {
		return findByIDError
	}

	return ctx.JSON(response.Response{
		Data: responseV1.NewUserResponse(*updatedUser),
	})
}

func (handler *UserHandler) Delete(ctx fiber.Ctx) error {
	id, idError := uuid.FromString(ctx.Params("id"))
	if idError != nil {
		return domain.ErrValidation
	}

	softDeleteError := handler.userService.SoftDelete(ctx.Context(), id)
	if softDeleteError != nil {
		if errors.Is(softDeleteError, domain.ErrNotFound) {
			return softDeleteError
		}
		return softDeleteError
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
