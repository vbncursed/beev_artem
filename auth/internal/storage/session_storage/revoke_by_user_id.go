package session_storage

import (
	"context"
	"encoding/json"

	"github.com/artem13815/hr/auth/internal/domain"
	"github.com/redis/go-redis/v9"
)

func (s *SessionStorage) RevokeAllSessionsByUserID(ctx context.Context, userID uint64) error {
	pattern := sessionKeyPrefix + "*"
	var cursor uint64 = 0

	for {
		var keys []string
		var err error
		keys, cursor, err = s.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		for _, key := range keys {
			data, err := s.rdb.Get(ctx, key).Bytes()
			if err != nil {
				if err == redis.Nil {
					continue
				}
				return err
			}

			var sess domain.Session
			if err := json.Unmarshal(data, &sess); err != nil {
				continue
			}

			if sess.UserID == userID {
				if err := s.rdb.Del(ctx, key).Err(); err != nil {
					return err
				}
			}
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}
