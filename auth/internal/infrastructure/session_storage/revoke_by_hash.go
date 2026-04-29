package session_storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/artem13815/hr/auth/internal/domain"
	"github.com/redis/go-redis/v9"
)

// RevokeSessionByRefreshHash atomically deletes a single session via GETDEL
// and removes its hash from the user_sessions:<user_id> index. Returns
// ErrSessionNotFound when the session is already gone.
//
// Note: GETDEL replaces the previous non-atomic GET+DEL. Logout still calls
// GetSessionByRefreshHash beforehand to verify ownership (uniform-error
// semantics from H4) — that's not a TOCTOU concern because Logout's race window
// is benign (worst case: two competing logouts of the same session, both
// succeed at deleting it, indistinguishable from one).
func (s *SessionStorage) RevokeSessionByRefreshHash(ctx context.Context, refreshHash []byte) error {
	data, err := s.rdb.GetDel(ctx, sessionKey(refreshHash)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrSessionNotFound
		}
		return fmt.Errorf("getdel session: %w", err)
	}

	var sess domain.Session
	if err := json.Unmarshal(data, &sess); err == nil {
		// Best-effort index cleanup.
		s.rdb.SRem(ctx, userSessionsKey(sess.UserID), refreshHashHex(refreshHash))
	}

	return nil
}
