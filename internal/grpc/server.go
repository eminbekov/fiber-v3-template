package grpc

import (
	"context"
	"log/slog"
	"time"

	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func NewServer(serverOptions ...grpc.ServerOption) *grpc.Server {
	allOptions := make([]grpc.ServerOption, 0, 2+len(serverOptions))
	allOptions = append(allOptions,
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			grpcRecovery.UnaryServerInterceptor(),
			unaryLoggingInterceptor,
		),
	)
	allOptions = append(allOptions, serverOptions...)

	return grpc.NewServer(allOptions...)
}

func unaryLoggingInterceptor(
	ctx context.Context,
	request any,
	unaryServerInfo *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	startedAt := time.Now()
	response, handlerError := handler(ctx, request)
	durationMilliseconds := time.Since(startedAt).Milliseconds()
	if handlerError != nil {
		slog.Error(
			"grpc request failed",
			"method",
			unaryServerInfo.FullMethod,
			"duration_ms",
			durationMilliseconds,
			"error",
			handlerError,
		)
		return nil, handlerError
	}

	slog.Info(
		"grpc request completed",
		"method",
		unaryServerInfo.FullMethod,
		"duration_ms",
		durationMilliseconds,
	)
	return response, nil
}
