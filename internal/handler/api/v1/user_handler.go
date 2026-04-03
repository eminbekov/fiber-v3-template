package v1

import (
	"fmt"
	"strconv"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	requestDTO "github.com/eminbekov/fiber-v3-template/internal/dto/request"
	"github.com/eminbekov/fiber-v3-template/internal/dto/response"
	responseV1 "github.com/eminbekov/fiber-v3-template/internal/dto/response/v1"
	"github.com/eminbekov/fiber-v3-template/internal/i18n"
	"github.com/eminbekov/fiber-v3-template/internal/service"
	"github.com/gofiber/fiber/v3"
	"github.com/gofrs/uuid/v5"
)

type UserHandler struct {
	userService *service.UserService
	translator  *i18n.Translator
}

func NewUserHandler(userService *service.UserService, translator *i18n.Translator) *UserHandler {
	return &UserHandler{
		userService: userService,
		translator:  translator,
	}
}

// Create registers a new user.
//
// @Summary      Create user
// @Description  Creates a new user and returns the created user payload.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        body  body      requestDTO.CreateUserRequest  true  "Create user request"
// @Success      201   {object}  response.Response[responseV1.UserResponse]
// @Failure      400   {object}  response.ErrorResponse  "Validation error"
// @Failure      401   {object}  response.ErrorResponse  "Unauthorized"
// @Failure      403   {object}  response.ErrorResponse  "Forbidden"
// @Failure      500   {object}  response.ErrorResponse  "Internal server error"
// @Security     BearerAuth
// @Router       /v1/users [post]
func (handler *UserHandler) Create(ctx fiber.Ctx) error {
	language := extractLanguage(ctx)

	var request requestDTO.CreateUserRequest
	if bindError := ctx.Bind().Body(&request); bindError != nil {
		return domain.ErrValidation
	}
	request.Normalize()

	fieldErrors := requestDTO.ValidateDTO(request)
	if len(fieldErrors) > 0 {
		if jsonError := ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Error: response.ErrorBody{
				Message: handler.translator.Translate(language, "general.validation_failed"),
				Details: fieldErrors,
			},
		}); jsonError != nil {
			return fmt.Errorf("userHandler.Create: %w", jsonError)
		}
		return nil
	}

	user := &domain.User{
		Username:     request.Username,
		PasswordHash: request.Password,
		FullName:     request.FullName,
		Phone:        request.Phone,
		Status:       domain.UserStatusActive,
	}
	if createError := handler.userService.Create(ctx.Context(), user); createError != nil {
		return fmt.Errorf("userHandler.Create: %w", createError)
	}

	if jsonError := ctx.Status(fiber.StatusCreated).JSON(response.Response[responseV1.UserResponse]{
		Data: responseV1.NewUserResponse(*user),
	}); jsonError != nil {
		return fmt.Errorf("userHandler.Create: %w", jsonError)
	}
	return nil
}

// FindByID returns a user by ID.
//
// @Summary      Get user by ID
// @Description  Returns one user by UUID identifier.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  response.Response[responseV1.UserResponse]
// @Failure      400  {object}  response.ErrorResponse  "Invalid ID"
// @Failure      401  {object}  response.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  response.ErrorResponse  "Forbidden"
// @Failure      404  {object}  response.ErrorResponse  "User not found"
// @Failure      500  {object}  response.ErrorResponse  "Internal server error"
// @Security     BearerAuth
// @Router       /v1/users/{id} [get]
func (handler *UserHandler) FindByID(ctx fiber.Ctx) error {
	language := extractLanguage(ctx)

	id, idError := uuid.FromString(ctx.Params("id"))
	if idError != nil {
		if jsonError := ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Error: response.ErrorBody{
				Message: handler.translator.Translate(language, "validation.invalid_id"),
			},
		}); jsonError != nil {
			return fmt.Errorf("userHandler.FindByID: %w", jsonError)
		}
		return nil
	}

	user, findByIDError := handler.userService.FindByID(ctx.Context(), id)
	if findByIDError != nil {
		return fmt.Errorf("userHandler.FindByID: %w", findByIDError)
	}

	if jsonError := ctx.JSON(response.Response[responseV1.UserResponse]{
		Data: responseV1.NewUserResponse(*user),
	}); jsonError != nil {
		return fmt.Errorf("userHandler.FindByID: %w", jsonError)
	}
	return nil
}

// List returns a paginated list of users.
//
// @Summary      List users
// @Description  Returns users with page and page_size query parameters.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        page       query     int  false  "Page number"      default(1)
// @Param        page_size  query     int  false  "Items per page"   default(20)
// @Success      200        {object}  response.Response[responseV1.UserListResponse]
// @Failure      400        {object}  response.ErrorResponse  "Validation error"
// @Failure      401        {object}  response.ErrorResponse  "Unauthorized"
// @Failure      403        {object}  response.ErrorResponse  "Forbidden"
// @Failure      500        {object}  response.ErrorResponse  "Internal server error"
// @Security     BearerAuth
// @Router       /v1/users [get]
func (handler *UserHandler) List(ctx fiber.Ctx) error {
	language := extractLanguage(ctx)

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
		if jsonError := ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Error: response.ErrorBody{
				Message: handler.translator.Translate(language, "general.validation_failed"),
				Details: fieldErrors,
			},
		}); jsonError != nil {
			return fmt.Errorf("userHandler.List: %w", jsonError)
		}
		return nil
	}

	users, totalCount, listError := handler.userService.List(ctx.Context(), request.Page, request.PageSize)
	if listError != nil {
		return fmt.Errorf("userHandler.List: %w", listError)
	}

	if jsonError := ctx.JSON(response.Response[responseV1.UserListResponse]{
		Data: responseV1.NewUserListResponse(users, totalCount, request.Page, request.PageSize),
	}); jsonError != nil {
		return fmt.Errorf("userHandler.List: %w", jsonError)
	}
	return nil
}

