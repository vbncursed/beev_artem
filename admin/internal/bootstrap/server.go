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

	"github.com/artem13815/hr/admin/config"
	"github.com/artem13815/hr/admin/internal/infrastructure/persistence"
	"github.com/artem13815/hr/admin/internal/pb/admin_api"
	"github.com/artem13815/hr/admin/internal/transport/grpc/transport"
	"github.com/artem13815/hr/admin/internal/transport/middleware"
	"github.com/artem13815/hr/admin/internal/usecase"
)

// AppRun is the lifecycle: load config, init storage + auth client,
// register the gRPC server, install LIFO cleanup hooks, listen for
// SIGINT/SIGTERM, GracefulStop → fallback Stop after 15s.
func AppRun(configPath string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	var cleanups []func()
	defer func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}()

	storage, err := InitPGStorage(cfg)
	if err != nil {
		return err
	}
	cleanups = append(cleanups, storage.Close)

	authCli, authConn, err := InitAuthClient(cfg)
	if err != nil {
		return err
	}
	cleanups = append(cleanups, func() { _ = authConn.Close() })

	svc := usecase.NewAdminService(storage, authCli)
	api := transport.NewAdminServiceAPI(svc)

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.UnaryRecoveryInterceptor,
			middleware.UnaryLoggingInterceptor,
			middleware.UnaryAuthInterceptor(authCli.Raw()),
		),
	)
	admin_api.RegisterAdminServiceServer(srv, api)

	lis, err := net.Listen("tcp", cfg.Server.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", cfg.Server.GRPCAddr, err)
	}

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("admin gRPC up", "addr", cfg.Server.GRPCAddr)
		if err := srv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			serveErr <- err
		}
		close(serveErr)
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}

	stopped := make(chan struct{})
	go func() { srv.GracefulStop(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(15 * time.Second):
		slog.Warn("graceful stop timed out, forcing")
		srv.Stop()
	}

	// Suppress unused import warning for persistence in tests-less builds.
	_ = (*persistence.StatsStorage)(nil)
	return nil
}
