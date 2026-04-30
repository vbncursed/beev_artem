package grpc

import (
	"context"

	pb "github.com/artem13815/hr/multiagent/internal/pb/multiagent_api"
)

type service interface {
	GenerateDecision(ctx context.Context, req *pb.GenerateDecisionRequest) (*pb.GenerateDecisionResponse, error)
}

type MultiAgentServiceAPI struct {
	pb.UnimplementedMultiAgentServiceServer
	svc service
}

func NewMultiAgentServiceAPI(s service) *MultiAgentServiceAPI {
	return &MultiAgentServiceAPI{svc: s}
}

func (a *MultiAgentServiceAPI) GenerateDecision(ctx context.Context, req *pb.GenerateDecisionRequest) (*pb.GenerateDecisionResponse, error) {
	return a.svc.GenerateDecision(ctx, req)
}
