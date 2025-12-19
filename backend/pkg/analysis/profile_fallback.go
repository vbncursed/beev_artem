package analysis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/artem13815/hr/pkg/resume"
)

// buildProfileFromParsed builds and stores resume profile based on parsed resume text.
// It is used as a fallback for flows where profile wasn't built at upload time.
func (s *service) buildProfileFromParsed(ctx context.Context, resumeID uuid.UUID) (resume.ProfileRecord, error) {
	parsed, err := s.resumes.GetParsed(ctx, resumeID)
	if err != nil {
		return resume.ProfileRecord{}, fmt.Errorf("не найден распарсенный текст резюме: %w", err)
	}
	return s.buildAndSaveProfile(ctx, resumeID, parsed.Text), nil
}

func (s *service) buildAndSaveProfile(ctx context.Context, resumeID uuid.UUID, resumeText string) resume.ProfileRecord {
	const maxChars = 12000
	rec := resume.ProfileRecord{
		ResumeID:  resumeID,
		Status:    resume.ProfileStatusPending,
		Model:     s.modelName,
		Error:     "",
		Profile:   resume.Profile{Skills: []string{}, Experience: []resume.ExperienceItem{}, Education: []resume.EducationItem{}},
		UpdatedAt: time.Now().UTC(),
	}
	_ = s.resumes.UpsertProfile(ctx, rec)

	text := strings.TrimSpace(resumeText)
	if text == "" {
		rec.Status = resume.ProfileStatusFailed
		rec.Error = "пустой текст резюме"
		rec.UpdatedAt = time.Now().UTC()
		_ = s.resumes.UpsertProfile(ctx, rec)
		return rec
	}
	if len(text) > maxChars {
		text = text[:maxChars]
	}

	if s.llm == nil {
		rec.Status = resume.ProfileStatusFailed
		rec.Error = "LLM не настроена"
		rec.UpdatedAt = time.Now().UTC()
		_ = s.resumes.UpsertProfile(ctx, rec)
		return rec
	}

	system := "Ты HR-аналитик. Отвечай на русском. Верни результат СТРОГО в JSON (без markdown/код-блоков/пояснений). Пустые массивы всегда возвращай как [], не null. Не выдумывай факты."
	user := fmt.Sprintf(
		"Текст резюме:\n<<<\n%s\n>>>\n\nВерни СТРОГО один JSON-объект по схеме:\n{\n  \"summary\": string,\n  \"skills\": string[],\n  \"experience\": [{\"company\":string,\"role\":string,\"start\":string,\"end\":string,\"description\":string}],\n  \"education\": [{\"institution\":string,\"degree\":string,\"start\":string,\"end\":string}]\n}\n\nПравила:\n- Никаких дополнительных полей\n- Никакого markdown\n- Если список пустой — []\n",
		text,
	)

	raw, err := s.llm.Ask(ctx, system, user)
	if err != nil {
		rec.Status = resume.ProfileStatusFailed
		rec.Error = err.Error()
		rec.UpdatedAt = time.Now().UTC()
		_ = s.resumes.UpsertProfile(ctx, rec)
		return rec
	}
	raw = strings.TrimSpace(raw)

	var p resume.Profile
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		// попытка извлечь JSON из текста
		if i := strings.Index(raw, "{"); i >= 0 {
			if j := strings.LastIndex(raw, "}"); j > i {
				_ = json.Unmarshal([]byte(raw[i:j+1]), &p)
			}
		}
	}
	// Если после парсинга профиль пустой (частый случай при мусорном ответе),
	// помечаем как failed, чтобы клиент видел причину.
	if p.Skills == nil {
		p.Skills = []string{}
	}
	if p.Experience == nil {
		p.Experience = []resume.ExperienceItem{}
	}
	if p.Education == nil {
		p.Education = []resume.EducationItem{}
	}
	if strings.TrimSpace(p.Summary) == "" && len(p.Skills) == 0 && len(p.Experience) == 0 && len(p.Education) == 0 {
		rec.Status = resume.ProfileStatusFailed
		rec.Error = "не удалось извлечь профиль из ответа LLM"
		rec.Profile = resume.Profile{Skills: []string{}, Experience: []resume.ExperienceItem{}, Education: []resume.EducationItem{}}
		rec.UpdatedAt = time.Now().UTC()
		_ = s.resumes.UpsertProfile(ctx, rec)
		return rec
	}

	rec.Profile = p
	rec.Status = resume.ProfileStatusOK
	rec.Error = ""
	rec.UpdatedAt = time.Now().UTC()
	_ = s.resumes.UpsertProfile(ctx, rec)
	return rec
}

// compile-time guard: make sure the repository has required methods (we rely on these in fallback).
var _ interface {
	GetParsed(context.Context, uuid.UUID) (resume.Parsed, error)
	UpsertProfile(context.Context, resume.ProfileRecord) error
} = (resume.Repository)(nil)

// silence unused import in case of future refactors
var _ = errors.New
