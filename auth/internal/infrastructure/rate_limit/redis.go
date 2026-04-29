// Package rate_limit is the Redis-backed rate limiter adapter. It exposes a
// concrete *Limiter; callers that need to swap implementations declare their
// own RateLimiter interface where they consume one (interface-on-the-callee).
package rate_limit

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Limiter is a fixed-window counter keyed by `kind:key`. Built around a Lua
// INCR+EXPIRE script that runs atomically on the Redis side so concurrent
// requests can't bump the counter past the threshold.
type Limiter struct {
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

func New(rdb *redis.Client, kind string, limitPerWindow int, window time.Duration) *Limiter {
	return &Limiter{
		rdb:    rdb,
		kind:   kind,
		limit:  int64(limitPerWindow),
		window: window,
	}
}

func (l *Limiter) Allow(ctx context.Context, key string) bool {
	if l == nil || l.limit <= 0 {
		return true
	}
	if l.rdb == nil {
		return false
	}

	if ctx == nil {
		return false
	}
	// Detach from the caller's cancellation but cap our own latency budget —
	// the rate-limit check should not stall the RPC if Redis is degraded.
	execCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 200*time.Millisecond)
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

// keyNormalizer maps every Redis-unfriendly character we want to scrub to a
// single underscore in one pass. Built once at package init so the hot path
// allocates nothing.
var keyNormalizer = strings.NewReplacer(
	"[", "_", "]", "_", ":", "_", " ", "_", "/", "_", "\\", "_",
)

// normalizeRedisKey turns key into a Redis-friendly slug: special characters
// are replaced with `_`, repeated `_` collapse to one, leading and trailing
// underscores are trimmed. Empty or fully-trimmed keys map to "unknown".
func normalizeRedisKey(key string) string {
	if key == "" {
		return "unknown"
	}

	key = keyNormalizer.Replace(key)

	for strings.Contains(key, "__") {
		key = strings.ReplaceAll(key, "__", "_")
	}

	key = strings.Trim(key, "_")

	if key == "" {
		return "unknown"
	}

	return key
}
