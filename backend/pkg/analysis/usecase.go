package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/artem13815/hr/pkg/llm"
	"github.com/artem13815/hr/pkg/nlp"
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
	repo      Repository
	resumes   resume.Repository
	vacancies vacancy.Repository
	llm       llm.ChatModel
	modelName string
}

func NewService(repo Repository, resumes resume.Repository, vacancies vacancy.Repository, model llm.ChatModel, modelName string) UseCase {
	return &service{
		repo:      repo,
		resumes:   resumes,
		vacancies: vacancies,
		llm:       model,
		modelName: modelName,
	}
}

type ErrProfileNotReady struct {
	Status resume.ProfileStatus
}

func (e ErrProfileNotReady) Error() string {
	if e.Status == "" {
		return "профиль резюме не готов"
	}
	return fmt.Sprintf("профиль резюме не готов (status=%s)", e.Status)
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

	// Prefer structured profile (built at upload time)
	var profRec resume.ProfileRecord
	if isAdmin {
		profRec, err = s.resumes.GetProfileAny(ctx, resumeID)
	} else {
		profRec, err = s.resumes.GetProfileForOwner(ctx, actorID, resumeID)
	}
	if err != nil {
		// Профиль мог не быть построен (например, резюме загружено через /resume/analyze).
		// Пытаемся построить профиль на лету из parsed text и сохранить.
		built, buildErr := s.buildProfileFromParsed(ctx, resumeID)
		if buildErr != nil {
			// Не удалось построить профиль автоматически — возвращаем причину.
			return Analysis{}, buildErr
		}
		profRec = built
	}
	if profRec.Status != resume.ProfileStatusOK {
		return Analysis{}, ErrProfileNotReady{Status: profRec.Status}
	}
	// normalize nil slices
	if profRec.Profile.Skills == nil {
		profRec.Profile.Skills = []string{}
	}

	matched, missing, score := computeSkillMatchFromProfile(v.Skills, profRec.Profile)
	if matched == nil {
		matched = []string{}
	}
	if missing == nil {
		missing = []string{}
	}

	rep := Report{
		CandidateSummary: strings.TrimSpace(profRec.Profile.Summary),
		MatchedSkills:    matched,
		MissingSkills:    missing,
	}

	// Try to enrich with LLM; on failure keep deterministic report.
	modelUsed := ""
	if s.llm != nil {
		llmReport, err := s.askLLM(ctx, v, profRec.Profile, matched, missing)
		if err == nil {
			if llmReport.CandidateSummary != "" {
				rep.CandidateSummary = llmReport.CandidateSummary
			}
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

func (s *service) askLLM(ctx context.Context, v vacancy.Vacancy, prof resume.Profile, matched, missing []string) (llmPayload, error) {
	system := "Ты HR-аналитик. Отвечай на русском. Верни результат СТРОГО в JSON (без markdown/код-блоков/пояснений/лишнего текста). Пустые массивы всегда возвращай как [], не null. Не выдумывай факты: если данных нет — пиши нейтрально."
	if prof.Skills == nil {
		prof.Skills = []string{}
	}
	if prof.Experience == nil {
		prof.Experience = []resume.ExperienceItem{}
	}
	if prof.Education == nil {
		prof.Education = []resume.EducationItem{}
	}
	user := fmt.Sprintf(
		"Дано:\nВакансия:\n- Название: %s\n- Описание: %s\n\nСтруктурный профиль резюме:\n- Summary: %s\n- Skills: %s\n- Experience count: %d\n- Education count: %d\n\nСовпавшие навыки: %s\nОтсутствующие навыки: %s\n\nВерни СТРОГО один JSON-объект со схемой:\n{\n  \"candidateSummary\": string,\n  \"uniqueStrengths\": string[],\n  \"aiRecommendationForHR\": string,\n  \"aiRecommendationForCandidate\": string[]\n}\n\nПравила:\n- Никаких дополнительных полей\n- Никакого markdown\n- Если список пустой — []\n",
		v.Title,
		v.Description,
		strings.TrimSpace(prof.Summary),
		strings.Join(prof.Skills, ", "),
		len(prof.Experience),
		len(prof.Education),
		strings.Join(matched, ", "),
		strings.Join(missing, ", "),
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

func computeSkillMatchFromProfile(skills []vacancy.SkillWeight, prof resume.Profile) (matched []string, missing []string, score float32) {
	// 1) Candidate skills index (normalized variants set)
	candSkillSet := make(map[string]struct{}, len(prof.Skills)*2)
	for _, s := range prof.Skills {
		for _, v := range nlp.SkillVariants(s) {
			candSkillSet[v] = struct{}{}
		}
	}

	// 2) Candidate "corpus" text for fallback matching (summary + experience + education)
	var corpusParts []string
	if strings.TrimSpace(prof.Summary) != "" {
		corpusParts = append(corpusParts, prof.Summary)
	}
	for _, e := range prof.Experience {
		if strings.TrimSpace(e.Company) != "" {
			corpusParts = append(corpusParts, e.Company)
		}
		if strings.TrimSpace(e.Role) != "" {
			corpusParts = append(corpusParts, e.Role)
		}
		if strings.TrimSpace(e.Description) != "" {
			corpusParts = append(corpusParts, e.Description)
		}
	}
	for _, e := range prof.Education {
		if strings.TrimSpace(e.Institution) != "" {
			corpusParts = append(corpusParts, e.Institution)
		}
		if strings.TrimSpace(e.Degree) != "" {
			corpusParts = append(corpusParts, e.Degree)
		}
	}
	corpus := nlp.NormalizeText(strings.Join(corpusParts, " "))

	var total float32
	var got float32
	for _, s := range skills {
		orig := strings.TrimSpace(s.Skill)
		variants := nlp.SkillVariants(orig)
		if len(variants) == 0 {
			continue
		}
		w := s.Weight
		if w < 0 {
			w = 0
		}
		total += w

		matchScore := float32(0)

		// A) Exact/synonym match in skills list => 1.0
		for _, v := range variants {
			if _, ok := candSkillSet[v]; ok {
				matchScore = 1.0
				break
			}
		}

		// B) Phrase mentioned in corpus => 0.8
		if matchScore < 1.0 && corpus != "" {
			for _, v := range variants {
				if nlp.ContainsPhrase(corpus, v) {
					matchScore = 0.8
					break
				}
			}
		}

		// C) Partial token overlap for multi-word skills => 0.6
		if matchScore < 0.8 {
			// build token set of candidate skills
			candTokens := map[string]struct{}{}
			for s := range candSkillSet {
				for _, t := range nlp.TokensList(s) {
					if t != "" {
						candTokens[t] = struct{}{}
					}
				}
			}
			best := float32(0)
			for _, v := range variants {
				toks := nlp.TokensList(v)
				if len(toks) < 2 {
					continue
				}
				hit := 0
				for _, t := range toks {
					if _, ok := candTokens[t]; ok {
						hit++
					}
				}
				ratio := float32(hit) / float32(len(toks))
				if ratio > best {
					best = ratio
				}
			}
			if best >= 0.6 {
				matchScore = 0.6
			}
		}

		got += w * matchScore

		// For report: treat >=0.6 as matched, else missing
		if matchScore >= 0.6 {
			matched = append(matched, orig)
		} else {
			missing = append(missing, orig)
		}
	}
	if total <= 0 {
		return matched, missing, 0
	}
	return matched, missing, got / total
}
