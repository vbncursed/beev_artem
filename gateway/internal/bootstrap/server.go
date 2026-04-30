package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/artem13815/hr/gateway/config"
)

// Lifecycle timeouts. ReadHeaderTimeout caps the slow-loris exposure
// pre-headers; ReadTimeout is the full request read budget; WriteTimeout
// is the response budget — generous for the merged swagger spec, which
// is the largest payload the gateway emits. IdleTimeout reaps lingering
// keepalive conns. shutdownTimeout caps the in-flight drain.
const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 60 * time.Second
	idleTimeout       = 120 * time.Second
	shutdownTimeout   = 15 * time.Second
)

// AppRun owns the gateway lifecycle: it spins up the HTTP server in a
// goroutine, blocks on SIGINT/SIGTERM, then triggers srv.Shutdown(ctx)
// to drain in-flight requests. onShutdown hooks run LIFO so resources
// tear down in the inverse of construction order — auth-client conn
// closes after the server stops accepting new requests.
func AppRun(handler http.Handler, cfg *config.Config, onShutdown ...func()) error {
	srv := &http.Server{
		Addr:              cfg.Server.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("gateway server listening",
			"http_addr", cfg.Server.HTTPAddr,
			"auth_grpc_addr", cfg.Auth.GRPCAddr,
			"vacancy_grpc_addr", cfg.Vacancy.GRPCAddr,
			"resume_grpc_addr", cfg.Resume.GRPCAddr,
			"analysis_grpc_addr", cfg.Analysis.GRPCAddr,
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- fmt.Errorf("gateway http server failed: %w", err)
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		runShutdown(onShutdown)
		return err
	case <-ctx.Done():
		slog.Info("shutdown signal received, draining gateway")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Warn("gateway shutdown failed", "err", err)
	} else {
		slog.Info("gateway shutdown complete")
	}

	runShutdown(onShutdown)
	return nil
}

func runShutdown(onShutdown []func()) {
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
