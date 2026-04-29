package session_storage

import (
	"github.com/redis/go-redis/v9"
)

type SessionStorage struct {
	rdb *redis.Client
}

func NewSessionStorage(rdb *redis.Client) *SessionStorage {
	return &SessionStorage{
		rdb: rdb,
	}
}

