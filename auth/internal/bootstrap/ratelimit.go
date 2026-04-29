package bootstrap

import (
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/artem13815/hr/auth/config"
	"github.com/artem13815/hr/auth/internal/infrastructure/rate_limit"
)

func InitRateLimiters(rdb *redis.Client, cfg *config.Config) (loginLimiter, registerLimiter, refreshLimiter *rate_limit.Limiter) {
	window := time.Minute
	return rate_limit.New(rdb, "login", cfg.Auth.RateLimitLoginPerMinute, window),
		rate_limit.New(rdb, "register", cfg.Auth.RateLimitRegisterPerMinute, window),
		rate_limit.New(rdb, "refresh", cfg.Auth.RateLimitRefreshPerMinute, window)
}
