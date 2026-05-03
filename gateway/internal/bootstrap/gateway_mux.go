package bootstrap

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/artem13815/hr/gateway/config"
	"github.com/artem13815/hr/gateway/internal/pb/admin_api"
	"github.com/artem13815/hr/gateway/internal/pb/analysis_api"
	"github.com/artem13815/hr/gateway/internal/pb/auth_api"
	"github.com/artem13815/hr/gateway/internal/pb/resume_api"
	"github.com/artem13815/hr/gateway/internal/pb/vacancy_api"
	transport_http "github.com/artem13815/hr/gateway/internal/transport/http"
)

// InitGatewayMux registers the grpc-gateway HTTP-to-gRPC handlers for
// every backend service. Each RegisterXxxHandlerFromEndpoint dials the
// backend; the connections are owned by grpc-gateway and torn down when
// ctx is cancelled (AppRun cancels it on shutdown).
func InitGatewayMux(ctx context.Context, cfg *config.Config) (*runtime.ServeMux, error) {
	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(transport_http.IncomingHeaderMatcher),
	)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := auth_api.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, cfg.Auth.GRPCAddr, opts); err != nil {
		return nil, fmt.Errorf("register auth handler: %w", err)
	}
	if err := vacancy_api.RegisterVacancyServiceHandlerFromEndpoint(ctx, mux, cfg.Vacancy.GRPCAddr, opts); err != nil {
		return nil, fmt.Errorf("register vacancy handler: %w", err)
	}
	if err := resume_api.RegisterResumeServiceHandlerFromEndpoint(ctx, mux, cfg.Resume.GRPCAddr, opts); err != nil {
		return nil, fmt.Errorf("register resume handler: %w", err)
	}
	if err := analysis_api.RegisterAnalysisServiceHandlerFromEndpoint(ctx, mux, cfg.Analysis.GRPCAddr, opts); err != nil {
		return nil, fmt.Errorf("register analysis handler: %w", err)
	}
	if err := admin_api.RegisterAdminServiceHandlerFromEndpoint(ctx, mux, cfg.Admin.GRPCAddr, opts); err != nil {
		return nil, fmt.Errorf("register admin handler: %w", err)
	}
	return mux, nil
}
