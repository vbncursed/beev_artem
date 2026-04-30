package persistence

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// helpers.go now contains only the persistence-layer plumbing the storage
// methods share: ID generation and JSON marshalling helpers. The scoring
// algorithm that used to live here has moved to internal/infrastructure/
// scorer where it can be swapped without touching storage.

func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func marshalJSON(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}
	return b, nil
}

func unmarshalJSON[T any](raw []byte, out *T) {
	_ = json.Unmarshal(raw, out)
}
