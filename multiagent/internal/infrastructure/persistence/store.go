package persistence

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/artem13815/hr/multiagent/internal/domain"
)

// StoreDecision persists the (request, response) pair as one audit row. The
// JSON columns intentionally use the domain types (not pb) — this is an
// internal audit log, no external consumer reads the JSON, so we avoid the
// pb dependency in the storage layer.
func (s *MultiAgentStorage) StoreDecision(ctx context.Context, req domain.DecisionRequest, resp *domain.DecisionResponse) error {
	id, err := newID()
	if err != nil {
		return fmt.Errorf("new id: %w", err)
	}
	rawReq, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}
	rawResp, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}
	_, err = s.db.Exec(ctx, `
INSERT INTO multiagent_decisions (id, model, mode, request_json, response_json)
VALUES ($1, $2, $3, $4, $5)
`, id, req.Model, int32(req.Mode), rawReq, rawResp)
	if err != nil {
		return fmt.Errorf("insert decision: %w", err)
	}
	return nil
}

func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
