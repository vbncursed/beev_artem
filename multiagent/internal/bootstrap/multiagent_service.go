package bootstrap

import (
	"github.com/artem13815/hr/multiagent/internal/infrastructure/llm/yandex"
	"github.com/artem13815/hr/multiagent/internal/infrastructure/persistence"
	"github.com/artem13815/hr/multiagent/internal/infrastructure/prompts"
	"github.com/artem13815/hr/multiagent/internal/usecase"
)

func InitMultiAgentService(storage *persistence.MultiAgentStorage, llm *yandex.Client, promptStore *prompts.Store) *usecase.MultiAgentService {
	return usecase.NewMultiAgentService(storage, llm, promptStore)
}
