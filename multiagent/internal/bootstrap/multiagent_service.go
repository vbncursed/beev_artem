package bootstrap

import (
	"github.com/artem13815/hr/multiagent/internal/services/multiagent_service"
	"github.com/artem13815/hr/multiagent/internal/storage/multiagent_storage"
)

func InitMultiAgentService(storage *multiagent_storage.MultiAgentStorage) *multiagent_service.MultiAgentService {
	return multiagent_service.NewMultiAgentService(storage)
}
