package session_storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/artem13815/hr/auth/internal/domain"
)

// CreateSession persists the session JSON under session:<hash> (with TTL) and
// adds <hash> to user_sessions:<user_id> in a single Redis pipeline. The set
// is the secondary index used by RevokeAllSessionsByUserID — without it that
// operation would have to SCAN the whole keyspace.
//
// If the SET succeeds but SADD fails the worst-case outcome is an "orphan"
// session that won't be enumerated by RevokeAll; it will still expire on its
// own TTL, so we don't compensate.
func (s *SessionStorage) CreateSession(ctx context.Context, userID uint64, refreshHash []byte, expiresAt time.Time, userAgent, ip string) error {
	sess := domain.Session{
		UserID:      userID,
		RefreshHash: refreshHash,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
		UserAgent:   userAgent,
		IP:          ip,
	}

	data, err := json.Marshal(sess)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		ttl = time.Second
	}

	pipe := s.rdb.Pipeline()
	pipe.Set(ctx, sessionKey(refreshHash), data, ttl)
	pipe.SAdd(ctx, userSessionsKey(userID), refreshHashHex(refreshHash))
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	return nil
}
