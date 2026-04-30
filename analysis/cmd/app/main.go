package main

import (
	"cmp"
	"log/slog"
	"os"
	"strings"

	"github.com/artem13815/hr/analysis/config"
	"github.com/artem13815/hr/analysis/internal/bootstrap"
	"github.com/artem13815/hr/analysis/internal/infrastructure/auth_client"
	"github.com/artem13815/hr/analysis/internal/infrastructure/multiagent_client"
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

	multiAgentClient, multiAgentCleanup, err := multiagent_client.New(cfg)
	if err != nil {
		storage.Close()
		return err
	}

	service := bootstrap.InitAnalysisService(storage, multiAgentClient)
	api := bootstrap.InitAnalysisServiceAPI(service)

	authClient, authCleanup, err := auth_client.New(cfg)
	if err != nil {
		multiAgentCleanup()
		storage.Close()
		return err
	}

	// Hooks run LIFO during shutdown — close the auth conn, then the
	// multiagent conn, then the pgxpool — mirroring construction order.
	return bootstrap.AppRun(api, authClient, cfg, storage.Close, multiAgentCleanup, authCleanup)
}

func defaultConfigPathByEnv(appEnv string) string {
	switch strings.ToLower(appEnv) {
	case "prod", "production":
		return "./config.docker.prod.yaml"
	default:
		return "./config.docker.dev.yaml"
	}
}
