package session_storage

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/artem13815/hr/auth/internal/domain"
	"github.com/redis/go-redis/v9"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

func (s *SessionStorage) GetSessionByRefreshHash(ctx context.Context, refreshHash []byte) (*domain.Session, error) {
	key := sessionKey(refreshHash)
	data, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	var sess domain.Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}

	return &sess, nil
}
