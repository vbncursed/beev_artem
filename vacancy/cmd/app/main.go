package main

import (
	"cmp"
	"log/slog"
	"os"
	"strings"

	"github.com/artem13815/hr/vacancy/config"
	"github.com/artem13815/hr/vacancy/internal/bootstrap"
	"github.com/artem13815/hr/vacancy/internal/infrastructure/auth_client"
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

	maClient, maCleanup, err := bootstrap.InitMultiAgentClient(cfg)
	if err != nil {
		storage.Close()
		return err
	}

	service := bootstrap.InitVacancyService(storage, maClient)
	api := bootstrap.InitVacancyServiceAPI(service)

	authClient, authCleanup, err := auth_client.New(cfg)
	if err != nil {
		maCleanup()
		storage.Close()
		return err
	}

	// Hooks run LIFO during shutdown — close auth and multiagent conns
	// before the pgxpool, mirroring construction order. multiagent goes
	// before auth in this list because it was constructed earlier;
	// runShutdown reverses the slice so auth tears down first.
	return bootstrap.AppRun(api, authClient, cfg, storage.Close, maCleanup, authCleanup)
}

func defaultConfigPathByEnv(appEnv string) string {
	switch strings.ToLower(appEnv) {
	case "prod", "production":
		return "./config.docker.prod.yaml"
	default:
		return "./config.docker.dev.yaml"
	}
}
