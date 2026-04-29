package bootstrap

import (
	"github.com/artem13815/hr/auth/internal/infrastructure/session_storage"
	"github.com/redis/go-redis/v9"
)

func InitSessionStorage(redisClient *redis.Client) *session_storage.SessionStorage {
	return session_storage.NewSessionStorage(redisClient)
}
