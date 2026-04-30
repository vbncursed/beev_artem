package grpc

import (
	"github.com/artem13815/hr/vacancy/internal/domain"
	pb_common "github.com/artem13815/hr/vacancy/internal/pb/common"
	pb_models "github.com/artem13815/hr/vacancy/internal/pb/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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
