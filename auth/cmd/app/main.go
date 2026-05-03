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

	authStorage, err := bootstrap.InitPGStorage(cfg)
	if err != nil {
		return err
	}

	redisClient := bootstrap.InitRedis(cfg)
	sessionStorage := bootstrap.InitSessionStorage(redisClient)

	authService := bootstrap.InitAuthService(authStorage, sessionStorage, cfg)

	loginLimiter, registerLimiter, refreshLimiter := bootstrap.InitRateLimiters(redisClient, cfg)

	jwtValidator := bootstrap.InitJWTValidator(cfg)
	authAPI := bootstrap.InitAuthServiceAPI(authService, jwtValidator, loginLimiter, registerLimiter, refreshLimiter)

	// Cleanups run LIFO during shutdown — close redis after the pgxpool,
	// mirroring construction order.
	return bootstrap.AppRun(authAPI, jwtValidator, cfg,
		authStorage.Close,
		func() {
			if err := redisClient.Close(); err != nil {
				slog.Warn("redis close failed", "err", err)
			}
		},
	)
}

func defaultConfigPathByEnv(appEnv string) string {
	switch strings.ToLower(appEnv) {
	case "prod", "production":
		return "./config.docker.prod.yaml"
	default:
		return "./config.docker.dev.yaml"
	}
}
