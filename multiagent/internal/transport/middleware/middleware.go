package middleware

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// multiagent is an internal-only service: it sits behind analysis on the
// docker-compose network and is never reached directly by end users. We
// therefore install only Recovery + Logging here, NOT an auth interceptor —
// authentication is the calling service's responsibility (analysis already
// validates JWTs at its own edge before forwarding to multiagent).
//
// If multiagent ever becomes a public-facing service, copy the auth
// interceptor pattern from auth/resume/vacancy and require an auth client.

// UnaryRecoveryInterceptor turns panics into codes.Internal so a single buggy
// handler can't take the whole process down. Always declare it first in the
// chain — must be the outermost wrapper to catch downstream panics.
func UnaryRecoveryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("rpc panic",
				"method", info.FullMethod,
				"panic", r,
				"stack", string(debug.Stack()),
			)
			err = status.Error(codes.Internal, "internal error")
		}
	}()
	return handler(ctx, req)
}

// UnaryLoggingInterceptor records one access log per RPC. Centralising it
// here means handlers only need to emit slog calls for domain forensics that
// can't be derived from method+code.
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
