package main

import (
	"cmp"
	"log/slog"
	"os"
	"strings"

	"github.com/artem13815/hr/admin/config"
	"github.com/artem13815/hr/admin/internal/bootstrap"
	"github.com/artem13815/hr/admin/internal/infrastructure/auth_client"

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

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}

	storage, err := bootstrap.InitPGStorage(cfg)
	if err != nil {
		return err
	}

	authClient, authCleanup, err := auth_client.New(cfg)
	if err != nil {
		storage.Close()
		return err
	}

	service := bootstrap.InitAdminService(storage, auth_client.NewRoleUpdater(authClient))
	api := bootstrap.InitAdminServiceAPI(service)

	// Hooks run LIFO during shutdown — close the auth conn before the
	// pgxpool, mirroring construction order.
	return bootstrap.AppRun(api, authClient, cfg, storage.Close, authCleanup)
}

func defaultConfigPathByEnv(appEnv string) string {
	switch strings.ToLower(appEnv) {
	case "prod", "production":
		return "./config.docker.prod.yaml"
	default:
		return "./config.docker.dev.yaml"
	}
}
