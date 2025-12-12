package analysis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/artem13815/hr/pkg/llm"
	"github.com/artem13815/hr/pkg/resume"
	"github.com/artem13815/hr/pkg/vacancy"
)

// UseCase — сценарии анализа резюме относительно вакансии.
type UseCase interface {
	Create(ctx context.Context, actorID uuid.UUID, isAdmin bool, resumeID, vacancyID uuid.UUID) (Analysis, error)
	Get(ctx context.Context, actorID uuid.UUID, isAdmin bool, id uuid.UUID) (Analysis, error)
	List(ctx context.Context, actorID uuid.UUID, isAdmin bool, limit, offset int) ([]Analysis, error)
	ListByVacancy(ctx context.Context, actorID uuid.UUID, isAdmin bool, vacancyID uuid.UUID, limit, offset int) ([]Analysis, error)
	Delete(ctx context.Context, actorID uuid.UUID, isAdmin bool, id uuid.UUID) error
}

type service struct {
	repo       Repository
	resumes    resume.Repository
	vacancies  vacancy.Repository
	llm        llm.ChatModel
	modelName  string
	maxTextLen int
}

func NewService(repo Repository, resumes resume.Repository, vacancies vacancy.Repository, model llm.ChatModel, modelName string) UseCase {
	return &service{
		repo:       repo,
		resumes:    resumes,
		vacancies:  vacancies,
		llm:        model,
		modelName:  modelName,
		maxTextLen: 12000,
	}
}

func (s *service) Create(ctx context.Context, actorID uuid.UUID, isAdmin bool, resumeID, vacancyID uuid.UUID) (Analysis, error) {
	// Access check: ensure actor owns both resume and vacancy (unless admin)
	var v vacancy.Vacancy
	var rMeta resume.Resume
	var err error
	if isAdmin {
		v, err = s.vacancies.GetByIDAny(ctx, vacancyID)
		if err != nil {
			return Analysis{}, err
		}
		rMeta, err = s.resumes.GetMetaAny(ctx, resumeID)
		if err != nil {
			return Analysis{}, err
		}
	} else {
		v, err = s.vacancies.GetByIDForOwner(ctx, actorID, vacancyID)
		if err != nil {
			return Analysis{}, err
		}
		rMeta, err = s.resumes.GetMetaForOwner(ctx, actorID, resumeID)
		if err != nil {
			return Analysis{}, err
		}
	}
	_ = rMeta // currently used only for checks; could be included later

	parsed, err := s.resumes.GetParsed(ctx, resumeID)
	if err != nil {
		return Analysis{}, err
	}
	text := strings.TrimSpace(parsed.Text)
	if text == "" {
		return Analysis{}, errors.New("пустой текст резюме")
	}
	if len(text) > s.maxTextLen {
		text = text[:s.maxTextLen]
	}

	matched, missing, score := computeSkillMatch(v.Skills, text)
	if matched == nil {
		matched = []string{}
	}
	if missing == nil {
		missing = []string{}
	}

	rep := Report{
		MatchedSkills: matched,
		MissingSkills: missing,
	}

	// Try to enrich with LLM; on failure keep deterministic report.
	modelUsed := ""
	if s.llm != nil {
		llmReport, err := s.askLLM(ctx, v, text, matched, missing)
		if err == nil {
			rep.CandidateSummary = llmReport.CandidateSummary
			rep.UniqueStrengths = llmReport.UniqueStrengths
			rep.AIRecommendationForHR = llmReport.AIRecommendationForHR
			rep.AIRecommendationForCandidate = llmReport.AIRecommendationForCandidate
			modelUsed = s.modelName
		} else {
			// degrade gracefully: store error text as HR note
			rep.AIRecommendationForHR = fmt.Sprintf("LLM временно недоступен: %v", err)
		}
	}

	a := Analysis{
		ID:        uuid.New(),
		ResumeID:  resumeID,
		VacancyID: vacancyID,
		Score:     score,
		Model:     modelUsed,
		Report:    rep,
		CreatedAt: time.Now().UTC(),
	}
	return s.repo.Create(ctx, a)
}

