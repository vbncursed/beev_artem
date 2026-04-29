package bootstrap

import (
	"time"

	"github.com/artem13815/hr/auth/config"
	"github.com/artem13815/hr/auth/internal/api/auth_service_api"
	"github.com/redis/go-redis/v9"
)

func InitRateLimiters(rdb *redis.Client, cfg *config.Config) (loginLimiter, registerLimiter, refreshLimiter auth_service_api.RateLimiter) {
	window := time.Minute

	loginLimiter = auth_service_api.NewRedisRateLimiter(
		rdb,
		"login",
		cfg.Auth.RateLimitLoginPerMinute,
		window,
	)

	registerLimiter = auth_service_api.NewRedisRateLimiter(
		rdb,
		"register",
		cfg.Auth.RateLimitRegisterPerMinute,
		window,
	)

	refreshLimiter = auth_service_api.NewRedisRateLimiter(
		rdb,
		"refresh",
		cfg.Auth.RateLimitRefreshPerMinute,
		window,
	)

	return loginLimiter, registerLimiter, refreshLimiter
}
