package grpc

import (
	"errors"
	"fmt"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/eminbekov/fiber-v3-template/internal/domain"
)

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return fmt.Errorf("toGRPCError: %w", status.Error(codes.NotFound, "resource not found"))
	case errors.Is(err, domain.ErrUnauthorized):
		return fmt.Errorf("toGRPCError: %w", status.Error(codes.Unauthenticated, "unauthorized"))
	case errors.Is(err, domain.ErrForbidden):
		return fmt.Errorf("toGRPCError: %w", status.Error(codes.PermissionDenied, "forbidden"))
	case errors.Is(err, domain.ErrConflict):
		return fmt.Errorf("toGRPCError: %w", status.Error(codes.AlreadyExists, "resource already exists"))
	case errors.Is(err, domain.ErrValidation):
		return fmt.Errorf("toGRPCError: %w", status.Error(codes.InvalidArgument, "validation failed"))
	default:
		slog.Error("internal grpc error", "error", err)
		return fmt.Errorf("toGRPCError: %w", status.Error(codes.Internal, "internal server error"))
	}
}
