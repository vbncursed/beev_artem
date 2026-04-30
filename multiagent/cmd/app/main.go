package main

import (
	"cmp"
	"fmt"
	"os"
	"strings"

	"github.com/artem13815/hr/multiagent/config"
	"github.com/artem13815/hr/multiagent/internal/bootstrap"
)

func main() {
	configPath := cmp.Or(os.Getenv("configPath"), defaultConfigPathByEnv(os.Getenv("APP_ENV")))

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	storage := bootstrap.InitPGStorage(cfg)
	service := bootstrap.InitMultiAgentService(storage)
	api := bootstrap.InitMultiAgentServiceAPI(service)
	bootstrap.AppRun(api, cfg)
}

func defaultConfigPathByEnv(appEnv string) string {
	switch strings.ToLower(appEnv) {
	case "prod", "production":
		return "./config.docker.prod.yaml"
	default:
		return "./config.docker.dev.yaml"
	}
}
