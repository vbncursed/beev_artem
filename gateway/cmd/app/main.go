package main

import (
	"cmp"
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/artem13815/hr/gateway/config"
	"github.com/artem13815/hr/gateway/internal/bootstrap"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	configPath := cmp.Or(os.Getenv("configPath"), defaultConfigPathByEnv(os.Getenv("APP_ENV")))

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	authClient, authCleanup, err := bootstrap.InitAuthClient(cfg)
	if err != nil {
		return err
	}

	// gwMux dials all four backend services. Its lifetime is bound to the
	// process — when AppRun returns, the process exits and the conns die.
	// We pass a Background ctx because runtime's
	// RegisterXxxHandlerFromEndpoint registers a finalizer on ctx.Done
	// that closes the conn; we don't want that fired during the in-flight
	// drain.
	gwMux, err := bootstrap.InitGatewayMux(context.Background(), cfg)
	if err != nil {
		authCleanup()
		return err
	}

	swaggerSpecs, err := bootstrap.InitSwaggerSpecs()
	if err != nil {
		authCleanup()
		return err
	}

	handler := bootstrap.InitHTTPHandler(authClient, gwMux, swaggerSpecs)

	// Cleanup runs LIFO during shutdown — auth conn closes after the
	// HTTP server stops accepting new requests.
	return bootstrap.AppRun(handler, cfg, authCleanup)
}

func defaultConfigPathByEnv(appEnv string) string {
	switch strings.ToLower(appEnv) {
	case "prod", "production":
		return "./config.docker.prod.yaml"
	default:
		return "./config.docker.dev.yaml"
	}
}
