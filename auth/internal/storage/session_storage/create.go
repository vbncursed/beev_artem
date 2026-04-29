package session_storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/artem13815/hr/auth/internal/domain"
)

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
		return err
	}

	key := sessionKey(refreshHash)
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		ttl = time.Second
	}

	return s.rdb.Set(ctx, key, data, ttl).Err()
}
