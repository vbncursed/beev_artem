package auth_service_api

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisRateLimiter struct {
	rdb    *redis.Client
	kind   string
	limit  int64
	window time.Duration
}

var incrExpireScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
  redis.call("EXPIRE", KEYS[1], ARGV[1])
else
  local ttl = redis.call("TTL", KEYS[1])
  if ttl < 0 then
    redis.call("EXPIRE", KEYS[1], ARGV[1])
  end
end
return current
`)

func newRedisRateLimiter(rdb *redis.Client, kind string, limitPerWindow int, window time.Duration) *redisRateLimiter {
	return &redisRateLimiter{
		rdb:    rdb,
		kind:   kind,
		limit:  int64(limitPerWindow),
		window: window,
	}
}

func NewRedisRateLimiter(rdb *redis.Client, kind string, limitPerWindow int, window time.Duration) RateLimiter {
	return newRedisRateLimiter(rdb, kind, limitPerWindow, window)
}

func (l *redisRateLimiter) Allow(ctx context.Context, key string) bool {
	if l == nil || l.limit <= 0 {
		return true
	}
	if l.rdb == nil {
		return false
	}

	if ctx == nil {
		return false
	}
	execCtx := context.WithoutCancel(ctx)

	var cancel context.CancelFunc
	execCtx, cancel = context.WithTimeout(execCtx, 200*time.Millisecond)
	defer cancel()

	normalizedKey := normalizeRedisKey(key)
	redisKey := fmt.Sprintf("rl:%s:%s", l.kind, normalizedKey)
	ttlSeconds := int64(l.window.Seconds())
	if ttlSeconds <= 0 {
		ttlSeconds = 60
	}

	n, err := incrExpireScript.Run(execCtx, l.rdb, []string{redisKey}, ttlSeconds).Int64()
	if err != nil {
		slog.Warn("rate limit redis error", "err", err, "key", redisKey, "kind", l.kind)
		return false
	}

	allowed := n <= l.limit
	if !allowed {
		slog.Info("rate limit exceeded", "key", redisKey, "count", n, "limit", l.limit, "kind", l.kind)
	}
	return allowed
}

// normalizeRedisKey нормализует ключ для использования в Redis
// Заменяет специальные символы на безопасные альтернативы
func normalizeRedisKey(key string) string {
	if key == "" {
		return "unknown"
	}

	key = strings.ReplaceAll(key, "[", "_")
	key = strings.ReplaceAll(key, "]", "_")
	key = strings.ReplaceAll(key, ":", "_")
	key = strings.ReplaceAll(key, " ", "_")
	key = strings.ReplaceAll(key, "/", "_")
	key = strings.ReplaceAll(key, "\\", "_")

	for strings.Contains(key, "__") {
		key = strings.ReplaceAll(key, "__", "_")
	}

	key = strings.Trim(key, "_")

	if key == "" {
		return "unknown"
	}

	return key
}