func (s *service) Get(ctx context.Context, actorID uuid.UUID, isAdmin bool, id uuid.UUID) (Analysis, error) {
	if isAdmin {
		return s.repo.GetByID(ctx, id)
	}
	return s.repo.GetByIDForOwner(ctx, actorID, id)
}

func (s *service) List(ctx context.Context, actorID uuid.UUID, isAdmin bool, limit, offset int) ([]Analysis, error) {
	if isAdmin {
		return s.repo.ListAll(ctx, limit, offset)
	}
	return s.repo.ListByOwner(ctx, actorID, limit, offset)
}

func (s *service) ListByVacancy(ctx context.Context, actorID uuid.UUID, isAdmin bool, vacancyID uuid.UUID, limit, offset int) ([]Analysis, error) {
	if isAdmin {
		return s.repo.ListByVacancyAny(ctx, vacancyID, limit, offset)
	}
	return s.repo.ListByVacancyForOwner(ctx, actorID, vacancyID, limit, offset)
}

func (s *service) Delete(ctx context.Context, actorID uuid.UUID, isAdmin bool, id uuid.UUID) error {
	if isAdmin {
		return s.repo.DeleteAny(ctx, id)
	}
	return s.repo.DeleteForOwner(ctx, actorID, id)
}

type llmPayload struct {
	CandidateSummary             string   `json:"candidateSummary"`
	UniqueStrengths              []string `json:"uniqueStrengths"`
	AIRecommendationForHR        string   `json:"aiRecommendationForHR"`
	AIRecommendationForCandidate []string `json:"aiRecommendationForCandidate"`
}

func (s *service) askLLM(ctx context.Context, v vacancy.Vacancy, resumeText string, matched, missing []string) (llmPayload, error) {
	system := "Ты HR-аналитик. Верни результат строго в JSON без пояснений."
	user := fmt.Sprintf(
		"Вакансия:\nНазвание: %s\nОписание: %s\n\nСовпавшие навыки: %s\nОтсутствующие навыки: %s\n\nТекст резюме:\n<<<\n%s\n>>>\n\nВерни JSON с полями:\n- candidateSummary (string)\n- uniqueStrengths (string[])\n- aiRecommendationForHR (string)\n- aiRecommendationForCandidate (string[])\n",
		v.Title,
		v.Description,
		strings.Join(matched, ", "),
		strings.Join(missing, ", "),
		resumeText,
	)
	raw, err := s.llm.Ask(ctx, system, user)
	if err != nil {
		return llmPayload{}, err
	}
	raw = strings.TrimSpace(raw)
	// best-effort JSON parse
	var out llmPayload
	if err := json.Unmarshal([]byte(raw), &out); err == nil {
		return out, nil
	}
	// try to extract JSON from fenced block
	if i := strings.Index(raw, "{"); i >= 0 {
		if j := strings.LastIndex(raw, "}"); j > i {
			_ = json.Unmarshal([]byte(raw[i:j+1]), &out)
			if out.AIRecommendationForHR != "" || out.CandidateSummary != "" {
				return out, nil
			}
		}
	}
	return llmPayload{}, fmt.Errorf("не удалось распарсить JSON ответ LLM")
}

func computeSkillMatch(skills []vacancy.SkillWeight, resumeText string) (matched []string, missing []string, score float32) {
	txt := strings.ToLower(resumeText)
	var total float32
	var got float32
	for _, s := range skills {
		name := strings.ToLower(strings.TrimSpace(s.Skill))
		if name == "" {
			continue
		}
		w := s.Weight
		if w < 0 {
			w = 0
		}
		total += w
		if strings.Contains(txt, name) {
			matched = append(matched, s.Skill)
			got += w
		} else {
			missing = append(missing, s.Skill)
		}
	}
	if total <= 0 {
		return matched, missing, 0
	}
	return matched, missing, got / total
}
