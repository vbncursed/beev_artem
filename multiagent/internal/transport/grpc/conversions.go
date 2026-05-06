package grpc

import (
	"github.com/artem13815/hr/multiagent/internal/domain"
	pb "github.com/artem13815/hr/multiagent/internal/pb/multiagent_api"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// pbToDomainRequest copies the wire request into the domain shape. Fields
// map 1:1 — this is a thin layer whose only job is to keep the pb type out
// of the usecase.
func pbToDomainRequest(req *pb.GenerateDecisionRequest) domain.DecisionRequest {
	if req == nil {
		return domain.DecisionRequest{}
	}
	return domain.DecisionRequest{
		Model:             req.GetModel(),
		Mode:              domain.AgentMode(req.GetMode()),
		Role:              req.GetRole(),
		CandidateSkills:   req.GetCandidateSkills(),
		MissingSkills:     req.GetMissingSkills(),
		CandidateSummary:  req.GetCandidateSummary(),
		ScoreExplanation:  req.GetScoreExplanation(),
		MatchScore:        req.GetMatchScore(),
		VacancyMustHave:   req.GetVacancyMustHave(),
		VacancyNiceToHave: req.GetVacancyNiceToHave(),
		ResumeText:        req.GetResumeText(),
	}
}

// domainToPBResponse maps the domain decision back to the wire shape so the
// gRPC layer can return it. nil input would never happen on the success
// path, but we guard so a buggy caller can't panic the server.
func domainToPBResponse(resp *domain.DecisionResponse) *pb.GenerateDecisionResponse {
	if resp == nil {
		return &pb.GenerateDecisionResponse{}
	}
	agents := make([]*pb.AgentResult, 0, len(resp.AgentResults))
	for _, a := range resp.AgentResults {
		agents = append(agents, &pb.AgentResult{
			AgentName:      a.AgentName,
			Summary:        a.Summary,
			StructuredJson: a.StructuredJSON,
			Confidence:     a.Confidence,
		})
	}
	return &pb.GenerateDecisionResponse{
		HrRecommendation:  resp.HRRecommendation,
		Confidence:        resp.Confidence,
		HrRationale:       resp.HRRationale,
		CandidateFeedback: resp.CandidateFeedback,
		SoftSkillsNotes:   resp.SoftSkillsNotes,
		AgentResults:      agents,
		RawTrace:          resp.RawTrace,
		CreatedAt:         timestamppb.New(resp.CreatedAt),
		YearsExperience:   resp.YearsExperience,
		CandidateSummary:  resp.CandidateSummary,
	}
}

func pbToDomainClassifyRequest(req *pb.ClassifyRoleRequest) domain.RoleClassifyRequest {
	if req == nil {
		return domain.RoleClassifyRequest{}
	}
	return domain.RoleClassifyRequest{
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
	}
}

func domainToPBClassifyResponse(resp *domain.RoleClassifyResponse) *pb.ClassifyRoleResponse {
	if resp == nil {
		return &pb.ClassifyRoleResponse{}
	}
	return &pb.ClassifyRoleResponse{Role: resp.Role}
}
