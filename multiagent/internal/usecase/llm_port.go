package usecase

import (
	"context"
	"errors"

	"github.com/artem13815/hr/multiagent/internal/domain"
)

//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i LLM,PromptStore -o ./mocks -s _mock.go -g

// LLM is the inference port the usecase consumes. Implementations live
// under internal/infrastructure/llm (one adapter per provider). The port
// stays vendor-agnostic: usecase passes the rendered prompt and receives
// the raw completion as text — no provider-specific options leak.
//
// CompletionRequest lives in `domain` so the generated mocks subpackage
// can reference it without forming an import cycle back to `usecase`.
type LLM interface {
	Complete(ctx context.Context, in domain.CompletionRequest) (string, error)
}

// PromptStore looks up the prompt template for a role. The adapter that
// reads embed.FS implements it; the usecase doesn't care where the bytes
// come from.
type PromptStore interface {
	Get(role string) string
}

// ErrLLMUnavailable is returned when the LLM provider is unreachable or
// returns 5xx / rate-limit errors. The transport layer can map it to
// codes.Unavailable; analysis upstream treats it as a soft failure and
// keeps the heuristic AI as the authoritative answer.
var ErrLLMUnavailable = errors.New("llm unavailable")

// ErrLLMInvalidResponse is returned when the LLM returns text that doesn't
// parse as the JSON contract we asked for. Same handling as Unavailable —
// a typed sentinel for callers to disambiguate logs.
var ErrLLMInvalidResponse = errors.New("llm invalid response")
