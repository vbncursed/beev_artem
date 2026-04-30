package bootstrap

import (
	"github.com/artem13815/hr/multiagent/internal/usecase"
	"github.com/artem13815/hr/multiagent/internal/infrastructure/persistence"
)

func InitMultiAgentService(storage *persistence.MultiAgentStorage) *usecase.MultiAgentService {
	return usecase.NewMultiAgentService(storage)
}
