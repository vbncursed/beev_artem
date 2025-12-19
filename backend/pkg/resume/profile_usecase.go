package resume

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/artem13815/hr/pkg/llm"
)

// ProfileUseCase извлекает структурные данные из текста резюме и сохраняет их.
type ProfileUseCase interface {
	BuildAndSave(ctx context.Context, resumeID uuid.UUID, resumeText string) ProfileRecord
}

type profileService struct {
	repo      Repository
	llm       llm.ChatModel
	modelName string
	maxChars  int
}

func NewProfileService(repo Repository, model llm.ChatModel, modelName string) ProfileUseCase {
	return &profileService{
		repo:      repo,
		llm:       model,
		modelName: modelName,
		maxChars:  12000,
	}
}

func (s *profileService) BuildAndSave(ctx context.Context, resumeID uuid.UUID, resumeText string) ProfileRecord {
	rec := ProfileRecord{
		ResumeID:  resumeID,
		Status:   ProfileStatusPending,
		Model:    s.modelName,
		Error:    "",
		Profile:  Profile{Skills: []string{}, Experience: []ExperienceItem{}, Education: []EducationItem{}},
		UpdatedAt: time.Now().UTC(),
	}
	_ = s.repo.UpsertProfile(ctx, rec)

	text := strings.TrimSpace(resumeText)
	if text == "" {
		rec.Status = ProfileStatusFailed
		rec.Error = "пустой текст резюме"
		rec.UpdatedAt = time.Now().UTC()
		_ = s.repo.UpsertProfile(ctx, rec)
		return rec
	}
	if len(text) > s.maxChars {
		text = text[:s.maxChars]
	}

	if s.llm == nil {
		rec.Status = ProfileStatusFailed
		rec.Error = "LLM не настроена"
		rec.UpdatedAt = time.Now().UTC()
		_ = s.repo.UpsertProfile(ctx, rec)
		return rec
	}

	system := "Ты HR-аналитик. Отвечай на русском. Верни результат СТРОГО в JSON (без markdown/код-блоков/пояснений). Пустые массивы всегда возвращай как [], не null. Не выдумывай факты."
	user := fmt.Sprintf(
		"Текст резюме:\n<<<\n%s\n>>>\n\nВерни СТРОГО один JSON-объект по схеме:\n{\n  \"summary\": string,\n  \"skills\": string[],\n  \"experience\": [{\"company\":string,\"role\":string,\"start\":string,\"end\":string,\"description\":string}],\n  \"education\": [{\"institution\":string,\"degree\":string,\"start\":string,\"end\":string}]\n}\n\nПравила:\n- Никаких дополнительных полей\n- Никакого markdown\n- Если список пустой — []\n",
		text,
	)

	raw, err := s.llm.Ask(ctx, system, user)
	if err != nil {
		rec.Status = ProfileStatusFailed
		rec.Error = err.Error()
		rec.UpdatedAt = time.Now().UTC()
		_ = s.repo.UpsertProfile(ctx, rec)
		return rec
	}
	raw = strings.TrimSpace(raw)

	var p Profile
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		// попытка извлечь JSON из текста
		if i := strings.Index(raw, "{"); i >= 0 {
			if j := strings.LastIndex(raw, "}"); j > i {
				_ = json.Unmarshal([]byte(raw[i:j+1]), &p)
			}
		}
	}
	// нормализуем nil-слайсы
	if p.Skills == nil {
		p.Skills = []string{}
	}
	if p.Experience == nil {
		p.Experience = []ExperienceItem{}
	}
	if p.Education == nil {
		p.Education = []EducationItem{}
	}

	rec.Profile = p
	rec.Status = ProfileStatusOK
	rec.Error = ""
	rec.UpdatedAt = time.Now().UTC()
	_ = s.repo.UpsertProfile(ctx, rec)
	return rec
}


