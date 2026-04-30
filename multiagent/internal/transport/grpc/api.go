package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/artem13815/hr/multiagent/internal/domain"
	pb "github.com/artem13815/hr/multiagent/internal/pb/multiagent_api"
	"github.com/artem13815/hr/multiagent/internal/usecase"
)

// service is the narrow port the transport layer needs. Defined here (the
// consumer side) so tests can inject a fake without depending on the real
// usecase package.
type service interface {
	GenerateDecision(ctx context.Context, req domain.DecisionRequest) (*domain.DecisionResponse, error)
}

type MultiAgentServiceAPI struct {
	pb.UnimplementedMultiAgentServiceServer
	svc service
}

func NewMultiAgentServiceAPI(s service) *MultiAgentServiceAPI {
	return &MultiAgentServiceAPI{svc: s}
}

var _ pb.MultiAgentServiceServer = (*MultiAgentServiceAPI)(nil)

// GenerateDecision is the gRPC entry point. It is the only place in the
// service that touches both pb and domain types — it converts pb -> domain
// for the usecase and domain -> pb for the response.
func (a *MultiAgentServiceAPI) GenerateDecision(ctx context.Context, req *pb.GenerateDecisionRequest) (*pb.GenerateDecisionResponse, error) {
	resp, err := a.svc.GenerateDecision(ctx, pbToDomainRequest(req))
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidArgument) {
			return nil, status.Error(codes.InvalidArgument, "Empty decision request.")
		}
		return nil, status.Error(codes.Internal, "Internal error.")
	}
	return domainToPBResponse(resp), nil
}
