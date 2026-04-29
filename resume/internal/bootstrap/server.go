package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/artem13815/hr/resume/config"
	"github.com/artem13815/hr/resume/internal/pb/auth_api"
	"github.com/artem13815/hr/resume/internal/pb/resume_api"
	transport_grpc "github.com/artem13815/hr/resume/internal/transport/grpc"
	"github.com/artem13815/hr/resume/internal/transport/middleware"
	"github.com/artem13815/hr/resume/internal/usecase"
)

// gRPC payload overhead (metadata, proto framing) on top of the raw file bytes.
const grpcMessageOverheadBytes = 1 * 1024 * 1024

// maxGRPCRecvMsgBytes caps a single inbound gRPC message. Transport-level
// defense for IngestResumeBatch (a unary RPC that carries inline file_data)
// against memory exhaustion. Streaming RPCs are unaffected because each
// chunk is a separate message.
const maxGRPCRecvMsgBytes = usecase.MaxResumeSizeBytes + grpcMessageOverheadBytes

// gracefulStopTimeout caps how long we wait for in-flight RPCs (notably long
// uploads) to finish on SIGTERM before forcing a Stop().
const gracefulStopTimeout = 15 * time.Second

// AppRun starts the gRPC server and blocks until SIGINT/SIGTERM. On shutdown
// it drains in-flight RPCs (GracefulStop with a Stop fallback) and invokes
// onShutdown hooks in reverse order, LIFO-style, so resources tear down in
// the inverse of their construction order.
func AppRun(api *transport_grpc.ResumeServiceAPI, authClient auth_api.AuthServiceClient, cfg *config.Config, onShutdown ...func()) error {
	lis, err := net.Listen("tcp", cfg.Server.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", cfg.Server.GRPCAddr, err)
	}

	// Recovery is outermost so it catches panics from any later interceptor
	// or handler. Auth runs after logging so a rejected request still emits
	// an access log line with codes.Unauthenticated.
	serverOpts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(maxGRPCRecvMsgBytes),
		grpc.ChainUnaryInterceptor(
			middleware.UnaryRecoveryInterceptor,
			middleware.UnaryLoggingInterceptor,
			middleware.UnaryAuthInterceptor(authClient),
		),
		grpc.ChainStreamInterceptor(
			middleware.StreamRecoveryInterceptor,
			middleware.StreamLoggingInterceptor,
			middleware.StreamAuthInterceptor(authClient),
		),
	}
	if cfg.Server.TLS.Enabled() {
		creds, err := credentials.NewServerTLSFromFile(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile)
		if err != nil {
			return fmt.Errorf("load tls keypair: %w", err)
		}
		serverOpts = append(serverOpts, grpc.Creds(creds))
		slog.Info("gRPC TLS enabled", "cert", cfg.Server.TLS.CertFile)
	}

	s := grpc.NewServer(serverOpts...)
	resume_api.RegisterResumeServiceServer(s, api)

	// Health service for k8s/compose probes. SERVING is set immediately —
	// by the time we register the server, pgxpool and the auth client have
	// already been constructed by bootstrap.Init*.
	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(s, healthSrv)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("gRPC server listening", "addr", cfg.Server.GRPCAddr)
		if err := s.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		runShutdown(onShutdown)
		return err
	case <-ctx.Done():
		slog.Info("shutdown signal received, draining gRPC")
	}

	// Flip health to NOT_SERVING so any LB watching gRPC health stops sending
	// new requests before we begin GracefulStop.
	healthSrv.Shutdown()

	gracefulStop(s)
	runShutdown(onShutdown)
	return nil
}

func gracefulStop(s *grpc.Server) {
	stopped := make(chan struct{})
	go func() {
		s.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		slog.Info("gRPC graceful stop complete")
	case <-time.After(gracefulStopTimeout):
		slog.Warn("gRPC graceful stop timeout, forcing", "timeout", gracefulStopTimeout)
		s.Stop()
	}
}

func runShutdown(onShutdown []func()) {
	// LIFO: resources tear down in reverse construction order.
	for i := len(onShutdown) - 1; i >= 0; i-- {
		func(fn func()) {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("shutdown hook panicked", "panic", r)
				}
			}()
			fn()
		}(onShutdown[i])
	}
}
