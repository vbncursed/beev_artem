package grpc

import (
	"context"

	pb_models "github.com/artem13815/hr/admin/internal/pb/models"
	"github.com/artem13815/hr/admin/internal/transport/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a *AdminServiceAPI) ListUsers(ctx context.Context, _ *pb_models.ListUsersRequest) (*pb_models.ListUsersResponse, error) {
	if _, ok := middleware.Get(ctx); !ok {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	users, err := a.svc.ListUsers(ctx)
	if err != nil {
		return nil, newError(codes.Internal, ErrCodeInternal, "Failed to list users.")
	}

	out := make([]*pb_models.AdminUserView, 0, len(users))
	for _, u := range users {
		out = append(out, &pb_models.AdminUserView{
			Id:                 u.ID,
			Email:              u.Email,
			Role:               u.Role,
			CreatedAt:          timestamppb.New(u.CreatedAt),
			VacanciesOwned:     u.VacanciesOwned,
			CandidatesUploaded: u.CandidatesUploaded,
		})
	}
	return &pb_models.ListUsersResponse{Users: out}, nil
}
