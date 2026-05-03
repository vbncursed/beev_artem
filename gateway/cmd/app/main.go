package main

import (
	"cmp"
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/artem13815/hr/gateway/config"
	"github.com/artem13815/hr/gateway/internal/bootstrap"

	// Blank-import errdetails so its protobuf types
	// (google.rpc.ErrorInfo, BadRequest, RetryInfo, ...) register in
	// protoregistry.GlobalTypes at init time. Without this import,
	// grpc-gateway fails to marshal `google.protobuf.Any` payloads
	// returned by backend services with:
	//   "unable to resolve type.googleapis.com/google.rpc.ErrorInfo: not found"
	// because the gateway binary itself never references the type.
	_ "google.golang.org/genproto/googleapis/rpc/errdetails"
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

	handler := bootstrap.InitHTTPHandler(authClient, gwMux, swaggerSpecs, cfg.Server.CORS.AllowedOrigins)

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
