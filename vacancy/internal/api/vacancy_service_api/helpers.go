package vacancy_service_api

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/artem13815/hr/vacancy/internal/domain"
	pb_common "github.com/artem13815/hr/vacancy/internal/pb/common"
	pb_models "github.com/artem13815/hr/vacancy/internal/pb/models"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type userContext struct {
	UserID  uint64
	Role    string
	IsAdmin bool
}

func getUserContext(ctx context.Context) (*userContext, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("metadata not found")
	}

	role := "user"
	if values := md.Get("x-user-role"); len(values) > 0 && strings.TrimSpace(values[0]) != "" {
		role = strings.ToLower(strings.TrimSpace(values[0]))
	}

	keys := []string{"x-user-id", "user-id"}
	for _, key := range keys {
		if values := md.Get(key); len(values) > 0 {
			userID, err := strconv.ParseUint(values[0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid user id")
			}
			if userID == 0 {
				return nil, fmt.Errorf("invalid user id")
			}
			return &userContext{
				UserID:  userID,
				Role:    role,
				IsAdmin: role == "admin",
			}, nil
		}
	}

	return nil, fmt.Errorf("user id not found")
}

func toPBSkills(skills []domain.SkillWeight) []*pb_models.SkillWeight {
	out := make([]*pb_models.SkillWeight, 0, len(skills))
	for _, s := range skills {
		out = append(out, &pb_models.SkillWeight{
			Name:       s.Name,
			Weight:     s.Weight,
			MustHave:   s.MustHave,
			NiceToHave: s.NiceToHave,
		})
	}
	return out
}

func fromPBSkills(skills []*pb_models.SkillWeight) []domain.SkillWeight {
	out := make([]domain.SkillWeight, 0, len(skills))
	for _, s := range skills {
		out = append(out, domain.SkillWeight{
			Name:       s.GetName(),
			Weight:     s.GetWeight(),
			MustHave:   s.GetMustHave(),
			NiceToHave: s.GetNiceToHave(),
		})
	}
	return out
}

func toPBVacancy(v domain.Vacancy) *pb_models.Vacancy {
	return &pb_models.Vacancy{
		Id:          v.ID,
		OwnerUserId: v.OwnerUserID,
		Title:       v.Title,
		Description: v.Description,
		Skills:      toPBSkills(v.Skills),
		Status:      toPBStatus(v.Status),
		Version:     v.Version,
		CreatedAt:   timestamppb.New(v.CreatedAt),
		UpdatedAt:   timestamppb.New(v.UpdatedAt),
	}
}

func toPBStatus(status string) pb_models.VacancyStatus {
	switch status {
	case domain.StatusArchived:
		return pb_models.VacancyStatus_VACANCY_STATUS_ARCHIVED
	default:
		return pb_models.VacancyStatus_VACANCY_STATUS_ACTIVE
	}
}

func toPBPage(limit, offset uint32, total uint64) *pb_common.PageResponse {
	return &pb_common.PageResponse{Limit: limit, Offset: offset, Total: total}
}
