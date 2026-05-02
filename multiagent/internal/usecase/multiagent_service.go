package usecase

//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i DecisionStorage,LLM,PromptStore -o ./mocks -s _mock.go -g

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/artem13815/hr/multiagent/internal/domain"
)

const completionTemperature = 0.3
const completionMaxTokens = 1500

const languageDirective = `

ВАЖНО — язык ответа:
- Все текстовые поля JSON-ответа (hr_rationale, candidate_feedback, soft_skills_notes, agent_results[*].summary) пиши СТРОГО на русском языке.
- Не используй английские слова и фразы в этих полях, кроме имён технологий, продуктов и аббревиатур (например: Go, PostgreSQL, gRPC, AWS, ИВЛ, МКБ-10).
- Поле hr_recommendation — это enum, оставь одно из значений как есть: "hire", "maybe", "no". Не переводи.
- Если в резюме встречаются английские формулировки — пересказывай их по-русски, не цитируй дословно.`

type DecisionStorage interface {
	StoreDecision(ctx context.Context, req domain.DecisionRequest, resp *domain.DecisionResponse) error
}

var ErrInvalidArgument = errors.New("invalid argument")

type MultiAgentService struct {
	storage DecisionStorage
	llm     LLM
	prompts PromptStore
}

func NewMultiAgentService(storage DecisionStorage, llm LLM, prompts PromptStore) *MultiAgentService {
	return &MultiAgentService{storage: storage, llm: llm, prompts: prompts}
}

func (s *MultiAgentService) GenerateDecision(ctx context.Context, req domain.DecisionRequest) (*domain.DecisionResponse, error) {
	if isEmptyRequest(req) {
		return nil, ErrInvalidArgument
	}

	instructions := s.prompts.Get(req.Role) + languageDirective
	input := buildInput(req)

	completion, err := s.llm.Complete(ctx, domain.CompletionRequest{
		Instructions:    instructions,
		Input:           input,
		Temperature:     completionTemperature,
		MaxOutputTokens: completionMaxTokens,
	})
	if err != nil {
		return nil, err
	}

	resp, err := parseDecision(completion)
	if err != nil {
		return nil, err
	}
	resp.RawTrace = "yandex-llm-v1"
	resp.CreatedAt = time.Now()

	if err := s.storage.StoreDecision(ctx, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func isEmptyRequest(req domain.DecisionRequest) bool {
	return req.Model == "" &&
		req.MatchScore == 0 &&
		len(req.CandidateSkills) == 0 &&
		len(req.MissingSkills) == 0 &&
		req.ResumeText == ""
}

func buildInput(req domain.DecisionRequest) string {
	payload := map[string]any{
		"vacancy": map[string]any{
			"role":         req.Role,
			"must_have":    req.VacancyMustHave,
			"nice_to_have": req.VacancyNiceToHave,
		},
		"candidate": map[string]any{
			"skills":      req.CandidateSkills,
			"summary":     req.CandidateSummary,
			"resume_text": req.ResumeText,
		},
		"score": map[string]any{
			"match_score":       req.MatchScore,
			"missing_skills":    req.MissingSkills,
			"score_explanation": req.ScoreExplanation,
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Sprintf("could not marshal input: %v", err)
	}
	return string(b)
}
