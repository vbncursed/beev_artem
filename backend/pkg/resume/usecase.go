package resume

import (
	"context"
	"fmt"
	"strings"

	"github.com/artem13815/hr/pkg/llm"
)

// AnalysisResult is a domain DTO with LLM recommendations.
type AnalysisResult struct {
	Model     string
	Answer    string
	CharsUsed int
	Filename  string
	Excerpted bool // true if input was truncated to fit limits
}

// AnalysisService describes the application use case for resume analysis.
type AnalysisService interface {
	Analyze(ctx context.Context, filename string, data []byte) (AnalysisResult, error)
}

type analysisService struct {
	llm            llm.ChatModel
	maxPromptChars int
}

// NewAnalysisService creates the default implementation.
func NewAnalysisService(model llm.ChatModel) AnalysisService {
	return &analysisService{
		llm:            model,
		maxPromptChars: 12_000,
	}
}

func (s *analysisService) Analyze(ctx context.Context, filename string, data []byte) (AnalysisResult, error) {
	text, err := ParseResumeText(filename, data)
	if err != nil {
		return AnalysisResult{}, err
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return AnalysisResult{}, fmt.Errorf("empty resume content")
	}
	excerpted := false
	if len(text) > s.maxPromptChars {
		text = text[:s.maxPromptChars]
		excerpted = true
	}
	system := "Ты HR-аналитик. Получишь текст резюме кандидата. Проанализируй его и дай рекомендации по улучшению. Отвечай на русском языке."
	user := fmt.Sprintf(
		"Текст резюме между маркерами:\n<<<\n%s\n>>>\nСформируй:\n1) Краткое резюме профиля (2-3 предложения)\n2) Ключевые навыки\n3) Предполагаемые роли/уровень\n4) Топ-10 рекомендаций по улучшению резюме\n5) Возможные недостающие детали (перечень)\nФорматируй ответ аккуратно, списками, без лишней воды.",
		text,
	)
	answer, err := s.llm.Ask(ctx, system, user)
	if err != nil {
		return AnalysisResult{}, err
	}
	return AnalysisResult{
		Model:     "", // filled by handler if needed from concrete impl
		Answer:    answer,
		CharsUsed: len(text),
		Filename:  filename,
		Excerpted: excerpted,
	}, nil
}
