package main

import (
	"cmp"
	"log/slog"
	"os"
	"strings"

	"github.com/artem13815/hr/auth/config"
	"github.com/artem13815/hr/auth/internal/bootstrap"
)

func main() {
	// Resolve config path: explicit configPath env wins; otherwise pick by
	// APP_ENV. cmp.Or picks the first non-zero string.
	configPath := cmp.Or(os.Getenv("configPath"), defaultConfigPathByEnv(os.Getenv("APP_ENV")))

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	authStorage, err := bootstrap.InitPGStorage(cfg)
	if err != nil {
		slog.Error("init pg storage", "err", err)
		os.Exit(1)
	}
	redisClient := bootstrap.InitRedis(cfg)
	sessionStorage := bootstrap.InitSessionStorage(redisClient)

	authService := bootstrap.InitAuthService(authStorage, sessionStorage, cfg)

	loginLimiter, registerLimiter, refreshLimiter := bootstrap.InitRateLimiters(redisClient, cfg)

	authAPI := bootstrap.InitAuthServiceAPI(authService, cfg, loginLimiter, registerLimiter, refreshLimiter)

	if err := bootstrap.AppRun(authAPI, cfg,
		// Cleanups run LIFO in bootstrap.runShutdown — order matches construction.
		authStorage.Close,
		func() {
			if err := redisClient.Close(); err != nil {
				slog.Warn("redis close failed", "err", err)
			}
		},
	); err != nil {
		slog.Error("server exited with error", "err", err)
		os.Exit(1)
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
