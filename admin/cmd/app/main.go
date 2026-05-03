package main

import (
	"cmp"
	"log/slog"
	"os"

	"github.com/artem13815/hr/admin/internal/bootstrap"

	// Register google.rpc errdetails types for grpc-gateway JSON marshaling.
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
	return bootstrap.AppRun(configPath)
}

func defaultConfigPathByEnv(env string) string {
	switch env {
	case "prod", "production":
		return "config.docker.prod.yaml"
	default:
		return "config.docker.dev.yaml"
	}
}
