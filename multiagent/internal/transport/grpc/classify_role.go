package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/artem13815/hr/multiagent/internal/pb/multiagent_api"
	"github.com/artem13815/hr/multiagent/internal/usecase"
)

// ClassifyRole maps a vacancy title+description to one of the registered
// prompt-template roles via the LLM. It is the only place pb and domain
// types meet for this RPC — the usecase stays vendor-agnostic.
//
// Error mapping:
//   - usecase.ErrInvalidArgument   -> codes.InvalidArgument (caller's fault)
//   - usecase.ErrLLMUnavailable    -> codes.Unavailable     (vacancy can fall back)
//   - usecase.ErrLLMInvalidResponse-> codes.Internal        (model misbehaved)
//   - anything else                -> codes.Internal
//
// The Unavailable code matters: vacancy's adapter checks for it explicitly
// and falls back to the keyword detector, so a flaky LLM never blocks
// vacancy create/update.
func (a *MultiAgentServiceAPI) ClassifyRole(ctx context.Context, req *pb.ClassifyRoleRequest) (*pb.ClassifyRoleResponse, error) {
	resp, err := a.svc.ClassifyRole(ctx, pbToDomainClassifyRequest(req))
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidArgument):
			return nil, status.Error(codes.InvalidArgument, "Empty classify request.")
		case errors.Is(err, usecase.ErrLLMUnavailable):
			return nil, status.Error(codes.Unavailable, "LLM unavailable.")
		case errors.Is(err, usecase.ErrLLMInvalidResponse):
			return nil, status.Error(codes.Internal, "LLM returned invalid response.")
		default:
			return nil, status.Error(codes.Internal, "Internal error.")
		}
	}
	return domainToPBClassifyResponse(resp), nil
}
