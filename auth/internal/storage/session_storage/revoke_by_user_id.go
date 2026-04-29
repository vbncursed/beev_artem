package session_storage

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RevokeAllSessionsByUserID deletes every active session for the user via the
// user_sessions:<user_id> secondary index. This replaces the previous
// implementation that did SCAN over the whole `session:*` keyspace + GET +
// JSON decode for every key — pathological under load and a hotspot risk for
// the Redis event loop.
//
// Stale members (sessions that already expired through Redis TTL) are tolerated:
// DEL on a missing key is a no-op, so we just batch-delete everything the
// index claims and then drop the index itself.
func (s *SessionStorage) RevokeAllSessionsByUserID(ctx context.Context, userID uint64) error {
	indexKey := userSessionsKey(userID)

	hashes, err := s.rdb.SMembers(ctx, indexKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return fmt.Errorf("smembers user sessions: %w", err)
	}
	if len(hashes) == 0 {
		// Nothing to enumerate; still drop the (empty) index for cleanliness.
		s.rdb.Del(ctx, indexKey)
		return nil
	}

	pipe := s.rdb.Pipeline()
	for _, hexHash := range hashes {
		raw, err := hex.DecodeString(hexHash)
		if err != nil {
			// Skip corrupted index entries — defensive; every member is
			// written by refreshHashHex so this should never happen.
			continue
		}
		pipe.Del(ctx, sessionKey(raw))
	}
	pipe.Del(ctx, indexKey)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("pipeline revoke all: %w", err)
	}

	return nil
}
