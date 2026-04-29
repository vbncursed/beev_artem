package multiagent_storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"

	pb "github.com/artem13815/hr/multiagent/internal/pb/multiagent_api"
)

func (s *MultiAgentStorage) StoreDecision(ctx context.Context, req *pb.GenerateDecisionRequest, resp *pb.GenerateDecisionResponse) error {
	id, err := newID()
	if err != nil {
		return err
	}
	rawReq, err := json.Marshal(req)
	if err != nil {
		return err
	}
	rawResp, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, `
INSERT INTO multiagent_decisions (id, model, mode, request_json, response_json)
VALUES ($1, $2, $3, $4, $5)
`, id, req.GetModel(), int32(req.GetMode()), rawReq, rawResp)
	return err
}

func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
