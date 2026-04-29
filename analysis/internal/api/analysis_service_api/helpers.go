package analysis_service_api

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/artem13815/hr/analysis/internal/domain"
	pb_common "github.com/artem13815/hr/analysis/internal/pb/common"
	pb_models "github.com/artem13815/hr/analysis/internal/pb/models"
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
			if err != nil || userID == 0 {
				return nil, fmt.Errorf("invalid user id")
			}
			return &userContext{UserID: userID, Role: role, IsAdmin: role == "admin"}, nil
		}
	}

	return nil, fmt.Errorf("user id not found")
}

func toPBStatus(s string) pb_models.AnalysisStatus {
	switch s {
	case domain.StatusQueued:
		return pb_models.AnalysisStatus_ANALYSIS_STATUS_QUEUED
	case domain.StatusRunning:
		return pb_models.AnalysisStatus_ANALYSIS_STATUS_RUNNING
	case domain.StatusDone:
		return pb_models.AnalysisStatus_ANALYSIS_STATUS_DONE
	case domain.StatusFailed:
		return pb_models.AnalysisStatus_ANALYSIS_STATUS_FAILED
	default:
		return pb_models.AnalysisStatus_ANALYSIS_STATUS_UNSPECIFIED
	}
}

func toPBOrderDesc(order pb_common.SortOrder) bool {
	return order != pb_common.SortOrder_SORT_ORDER_ASC
}

func toPBPage(limit, offset uint32, total uint64) *pb_common.PageResponse {
	return &pb_common.PageResponse{Limit: limit, Offset: offset, Total: total}
}

func toPBAnalysis(a domain.Analysis) *pb_models.Analysis {
	agentResults := make([]*pb_models.AgentResult, 0, len(a.AI.AgentResults))
	for _, item := range a.AI.AgentResults {
		agentResults = append(agentResults, &pb_models.AgentResult{
			AgentName:      item.AgentName,
			Summary:        item.Summary,
			StructuredJson: item.StructuredJSON,
			Confidence:     item.Confidence,
		})
	}

	return &pb_models.Analysis{
		Id:             a.ID,
		VacancyId:      a.VacancyID,
		CandidateId:    a.CandidateID,
		ResumeId:       a.ResumeID,
		VacancyVersion: a.VacancyVersion,
		Status:         toPBStatus(a.Status),
		MatchScore:     a.MatchScore,
		Profile: &pb_models.CandidateProfile{
			Skills:          a.Profile.Skills,
			YearsExperience: a.Profile.YearsExperience,
			Positions:       a.Profile.Positions,
			Technologies:    a.Profile.Technologies,
			Education:       a.Profile.Education,
			Summary:         a.Profile.Summary,
		},
		Breakdown: &pb_models.ScoreBreakdown{
			MatchedSkills:   a.Breakdown.MatchedSkills,
			MissingSkills:   a.Breakdown.MissingSkills,
			ExtraSkills:     a.Breakdown.ExtraSkills,
			BaseScore:       a.Breakdown.BaseScore,
			MustHavePenalty: a.Breakdown.MustHavePenalty,
			NiceToHaveBonus: a.Breakdown.NiceToHaveBonus,
			Explanation:     a.Breakdown.Explanation,
		},
		Ai: &pb_models.AIDecision{
			HrRecommendation:  a.AI.HRRecommendation,
			Confidence:        a.AI.Confidence,
			HrRationale:       a.AI.HRRationale,
			CandidateFeedback: a.AI.CandidateFeedback,
			SoftSkillsNotes:   a.AI.SoftSkillsNotes,
			AgentResults:      agentResults,
			RawTrace:          a.AI.RawTrace,
		},
		ErrorMessage: a.ErrorMessage,
		CreatedAt:    timestamppb.New(a.CreatedAt),
		UpdatedAt:    timestamppb.New(a.UpdatedAt),
	}
}
