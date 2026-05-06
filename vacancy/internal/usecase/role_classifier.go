package usecase

import "context"

// RoleClassifier infers the multiagent prompt role from a vacancy's title
// and description. The implementation lives in
// internal/infrastructure/multiagent_client and proxies to the multiagent
// service over gRPC; tests inject a mock.
//
// Contract:
//   - Returned role is opaque to the usecase — it's stored verbatim on the
//     vacancy and forwarded to multiagent.GenerateDecision.role downstream.
//     Validation against the prompt set lives on the multiagent side.
//   - Returns ErrLLMUnavailable on any provider/transport failure so the
//     usecase can fall back to the deterministic keyword detector.
//   - Any other error is treated the same way (fall back) — the usecase
//     never lets a classifier failure block CRUD.
//
// Why a port and not a direct gRPC call from the service: keeps the
// usecase test hermetic (no network, no protobuf), and lets the deterministic
// fallback live entirely in pure code.
type RoleClassifier interface {
	Classify(ctx context.Context, title, description string) (string, error)
}
