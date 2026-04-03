package grpc

import (
	"fmt"

	"github.com/gofrs/uuid/v5"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/eminbekov/fiber-v3-template/gen/proto/user/v1"
	"github.com/eminbekov/fiber-v3-template/internal/domain"
)

func domainUserToProto(domainUser *domain.User) *userv1.User {
	if domainUser == nil {
		return nil
	}

	return &userv1.User{
		Id:        domainUser.ID.String(),
		Username:  domainUser.Username,
		FullName:  domainUser.FullName,
		Phone:     domainUser.Phone,
		Status:    domainUser.Status,
		CreatedAt: timestamppb.New(domainUser.CreatedAt),
		UpdatedAt: timestamppb.New(domainUser.UpdatedAt),
	}
}

func protoUserIDToUUID(userID string) (uuid.UUID, error) {
	parsedUUID, parseError := uuid.FromString(userID)
	if parseError != nil {
		return uuid.Nil, fmt.Errorf("protoUserIDToUUID: %w", parseError)
	}
	return parsedUUID, nil
}
