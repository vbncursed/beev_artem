package session_storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/artem13815/hr/auth/internal/domain"
	"github.com/redis/go-redis/v9"
)

// ConsumeSessionByRefreshHash atomically reads and deletes a session in one Redis
// round-trip via GETDEL. Returns ErrSessionNotFound if the session does not exist
// (or was already consumed by a concurrent caller).
//
// This is the primitive used by Refresh to enforce one-shot refresh tokens: even
// under concurrent replays of the same token, only one caller observes a non-nil
// session.
func (s *SessionStorage) ConsumeSessionByRefreshHash(ctx context.Context, refreshHash []byte) (*domain.Session, error) {
	key := sessionKey(refreshHash)

	data, err := s.rdb.GetDel(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("getdel session: %w", err)
	}

	var sess domain.Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	return &sess, nil
}
