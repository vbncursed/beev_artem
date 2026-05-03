package middleware

import (
	"context"
	"log/slog"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryRecoveryInterceptor turns panics into codes.Internal so a single
// buggy handler can't take the whole process down. Always declared first
// in the chain — must be the outermost wrapper.
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
