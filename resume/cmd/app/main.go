package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/artem13815/hr/resume/config"
	"github.com/artem13815/hr/resume/internal/bootstrap"
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
	service := bootstrap.InitResumeService(storage)
	api := bootstrap.InitResumeServiceAPI(service)

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
