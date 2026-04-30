package main

import (
	"cmp"
	"log/slog"
	"os"
	"strings"

	"github.com/artem13815/hr/multiagent/config"
	"github.com/artem13815/hr/multiagent/internal/bootstrap"
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
	service := bootstrap.InitMultiAgentService(storage)
	api := bootstrap.InitMultiAgentServiceAPI(service)

	return bootstrap.AppRun(api, cfg, storage.Close)
}

func defaultConfigPathByEnv(appEnv string) string {
	switch strings.ToLower(appEnv) {
	case "prod", "production":
		return "./config.docker.prod.yaml"
	default:
		return "./config.docker.dev.yaml"
	}
}
