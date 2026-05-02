package bootstrap

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/artem13815/hr/gateway/internal/pb/auth_api"
	transport_http "github.com/artem13815/hr/gateway/internal/transport/http"
)

// InitHTTPHandler composes the root handler tree:
//
//	withLogging
//	  └─ withCORS                              (preflight short-circuits here)
//	        └─ withClientIP
//	              └─ withJSONContentType
//	                    └─ rootMux:
//	                          /healthz                -> HealthHandler
//	                          /docs                   -> SwaggerHandler
//	                          /openapi/<svc>.yaml     -> SwaggerHandler (per-service)
//	                          /                       -> withAuthContext(grpc-gateway mux)
//
// SwaggerHandler is one sub-mux that owns both /docs and every
// /openapi/<slug>.yaml path. We mount it at both /docs and the /openapi/
// prefix so requests to either area dispatch into it.
//
// Order of decoration matters: logging is outermost so 401-out-at-edge
// requests still get an access log line; CORS sits second so OPTIONS
// preflights short-circuit before hitting auth/clientIP/JSON; clientIP
// runs before auth so the auth check sees the populated header in case
// it ever needs IP-scoped rate limiting.
func InitHTTPHandler(authClient auth_api.AuthServiceClient, gwMux *runtime.ServeMux, swaggerSpecs []transport_http.SwaggerSpec, allowedOrigins []string) http.Handler {
	root := http.NewServeMux()

	swaggerH := transport_http.SwaggerHandler(swaggerSpecs)
	root.Handle("/docs", swaggerH)
	root.Handle("/openapi/", swaggerH)
	root.Handle("/healthz", transport_http.HealthHandler())
	root.Handle("/", transport_http.WithAuthContext(authClient, gwMux))

	return transport_http.WithLogging(
		transport_http.WithCORS(allowedOrigins)(
			transport_http.WithClientIP(
				transport_http.WithJSONContentType(root),
			),
		),
	)
}
