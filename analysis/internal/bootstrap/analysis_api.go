package bootstrap

import (
	"github.com/artem13815/hr/analysis/internal/api/analysis_service_api"
	"github.com/artem13815/hr/analysis/internal/services/analysis_service"
)

func InitAnalysisServiceAPI(service *analysis_service.AnalysisService) *analysis_service_api.AnalysisServiceAPI {
	return analysis_service_api.NewAnalysisServiceAPI(service)
}
