package grpc

import (
	"context"

	userv1 "github.com/eminbekov/fiber-v3-template/gen/proto/user/v1"
	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"github.com/eminbekov/fiber-v3-template/internal/service"
)

type UserServer struct {
	userv1.UnimplementedUserServiceServer
	userService *service.UserService
}

func NewUserServer(userService *service.UserService) userv1.UserServiceServer {
	return &UserServer{
		userService: userService,
	}
}

func (server *UserServer) GetUser(ctx context.Context, request *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	userID, parseError := protoUserIDToUUID(request.GetId())
	if parseError != nil {
		return nil, toGRPCError(domain.ErrValidation)
	}

	user, findError := server.userService.FindByID(ctx, userID)
	if findError != nil {
		return nil, toGRPCError(findError)
	}

	return &userv1.GetUserResponse{
		User: domainUserToProto(user),
	}, nil
}

func (server *UserServer) ListUsers(ctx context.Context, request *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	users, totalCount, listError := server.userService.List(ctx, int(request.GetPage()), int(request.GetPageSize()))
	if listError != nil {
		return nil, toGRPCError(listError)
	}

	userList := make([]*userv1.User, 0, len(users))
	for _, item := range users {
		domainUser := item
		userList = append(userList, domainUserToProto(&domainUser))
	}

	return &userv1.ListUsersResponse{
		Users:      userList,
		TotalCount: totalCount,
	}, nil
}

func (server *UserServer) CreateUser(ctx context.Context, request *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	user := &domain.User{
		Username:     request.GetUsername(),
		PasswordHash: request.GetPassword(),
		FullName:     request.GetFullName(),
		Phone:        request.GetPhone(),
		Status:       domain.UserStatusActive,
	}

	if createError := server.userService.Create(ctx, user); createError != nil {
		return nil, toGRPCError(createError)
	}

	return &userv1.CreateUserResponse{
		User: domainUserToProto(user),
	}, nil
}

func (server *UserServer) UpdateUser(ctx context.Context, request *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	requesterID, requesterIDError := protoUserIDToUUID(request.GetRequesterId())
	if requesterIDError != nil {
		return nil, toGRPCError(domain.ErrValidation)
	}

	userID, userIDError := protoUserIDToUUID(request.GetId())
	if userIDError != nil {
		return nil, toGRPCError(domain.ErrValidation)
	}

	user := &domain.User{
		ID:       userID,
		Username: request.GetUsername(),
		FullName: request.GetFullName(),
		Phone:    request.GetPhone(),
		Status:   request.GetStatus(),
	}

	if updateError := server.userService.Update(ctx, requesterID, user); updateError != nil {
		return nil, toGRPCError(updateError)
	}

	updatedUser, findError := server.userService.FindByID(ctx, userID)
	if findError != nil {
		return nil, toGRPCError(findError)
	}

	return &userv1.UpdateUserResponse{
		User: domainUserToProto(updatedUser),
	}, nil
}

func (server *UserServer) DeleteUser(ctx context.Context, request *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	userID, parseError := protoUserIDToUUID(request.GetId())
	if parseError != nil {
		return nil, toGRPCError(domain.ErrValidation)
	}

	if deleteError := server.userService.SoftDelete(ctx, userID); deleteError != nil {
		return nil, toGRPCError(deleteError)
	}

	return &userv1.DeleteUserResponse{}, nil
}
