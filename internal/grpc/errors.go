package grpc

import (
	"errors"
	"log/slog"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, "resource not found")
	case errors.Is(err, domain.ErrUnauthorized):
		return status.Error(codes.Unauthenticated, "unauthorized")
	case errors.Is(err, domain.ErrForbidden):
		return status.Error(codes.PermissionDenied, "forbidden")
	case errors.Is(err, domain.ErrConflict):
		return status.Error(codes.AlreadyExists, "resource already exists")
	case errors.Is(err, domain.ErrValidation):
		return status.Error(codes.InvalidArgument, "validation failed")
	default:
		slog.Error("internal grpc error", "error", err)
		return status.Error(codes.Internal, "internal server error")
	}
}
