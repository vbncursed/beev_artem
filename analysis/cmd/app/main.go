package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/artem13815/hr/analysis/config"
	"github.com/artem13815/hr/analysis/internal/bootstrap"
)

func main() {
	configPath := os.Getenv("configPath")
	if configPath == "" {
		configPath = defaultConfigPathByEnv(os.Getenv("APP_ENV"))
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	storage := bootstrap.InitPGStorage(cfg)
	service := bootstrap.InitAnalysisService(storage, cfg)
	api := bootstrap.InitAnalysisServiceAPI(service)

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
