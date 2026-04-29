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

	"github.com/artem13815/hr/auth/config"
	"github.com/artem13815/hr/auth/internal/infrastructure/jwt"
	"github.com/artem13815/hr/auth/internal/pb/auth_api"
	transport_grpc "github.com/artem13815/hr/auth/internal/transport/grpc"
	"github.com/artem13815/hr/auth/internal/transport/middleware"
)

const gracefulStopTimeout = 15 * time.Second

// AppRun starts the gRPC server and blocks until SIGINT/SIGTERM. On shutdown it
// drains in-flight RPCs (GracefulStop with a timeout fallback to Stop) and
// invokes the supplied cleanup functions in reverse order, LIFO-style.
func AppRun(api *transport_grpc.AuthServiceAPI, validator *jwt.Validator, cfg *config.Config, onShutdown ...func()) error {
	lis, err := net.Listen("tcp", cfg.Server.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", cfg.Server.GRPCAddr, err)
	}

	serverOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			middleware.UnaryRecoveryInterceptor,
			middleware.UnaryLoggingInterceptor,
			middleware.UnaryAuthInterceptor(validator),
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
	auth_api.RegisterAuthServiceServer(s, api)

	// Health service for k8s/compose probes. SERVING is set immediately because
	// by the time we register the server, pgxpool and Redis have already been
	// constructed and validated by bootstrap.Init*.
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
	// LIFO so that resources are torn down in reverse construction order.
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
