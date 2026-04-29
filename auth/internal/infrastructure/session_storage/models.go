package session_storage

import (
	"encoding/hex"
	"strconv"
)

const (
	sessionKeyPrefix      = "session:"
	userSessionsKeyPrefix = "user_sessions:"
)

func sessionKey(refreshHash []byte) string {
	return sessionKeyPrefix + hex.EncodeToString(refreshHash)
}

// userSessionsKey returns the Redis key of a SET containing every refresh-hash
// (hex-encoded) currently active for a user. The set is the secondary index
// that turns RevokeAllSessionsByUserID from an O(N) keyspace SCAN into
// O(sessions-for-user).
func userSessionsKey(userID uint64) string {
	return userSessionsKeyPrefix + strconv.FormatUint(userID, 10)
}

// refreshHashHex returns the canonical member representation used inside the
// user-sessions set. Hex matches sessionKey's suffix so debugging is symmetric.
func refreshHashHex(refreshHash []byte) string {
	return hex.EncodeToString(refreshHash)
}
