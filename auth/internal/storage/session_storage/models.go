package session_storage

import "fmt"

const (
	sessionKeyPrefix = "session:"
)

func sessionKey(refreshHash []byte) string {
	return sessionKeyPrefix + fmt.Sprintf("%x", refreshHash)
}
