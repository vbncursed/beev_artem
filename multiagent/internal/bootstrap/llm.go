package bootstrap

import (
	"github.com/artem13815/hr/multiagent/config"
	"github.com/artem13815/hr/multiagent/internal/infrastructure/llm/yandex"
	"github.com/artem13815/hr/multiagent/internal/infrastructure/prompts"
)

// InitYandexLLM builds the Yandex Cloud Foundation Models adapter from
// config. cfg.Yandex.APIKey is already loaded from env by LoadConfig.
func InitYandexLLM(cfg *config.Config) *yandex.Client {
	return yandex.New(yandex.Config{
		FolderID:        cfg.Yandex.FolderID,
		APIKey:          cfg.Yandex.APIKey,
		Model:           cfg.Yandex.Model,
		RequestTimeout:  cfg.Yandex.RequestTimeout,
		MaxOutputTokens: cfg.Yandex.MaxOutputTokens,
		RateLimitRPS:    cfg.RateLimit.RPS,
		RateLimitBurst:  cfg.RateLimit.Burst,
	})
}

// InitPromptStore loads the embedded role prompts. Returns an error
// instead of panicking so main can include it in the structured-log
// fatal line and so tests can swap in a fake.
func InitPromptStore() (*prompts.Store, error) {
	return prompts.New()
}
