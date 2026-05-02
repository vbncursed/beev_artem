package middleware

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryLoggingInterceptor records one access log per RPC. Centralising it
// here means handlers only need to emit slog calls for domain forensics that
// can't be derived from method+code (e.g. user-mismatch on a delete).
func UnaryLoggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	slog.Info("rpc",
		"method", info.FullMethod,
		"code", status.Code(err).String(),
		"duration", time.Since(start),
	)
	return resp, err
}

func StreamLoggingInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	start := time.Now()
	err := handler(srv, ss)
	slog.Info("rpc",
		"method", info.FullMethod,
		"code", status.Code(err).String(),
		"duration", time.Since(start),
	)
	return err
}