// Update updates an existing user by ID.
//
// @Summary      Update user
// @Description  Updates user fields and returns the updated user payload.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id    path      string                       true  "User ID"
// @Param        body  body      requestDTO.UpdateUserRequest true  "Update user request"
// @Success      200   {object}  response.Response[responseV1.UserResponse]
// @Failure      400   {object}  response.ErrorResponse  "Validation error"
// @Failure      401   {object}  response.ErrorResponse  "Unauthorized"
// @Failure      403   {object}  response.ErrorResponse  "Forbidden"
// @Failure      404   {object}  response.ErrorResponse  "User not found"
// @Failure      500   {object}  response.ErrorResponse  "Internal server error"
// @Security     BearerAuth
// @Router       /v1/users/{id} [put]
func (handler *UserHandler) Update(ctx fiber.Ctx) error {
	language := extractLanguage(ctx)

	id, idError := uuid.FromString(ctx.Params("id"))
	if idError != nil {
		if jsonError := ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Error: response.ErrorBody{
				Message: handler.translator.Translate(language, "validation.invalid_id"),
			},
		}); jsonError != nil {
			return fmt.Errorf("userHandler.Update: %w", jsonError)
		}
		return nil
	}

	var request requestDTO.UpdateUserRequest
	if bindError := ctx.Bind().Body(&request); bindError != nil {
		return domain.ErrValidation
	}
	request.Normalize()

	fieldErrors := requestDTO.ValidateDTO(request)
	if len(fieldErrors) > 0 {
		if jsonError := ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Error: response.ErrorBody{
				Message: handler.translator.Translate(language, "general.validation_failed"),
				Details: fieldErrors,
			},
		}); jsonError != nil {
			return fmt.Errorf("userHandler.Update: %w", jsonError)
		}
		return nil
	}

	user := &domain.User{
		ID:       id,
		Username: request.Username,
		FullName: request.FullName,
		Phone:    request.Phone,
		Status:   request.Status,
	}
	requesterIDValue := ctx.Locals("user_id")
	requesterID, isRequesterID := requesterIDValue.(uuid.UUID)
	if !isRequesterID {
		return domain.ErrUnauthorized
	}
	if updateError := handler.userService.Update(ctx.Context(), requesterID, user); updateError != nil {
		return fmt.Errorf("userHandler.Update: %w", updateError)
	}

	updatedUser, findByIDError := handler.userService.FindByID(ctx.Context(), id)
	if findByIDError != nil {
		return fmt.Errorf("userHandler.Update: %w", findByIDError)
	}

	if jsonError := ctx.JSON(response.Response[responseV1.UserResponse]{
		Data: responseV1.NewUserResponse(*updatedUser),
	}); jsonError != nil {
		return fmt.Errorf("userHandler.Update: %w", jsonError)
	}
	return nil
}

// Delete soft-deletes a user by ID.
//
// @Summary      Delete user
// @Description  Soft-deletes a user and returns no content on success.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      204  {string}  string  "No Content"
// @Failure      400  {object}  response.ErrorResponse  "Invalid ID"
// @Failure      401  {object}  response.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  response.ErrorResponse  "Forbidden"
// @Failure      404  {object}  response.ErrorResponse  "User not found"
// @Failure      500  {object}  response.ErrorResponse  "Internal server error"
// @Security     BearerAuth
// @Router       /v1/users/{id} [delete]
func (handler *UserHandler) Delete(ctx fiber.Ctx) error {
	language := extractLanguage(ctx)

	id, idError := uuid.FromString(ctx.Params("id"))
	if idError != nil {
		if jsonError := ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse{
			Error: response.ErrorBody{
				Message: handler.translator.Translate(language, "validation.invalid_id"),
			},
		}); jsonError != nil {
			return fmt.Errorf("userHandler.Delete: %w", jsonError)
		}
		return nil
	}

	if softDeleteError := handler.userService.SoftDelete(ctx.Context(), id); softDeleteError != nil {
		return fmt.Errorf("userHandler.Delete: %w", softDeleteError)
	}

	if sendError := ctx.SendStatus(fiber.StatusNoContent); sendError != nil {
		return fmt.Errorf("userHandler.Delete: %w", sendError)
	}
	return nil
}

func extractLanguage(ctx fiber.Ctx) string {
	language, isLanguage := ctx.Locals("language").(string)
	if isLanguage && language != "" {
		return language
	}

	return "en"
}
