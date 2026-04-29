package session_storage

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func (s *SessionStorage) RevokeSessionByRefreshHash(ctx context.Context, refreshHash []byte) error {
	key := sessionKey(refreshHash)

	_, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrSessionNotFound
		}
		return err
	}

	return s.rdb.Del(ctx, key).Err()
}
