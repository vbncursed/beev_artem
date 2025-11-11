package llm

import "context"

// ChatModel is a minimal abstraction for chat-based LLMs used by the domain.
// It intentionally hides concrete providers to preserve dependency direction.
type ChatModel interface {
	Ask(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}
