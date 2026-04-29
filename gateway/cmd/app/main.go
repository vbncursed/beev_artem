package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/artem13815/hr/gateway/config"
	"github.com/artem13815/hr/gateway/internal/bootstrap"
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

	if err := bootstrap.AppRun(cfg); err != nil {
		panic(fmt.Sprintf("failed to run gateway app: %v", err))
	}
}

func defaultConfigPathByEnv(appEnv string) string {
	switch strings.ToLower(appEnv) {
	case "prod", "production":
		return "./config.docker.prod.yaml"
	default:
		return "./config.docker.dev.yaml"
	}
}
