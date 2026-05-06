package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/artem13815/hr/multiagent/internal/domain"
)

// classifierTemperature is pinned to 0 so the same vacancy text always maps
// to the same role. Determinism matters more than creativity for a label
// task with a closed vocabulary.
const classifierTemperature = 0.0

// classifierMaxTokens caps the response. The expected payload is a one-line
// JSON object (~25 tokens); anything longer is the model rambling and the
// parser will reject it anyway.
const classifierMaxTokens = 64

// defaultRole is the catch-all returned when the LLM is unsure. Mirrors the
// PromptStore fallback for unknown roles so analysis downstream lands on
// default.txt either way.
const defaultRole = "default"

// classifierInstructionsTemplate is the system prompt. The {{ROLES}} marker
// is replaced at request time with the live role list from PromptStore so
// adding a templates/<role>.txt automatically extends the classifier
// vocabulary — no code change.
const classifierInstructionsTemplate = `Ты классификатор ролей вакансий. На вход — заголовок и описание вакансии на русском или английском языке.

Твоя задача: выбрать ровно одну роль из закрытого списка. Список допустимых ролей:
{{ROLES}}

Если ни одна роль не подходит однозначно, верни "default".

Ответ строго в формате JSON, одной строкой, без markdown-обёртки и без пояснений:
{"role": "<одно из значений списка или default>"}

Никакого текста до или после JSON. Никаких комментариев. Любое отклонение от формата — ошибка.`

// ClassifyRole maps a vacancy title+description onto one of the registered
// prompt-template roles via the LLM. Returns ErrInvalidArgument if both
// fields are empty (no signal to classify on), ErrLLMUnavailable if the
// provider call fails, ErrLLMInvalidResponse if the model returns text we
// can't parse or a role outside the allowed set.
//
// Successful responses are guaranteed to be one of:
//   - a value from PromptStore.ListRoles()
//   - the literal "default"
//
// so callers can pass the result back to GenerateDecision.role unchanged.
func (s *MultiAgentService) ClassifyRole(ctx context.Context, req domain.RoleClassifyRequest) (*domain.RoleClassifyResponse, error) {
	if strings.TrimSpace(req.Title) == "" && strings.TrimSpace(req.Description) == "" {
		return nil, ErrInvalidArgument
	}

	roles := s.prompts.ListRoles()
	instructions := buildClassifierInstructions(roles)
	input := buildClassifierInput(req)

	completion, err := s.llm.Complete(ctx, domain.CompletionRequest{
		Instructions:    instructions,
		Input:           input,
		Temperature:     classifierTemperature,
		MaxOutputTokens: classifierMaxTokens,
	})
	if err != nil {
		// Wrap with a fixed sentinel — analysis-style swallowing happens
		// upstream in the vacancy adapter, not here. The usecase keeps
		// the typed error so transport can map cleanly.
		if errors.Is(err, ErrLLMUnavailable) {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %v", ErrLLMUnavailable, err)
	}

	role, err := parseRoleClassification(completion, roles)
	if err != nil {
		return nil, err
	}
	return &domain.RoleClassifyResponse{Role: role}, nil
}

// buildClassifierInstructions renders the role list into the prompt. The
// list is bullet-formatted on separate lines so the model sees each role
// as a discrete token rather than a comma-separated blob.
func buildClassifierInstructions(roles []string) string {
	var b strings.Builder
	for _, r := range roles {
		b.WriteString("- ")
		b.WriteString(r)
		b.WriteByte('\n')
	}
	b.WriteString("- default\n")
	return strings.Replace(classifierInstructionsTemplate, "{{ROLES}}", b.String(), 1)
}

// buildClassifierInput formats the vacancy text for the user-side payload.
// Plain labels beat JSON here — there's no nested structure and the model
// reads natural language faster than a wrapping object.
func buildClassifierInput(req domain.RoleClassifyRequest) string {
	return fmt.Sprintf("Заголовок: %s\n\nОписание: %s", req.Title, req.Description)
}
